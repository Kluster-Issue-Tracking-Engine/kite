# Controller Development Guide
This guide shows you how to add custom controllers to the **Kite Bridge Operator** using [operator-sdk](https://sdk.operatorframework.io/docs/building-operators/golang/).

The operator is designed to be **extensible**, so you can monitor any Kubernetes resource and integrate with Kite.

---

## Table of Contents
1. [Quick Start](#quick-start)
2. [Architecture Overview](#architecture-overview)
3. [Creating a New Controller](#creating-a-new-controller)
   - [Step 1: Scaffold your controller](#step-1-scaffold-your-controller)
   - [Step 2: Implement business logic](#step-2-implement-business-logic)
   - [Step 3: Extend the Kite Client](#step-3-extend-the-kite-client)
   - [Step 4: Register your controller](#step-4-register-your-controller)
   - [Step 5: Add imports & scheme registration](#step-5-add-imports--scheme-registration)
   - [Step 6: Write tests](#step-6-write-tests)
4. [Working with Custom Resources](#working-with-custom-resources)
5. [Configuration](#configuration)
6. [RBAC Permissions](#rbac-permissions)
7. [Recommended Best Practices](#recommended-best-practices)

---

## Quick Start
The operator already includes a working [`PipelineRun` controller](../internal/controller/pipelinerun_controller.go). Use it as a reference for creating controllers for your own resources.

---

## Architecture Overview
The operator follows the standard [controller-runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime) pattern:
```bash
packages/operator/
├── cmd/main.go                          # Operator entry point
├── internal/
│   ├── controller/                      # Controllers directory
│   │   ├── pipelinerun_controller.go    # Example controller
│   │   └── your_resource_controller.go  # Your new controller
│   └── clients/
│       └── kite.go                      # Kite API client interface
```

---

## Creating a New Controller
### Step 1. Scaffold your controller
```bash
cd packages/operator

# For built-in Kubernetes resources (no CRD needed):
operator-sdk create api --group apps --version v1 --kind Deployment --controller --resource=false

# For Custom Resources (CRD must already exist in cluster):
operator-sdk create api --group <your-group> --version <version> --kind <YourKind> --controller --resource=false

# Example for monitoring custom MyApp resources:
operator-sdk create api --group myapp.example.com --version v1 --kind MyApp --controller --resource=false
```

**Why `--resource=false`?**:
- The `--resource=false` flag means operator-sdk won't generate CRD manifests.
- Your CRDs must already be installed in the target cluster, the operator just needs to watch them.
- This operator does not own the CRDs, it only watches them.

---

### Step 2: Implement business logic
Use the [`PipelineRun` controller](../internal/controller/pipelinerun_controller.go) as a starting point.
Here is an example that watches Deployments.
```go
package controller

import (
    "context"
    "time"

    clients "github.com/konflux-ci/kite/packages/operator/internal/clients"
    "github.com/sirupsen/logrus"
    appsv1 "k8s.io/api/apps/v1"  // Your resource type
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

// YourResourceReconciler reconciles your resource
type YourResourceReconciler struct {
    client.Client
    Scheme     *runtime.Scheme
    KiteClient clients.KiteWebhookClient
    Logger     *logrus.Logger
}

// Define your constants
const (
    RetryWaitPeriod = time.Minute * 2
)

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get

func (r *YourResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Fetch your resource
    var resource appsv1.Deployment
    if err := r.Get(ctx, req.NamespacedName, &resource); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // 2. Check if resource meets your conditions for creating/resolving issues
    if r.shouldCreateIssue(&resource) {
        return r.handleResourceFailure(ctx, &resource)
    } else if r.shouldResolveIssue(&resource) {
        return r.handleResourceSuccess(ctx, &resource)
    }

    return ctrl.Result{}, nil
}

// Implement your business logic
func (r *YourResourceReconciler) shouldCreateIssue(resource *appsv1.Deployment) bool {
    // Your logic here - check resource conditions, status, etc.
    return false
}

func (r *YourResourceReconciler) shouldResolveIssue(resource *appsv1.Deployment) bool {
    // Your logic here
    return false
}

func (r *YourResourceReconciler) handleResourceFailure(ctx context.Context, resource *appsv1.Deployment) (ctrl.Result, error) {
    // Create payload for Kite
    payload := clients.YourFailurePayload{
        ResourceName: resource.Name,
        Namespace:    resource.Namespace,
        Reason:       "Your failure reason",
        Severity:     "major",
    }

    if err := r.KiteClient.ReportYourResourceFailure(ctx, payload); err != nil {
        r.Logger.WithError(err).Error("Failed to report resource failure")
        return ctrl.Result{RequeueAfter: RetryWaitPeriod}, err
    }

    return ctrl.Result{}, nil
}

func (r *YourResourceReconciler) handleResourceSuccess(ctx context.Context, resource *appsv1.Deployment) (ctrl.Result, error) {
    // Similar to failure, but for success
    payload := clients.YourSuccessPayload{
        ResourceName: resource.Name,
        Namespace:    resource.Namespace,
    }

    if err := r.KiteClient.ReportYourResourceSuccess(ctx, payload); err != nil {
        r.Logger.WithError(err).Error("Failed to report resource success")
        return ctrl.Result{RequeueAfter: RetryWaitPeriod}, err
    }

    return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *YourResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&appsv1.Deployment{}). // Your resource type
        Named("deployment").        // Controller name
        Complete(r)
}
```

---

### Step 3: Extend the Kite Client
Add your new webhook methods in `internal/clients/kite.go`:
```go
// Add methods to KiteWebhookClient interface
type KiteWebhookClient interface {
    ReportPipelineFailure(ctx context.Context, payload PipelineFailurePayload) error
    ReportPipelineSuccess(ctx context.Context, payload PipelineSuccessPayload) error
    // Add your methods:
    ReportYourResourceFailure(ctx context.Context, payload YourFailurePayload) error
    ReportYourResourceSuccess(ctx context.Context, payload YourSuccessPayload) error
}

// Add your payload structs
type YourFailurePayload struct {
    ResourceName string `json:"resourceName"`
    Namespace    string `json:"namespace"`
    Reason       string `json:"reason"`
    Severity     string `json:"severity,omitempty"`
    // Add fields specific to your resource
}

type YourSuccessPayload struct {
    ResourceName string `json:"resourceName"`
    Namespace    string `json:"namespace"`
}

// Implement the methods in KiteClient
func (k *KiteClient) ReportYourResourceFailure(ctx context.Context, payload YourFailurePayload) error {
    url := fmt.Sprintf("%s/api/v1/webhooks/your-resource-failure?namespace=%s", k.baseURL, payload.Namespace)
    return k.sendWebhook(ctx, url, payload, "your-resource-failure")
}

func (k *KiteClient) ReportYourResourceSuccess(ctx context.Context, payload YourSuccessPayload) error {
    url := fmt.Sprintf("%s/api/v1/webhooks/your-resource-success?namespace=%s", k.baseURL, payload.Namespace)
    return k.sendWebhook(ctx, url, payload, "your-resource-success")
}
```

Don’t forget to:
- Add new payload structs
- Implement the methods
- Regenerate mocks if you use them in tests

---

### Step 4: Register your Controller:
Update `cmd/main.go` to set up your reconciler with the manager:
```go
func main() {
    // ... existing setup code ...

    // Initialize Kite client
    logger := logrus.New()
    kiteClient := clients.NewKiteClient(kiteURL, logger)

    // Setup existing controllers
    if err = (&controller.PipelineRunReconciler{
        Client:     mgr.GetClient(),
        Scheme:     mgr.GetScheme(),
        KiteClient: kiteClient,
        Logger:     logger,
    }).SetupWithManager(mgr); err != nil {
        setupLog.Error(err, "unable to create controller", "controller", "PipelineRun")
        os.Exit(1)
    }

    // Setup your new controller
    if err = (&controller.YourResourceReconciler{
        Client:     mgr.GetClient(),
        Scheme:     mgr.GetScheme(),
        KiteClient: kiteClient,
        Logger:     logger,
    }).SetupWithManager(mgr); err != nil {
        setupLog.Error(err, "unable to create controller", "controller", "YourResource")
        os.Exit(1)
    }

    // ... rest of main function ...
}
```

---

### Step 5: Add Imports & Scheme registration
Update the imports in `cmd/main.go` to include your resource types:
```go
import (
    // ... existing imports ...

    // For built-in Kubernetes resources:
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"

    // For Custom Resources, you have two options:

    // Option 1: Import from existing module (if CRDs come from another operator)
    tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"

    // Option 2: Generate and import your own types (example)
    // myappv1 "github.com/your-org/your-api/pkg/apis/myapp/v1"
)

func init() {
    utilruntime.Must(clientgoscheme.AddToScheme(scheme))
    utilruntime.Must(tektonv1.AddToScheme(scheme))

    // Add your resource schemes
    // For built-in resources:
    utilruntime.Must(appsv1.AddToScheme(scheme))

    // For custom resources:
    // utilruntime.Must(myappv1.AddToScheme(scheme))
}
```

---

### Step 6: Write teests
Use the [PipelineRun controller tests](../internal/controller/pipelinerun_controller_test.go) as a guide.

Run all controller tests with:
```go
go test ./packages/operator/internal/controller/...
```

---

## Working with Custom Resources
:construction: **TODO** - Waiting for teams to integrate with their custom resources first.

## Configuration
### Environment variables
- `KITE_API_URL`: API URL for Kite backend (default: `http://localhost:8080`)
- `ENABLE_HTTP2`: Enable HTTP/2 (default: `true`, set `false` for local dev)

### RBAC Permissions
Add RBAC rules with `+kubebuilder:rbac` annotations. Example for Deployments.
```go
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
```


## Recommended Best Practices
1. **Follow the existing pattern** - start with the [PipelineRun controller](../internal/controller/pipelinerun_controller.go).
2. **Add proper logging** - use structured logs with key fields (name, namespace).
3. **Handle errors gracefully** - requeue with `RequeueAfter` when needed.
4. **Test thoroughly** - write unit tests and verify against a real cluster.
5. **Use least privilege RBAC** - only request what your controller needs.
6. **Avoid noisy issues** - don’t open issues for every transient (impermanent) state.
7. **Enrich payloads** - include enough context for quick triage.
8. **Update docs** - document new payloads in [Webhooks.md](../../backend/docs/Webhooks.md)
