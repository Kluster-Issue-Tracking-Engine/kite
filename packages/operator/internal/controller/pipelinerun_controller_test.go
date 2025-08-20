/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	knative "knative.dev/pkg/apis"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	KiteBridgeOperatorNamespace = "kite-bridge-operator"
)

func setupPipelineRun(name string, options PipelineRunBuilderOptions) {
	builder := NewPipelineRunBuilder(name, KiteBridgeOperatorNamespace)
	var pipelineRun *v1.PipelineRun
	if options.Labels != nil {
		builder.WithLabels(options.Labels)
	}

	pipelineRun = builder.Build()
	Expect(k8sClient.Create(ctx, pipelineRun)).Should(Succeed())
	current := &v1.PipelineRun{}
	key := types.NamespacedName{Name: name, Namespace: KiteBridgeOperatorNamespace}

	Eventually(func(g Gomega) {
		g.Expect(k8sClient.Get(ctx, key, current)).To(Succeed())
	}).Should(Succeed())

	// Things like Status need to be updated after
	if options.Conditions != nil {
		current.Status.Conditions = options.Conditions
	}
	if options.CompletionTime != nil {
		current.Status.CompletionTime = options.CompletionTime
	}

	Eventually(func(g Gomega) {
		g.Expect(k8sClient.Status().Update(ctx, current)).To(Succeed())
	}).Should(Succeed())
}

func tearDownPipelineRuns() {
	pipelineRuns := listPipelineRuns(KiteBridgeOperatorNamespace)
	for _, pipelineRuns := range pipelineRuns {
		Expect(k8sClient.Delete(ctx, &pipelineRuns)).Should(Succeed())
	}
	Eventually(func() []v1.PipelineRun {
		return listPipelineRuns(KiteBridgeOperatorNamespace)
	}).Should(BeEmpty())
}

var _ = Describe("PipelineRun Controller", func() {
	var (
		reconciler     *PipelineRunReconciler
		mockKiteClient *MockKiteClient
		logBuffer      bytes.Buffer
		logger         *logrus.Logger
	)

	BeforeEach(func() {
		createNamespace(KiteBridgeOperatorNamespace)
		mockKiteClient = &MockKiteClient{}
		logger = logrus.New()
		logger.SetOutput(&logBuffer)

		reconciler = &PipelineRunReconciler{
			Client:     k8sClient,
			Scheme:     k8sClient.Scheme(),
			KiteClient: mockKiteClient,
			Logger:     logger,
		}
	})

	AfterEach(func() {
		logBuffer.Reset()
		tearDownPipelineRuns()
	})

	Context("When a PipelineRun fails", func() {
		var (
			prName    = "failed-pipeline-xyz"
			lookupKey = types.NamespacedName{Name: prName, Namespace: KiteBridgeOperatorNamespace}
			pr        = &v1.PipelineRun{}
		)

		BeforeEach(func() {
			now := metav1.Now()
			setupPipelineRun(prName, PipelineRunBuilderOptions{
				Conditions: []knative.Condition{
					{
						Type:    "Succeeded",
						Message: "Tasks Completed: 1 (Failed: 0, Cancelled: 0), Skipped: 0",
						Status:  "False",
						Reason:  "Failed",
					},
				},
				Labels: map[string]string{
					"tekton.dev/pipeline": "failed-pipeline",
				},
				CompletionTime: &now,
			})

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, lookupKey, pr)).To(Succeed())
			}).Should(Succeed())
		})

		It("should successfully reconcile the PipelineRun when it fails", func() {
			// Trigger reconciliation
			result, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: lookupKey,
			})

			// Verify reconciliation succeeded
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			// Verify Kite client was called with failure
			Expect(mockKiteClient.FailureReports).To(HaveLen(1))
			failureReport := mockKiteClient.FailureReports[0]

			// Verify PR has what we expect
			Expect(failureReport.PipelineName).To(Equal("failed-pipeline"))
			Expect(failureReport.Namespace).To(Equal(KiteBridgeOperatorNamespace))
			Expect(failureReport.FailureReason).To(ContainSubstring("Tasks Completed"))
			Expect(failureReport.RunID).To(Equal(string(pr.UID)))
			Expect(failureReport.Severity).To(Equal("major"))
		})

		It("should retry when Kite client fails", func() {
			// Lets set it up to fail
			mockKiteClient.ShouldFail = true
			result, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: lookupKey,
			})

			Expect(err).To(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(RetryWaitPeriod))
			// We should still have some record of attempting to call KITE
			Expect(mockKiteClient.FailureReports).To(HaveLen(1))
		})
	})

	Context("When a PipelineRun succeeds", func() {
		var (
			prName    = "successful-pipeline-xyz"
			lookupKey = types.NamespacedName{Name: prName, Namespace: KiteBridgeOperatorNamespace}
			pr        = &v1.PipelineRun{}
		)

		BeforeEach(func() {
			now := metav1.Now()
			setupPipelineRun(prName, PipelineRunBuilderOptions{
				Conditions: []knative.Condition{
					{
						Type:    "Succeeded",
						Message: "1 (Failed: 0, Cancelled 0), Skipped: 0",
						Status:  "True",
						Reason:  "Succeeded",
					},
				},
				Labels: map[string]string{
					"tekton.dev/pipeline": "successful-pipeline",
				},
				CompletionTime: &now,
			})

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, lookupKey, pr)).To(Succeed())
			}).Should(Succeed())
		})

		It("Should successfully reconcile the PipelineRun if it succeeds", func() {
			// Trigger reconciliation
			result, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: lookupKey,
			})

			// Verify reconciliation
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			// Verify KITE client was called for success
			Expect(mockKiteClient.SuccessReports).To(HaveLen(1))
			successfulPayload := mockKiteClient.SuccessReports[0]

			// Verify PR is what we expect
			Expect(successfulPayload.PipelineName).To(Equal("successful-pipeline"))
			Expect(successfulPayload.Namespace).To(Equal(KiteBridgeOperatorNamespace))
			Expect(successfulPayload).NotTo(HaveExistingField("FailureReason"))
		})
	})

	Context("When a pipeline is not completed", func() {
		var (
			prName    = "pending-pipeline-xyz"
			lookupKey = types.NamespacedName{Name: prName, Namespace: KiteBridgeOperatorNamespace}
			pr        = &v1.PipelineRun{}
		)

		BeforeEach(func() {
			setupPipelineRun(prName, PipelineRunBuilderOptions{
				Conditions: []knative.Condition{
					{
						Type:    "Unknown",
						Message: "running...",
						Status:  "Unknown",
						Reason:  "Running",
					},
				},
				Labels: map[string]string{
					"tekton.dev/pipeline": "pending-pipeline",
				},
				CompletionTime: nil,
			})

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, lookupKey, pr)).ToNot(Succeed())
			}).ShouldNot(Succeed())
		})

		It("Should ignore pipeline runs that are not done running", func() {
			result, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: lookupKey,
			})

			// An error shouldn't trigger for this
			Expect(err).ToNot(HaveOccurred())
			// We should get back an empty result
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})

	Context("When determining severity", func() {
		var (
			prName    = "some-pipeline-prod-xyz"
			lookupKey = types.NamespacedName{Name: prName, Namespace: KiteBridgeOperatorNamespace}
			pr        = &v1.PipelineRun{}
		)

		BeforeEach(func() {
			now := metav1.Now()
			setupPipelineRun(prName, PipelineRunBuilderOptions{
				Conditions: []knative.Condition{
					{
						Type:    "Succeeded",
						Message: "1 (Failed: 0, Cancelled 0), Skipped: 0",
						Status:  "True",
						Reason:  "Succeeded",
					},
				},
				Labels: map[string]string{
					"tekton.dev/pipeline": "some-pipeline-rpod",
				},
				CompletionTime: &now,
			})

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, lookupKey, pr)).To(Succeed())
			}).Should(Succeed())
		})

		It("Should give correct priority based on defined logic ", func() {
			// Major
			priority := reconciler.determineSeverity(pr)
			Expect(priority).To(Equal("major"))

			// Release
			// Remove 'prod' from name
			pr.Name = "some-pipeline-xyz"
			// Make this a release PR
			pr.Labels = make(map[string]string)
			pr.Labels["appstudio.openshift.io/service"] = "release"
			priority = reconciler.determineSeverity(pr)
			Expect(priority).To(Equal("critical"))

			// Build
			pr.Labels = make(map[string]string)
			pr.Labels["pipelines.appstudio.openshift.io/type"] = "build"
			priority = reconciler.determineSeverity(pr)
			Expect(priority).To(Equal("medium"))
			// Test
			pr.Labels["pipelines.appstudio.openshift.io/type"] = "test"
			priority = reconciler.determineSeverity(pr)
			Expect(priority).To(Equal("medium"))

			// Default
			pr.Labels = nil
			priority = reconciler.determineSeverity(pr)
			Expect(priority).To(Equal("major"))
		})
	})

	Context("When extracting pipeline name from pipeline run", func() {
		var reconciler *PipelineRunReconciler

		BeforeEach(func() {
			reconciler = &PipelineRunReconciler{Logger: logrus.New()}
		})

		It("should use PipelineRef if available", func() {
			pr := &v1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{Name: "my-pipeline-run"},
				Spec: v1.PipelineRunSpec{
					PipelineRef: &v1.PipelineRef{
						Name: "my-pipeline",
					},
				},
			}
			pipelineName := reconciler.getPipelineName(pr)
			Expect(pipelineName).To(Equal("my-pipeline"))
		})

		It("should use tekton.dev/pipeline label when PipelineRef is not set", func() {
			pr := &v1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{
					Name: "basic-pipeline-run",
					Labels: map[string]string{
						"tekton.dev/pipeline": "basic-pipeline",
					},
				},
			}
			pipelineName := reconciler.getPipelineName(pr)
			Expect(pipelineName).To(Equal("basic-pipeline"))
		})

		It("should fallback to PipelineRun name", func() {
			pr := &v1.PipelineRun{
				ObjectMeta: metav1.ObjectMeta{
					Name: "some-pipeline-run",
				},
			}
			pipelineRunName := reconciler.getPipelineName(pr)
			Expect(pipelineRunName).To(Equal("some-pipeline-run"))
		})
	})

	Context("When a PipelineRun doesn't exist", func() {
		It("should handle not found gracefully", func() {
			result, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "not-found",
					Namespace: KiteBridgeOperatorNamespace,
				},
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})
})
