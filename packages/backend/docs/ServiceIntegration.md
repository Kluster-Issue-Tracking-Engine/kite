# Service Integration Guide

This guide explains how external services can integrate with Kite to automatically create and resolve issues based on their workflows.

## Overview

Kite allows any service that can send HTTP requests to create and manage issues

There are two main ways to integrate:
1. **Direct API Integration** - Use the standard REST API endpoints
2. **Custom Webhook Endpoints** - Request and use specialized webhook endpoints for specific workflows

## Custom Webhook Endpoints

Webhook endpoints provide a simple way to integrate common workflows. They handle the complexity of issue creation, duplicate checking and automatic resolution.

### Example Webhook endpoints

The following are example webhook endpoints for pipeline failures. There is a short, [live demo](https://drive.google.com/file/d/19a5j_hOxEY0r2wcchGdbLAEDfatDKnXB/view?usp=sharing) if you're interested in seeing them in action.

#### Pipeline Failure Webhook
**Endpoint**: `POST /api/v1/webhooks/pipeline-failure`

Creates issues when pipelines fail. Automatically handles duplicate checking and updates existing issues.

**Request Payload**:
```json
{
  "pipelineName": "frontend-build",
  "namespace": "team-alpha",
  "failureReason": "Dependency conflict with React version",
  "runId": "run-123",
  "logsUrl": "https://your-ci.com/logs/run-123"
}
```

**What it does**:
- Creates an issue with title "Pipeline run failed: frontend-build"
- Sets issue type to "pipeline" and severity "major"
- Links to pipeline logs for easy debugging
- If a similar issue already exists for the same pipeline, it updates that issue instead of creating a duplicate

Internally the issue generated from that payload looks something like this:
```json
{
	"id": "986d686c-bce6-44be-b6ba-a7b5b88eec58",
	"title": "Pipeline run failed: frontend-build",
	"description": "The pipeline run frontend-build failed with reason: Dependency conflict with React version",
	"severity": "major",
	"issueType": "pipeline",
	"state": "ACTIVE",
	"detectedAt": "2025-06-17T18:13:29.007244Z",
	"resolvedAt": null,
	"namespace": "team-alpha",
	"scopeId": "1a483caf-f349-4a9d-879a-df74a2b55eb3",
	"scope": {
		"id": "1a483caf-f349-4a9d-879a-df74a2b55eb3",
		"resourceType": "pipelinerun",
		"resourceName": "frontend-build",
		"resourceNamespace": "team-alpha"
	},
	"links": [
		{
			"id": "d4ff5876-8c88-4703-9836-f6ac66e73ba0",
			"title": "Pipeline Run Logs",
			"url": "https://your-ci.com/logs/run-123",
			"issueId": "986d686c-bce6-44be-b6ba-a7b5b88eec58"
		}
	],
	"relatedFrom": [],
	"relatedTo": [],
	"createdAt": "2025-06-17T18:13:29.008239Z",
	"updatedAt": "2025-06-17T18:13:29.008239Z"
}
```

### Pipeline Success Webhook
**Endpoint**: `POST /api/v1/webhooks/pipeline-success`
Automatically resolves pipeline issues when pipelines succeed.

**Request Payload**:
```json
{
  "pipelineName": "frontend-build",
  "namespace": "team-alpha"
}
```

**What it does**:
- Finds all active issues related to the specified `pipelineName`
- Marks them as "RESOLVED"
- Sets the resolution timestamp

After hitting this endpoint, the issue created from the failure endpoint will be updated:
```json
{
	"id": "986d686c-bce6-44be-b6ba-a7b5b88eec58",
	"title": "Pipeline run failed: frontend-build",
	"description": "The pipeline run frontend-build failed with reason: Dependency conflict with React version",
	"severity": "major",
	"issueType": "pipeline",
	"state": "RESOLVED",
	"detectedAt": "2025-06-17T18:13:29.007244Z",
	"resolvedAt": "2025-06-17T18:32:36.107527Z",
	"namespace": "team-alpha",
	"scopeId": "1a483caf-f349-4a9d-879a-df74a2b55eb3",
	"scope": {
		"id": "1a483caf-f349-4a9d-879a-df74a2b55eb3",
		"resourceType": "pipelinerun",
		"resourceName": "frontend-build",
		"resourceNamespace": "team-alpha"
	},
	"links": [
		{
			"id": "d4ff5876-8c88-4703-9836-f6ac66e73ba0",
			"title": "Pipeline Run Logs",
			"url": "https://your-ci.com/logs/run-123",
			"issueId": "986d686c-bce6-44be-b6ba-a7b5b88eec58"
		}
	],
	"relatedFrom": [],
	"relatedTo": [],
	"createdAt": "2025-06-17T18:13:29.008239Z",
	"updatedAt": "2025-06-17T18:32:36.107527Z"
}
```

## Creating Custom Webhook Endpoints
You can request custom webhook endpoints for your specific workflow. Here are some example endpoints and payloads:

### Example: Build Failure
```json
// POST /api/v1/webhooks/build-failure
{
  "componentName": "user-service",
  "namespace": "team-beta",
  "buildId": "build-456",
  "errorMessage": "Compilation failed: missing dependency",
  "buildLogsUrl": "https://build-system.com/logs/456"
}
```

### Example: Deployment Failure
```json
// POST /api/v1/webhooks/deployment-failure
{
	"applicationName": "e-commerce-app",
  "namespace": "team-gamma",
  "deploymentId": "deploy-789",
  "failureReason": "Resource quota exceeded",
  "dashboardUrl": "https://dashboard.com/deployments/789"
}
```

### Example: Security Scan Failure
```json
// POST /api/v1/webhooks/security-scan-failure
{
  "componentName": "api-gateway",
  "namespace": "team-delta",
  "scanId": "scan-321",
  "vulnerabilities": ["CVE-2024-1234", "CVE-2024-5678"],
  "reportUrl": "https://security.com/reports/321"
}
```

## Issue Grouping with Scope Objects
The `scope` object is the key to how Kite groups and manages related issues. It defines what resource an issue is related to.

### Scope Object Structure
```json
{
	"resourceType": "component", // What kind of resource
	"resourceName": "user-service", // Specific resource name
	"resourceNamespace": "team-beta", // Where the resource lives
}
```

### Common Resource Types
- `pipelinerun` - Issues related to pipeline executions
- `component` - Issues with application components
- `application` - Issues with entire applications
- `workspace` - Issues affecting workspaces
- `environment` - Environment-specific issues

### How grouping works

Issues with the same scope are considered related:
```json
// These two issues will be grouped together because they have the same scope
{
  "title": "Build failed for user-service",
	...
  "scope": {
    "resourceType": "component",
    "resourceName": "user-service",
    "resourceNamespace": "team-beta"
  }
	...
}

{
  "title": "Tests failing for user-service",
	...
  "scope": {
    "resourceType": "component",
    "resourceName": "user-service",
    "resourceNamespace": "team-beta"
  }
	...
}
```

### Benefits of Scope-Based Grouping

- **Prevents Duplicates** - Won't create multiple active issues for the same resource
- **Easy Filtering** - Find all issues for a specific component or application
- **Automatic Resolution** - Resolve all issues for a resource when it's fixed
- **Better Organization** - Issues are naturally organized by what they affect

### Automatic Issue Resolution (via custom webhooks):

Automatic resolution uses the `scope` object to find and resolve related issues when problems are fixed.

Going back to the pipeline webhook example, lets create two issues with the same `scope`:

**Payloads**
```json
// POST /api/v1/webhooks/pipeline-failure

// Payload 1
{
  "pipelineName": "frontend-build",
  "namespace": "team-alpha",
  "failureReason": "Dependency conflict with React version",
  "runId": "run-123",
  "logsUrl": "https://your-ci.com/logs/run-123"
}

// Payload 2
{
  "pipelineName": "frontend-build",
  "namespace": "team-alpha",
  "failureReason": "Yarn version outdated, needs update",
  "runId": "run-456",
  "logsUrl": "https://your-ci.com/logs/run-456"
}
```

These requests would generate the following issues:
```json
// From Payload 1
{
	"id": "4a99b4a0-fbde-4afd-9bfc-5307178ababb",
	"title": "Pipeline run failed: frontend-build",
	"description": "The pipeline run frontend-build failed with reason: Dependency conflict with React version",
	"severity": "major",
	"issueType": "pipelinerun",
	"state": "ACTIVE",
	"detectedAt": "2025-06-17T19:49:31.248829Z",
	"resolvedAt": null,
	"namespace": "team-alpha",
	"scopeId": "abb27ca3-fd9d-48e1-8da7-0a776a4915d1",
	"scope": {
		"id": "abb27ca3-fd9d-48e1-8da7-0a776a4915d1",
		"resourceType": "pipelinerun",
		"resourceName": "frontend-build",
		"resourceNamespace": "team-alpha"
	},
	"links": [
		{
			"id": "8eb2802b-a3c4-4158-8df2-73396c646572",
			"title": "Pipeline Run Logs",
			"url": "https://your-ci.com/logs/run-123",
			"issueId": "4a99b4a0-fbde-4afd-9bfc-5307178ababb"
		}
	],
	"relatedFrom": [],
	"relatedTo": [],
	"createdAt": "2025-06-17T19:49:31.249607Z",
	"updatedAt": "2025-06-17T19:49:31.249607Z"
},
// From Payload 2
{
	"id": "58c96080-aba8-40c3-a9e9-585a5fd64692",
	"title": "Pipeline run failed: frontend-build",
	"description": "The pipeline run frontend-build failed with reason: Yarn version outdated, needs update",
	"severity": "major",
	"issueType": "pipelinerun",
	"state": "ACTIVE",
	"detectedAt": "2025-06-17T19:36:07.860624Z",
	"resolvedAt": null,
	"namespace": "team-alpha",
	"scopeId": "390e6c1d-7078-4581-9d35-8c49bce42301",
	"scope": {
		"id": "390e6c1d-7078-4581-9d35-8c49bce42301",
		"resourceType": "pipelinerun",
		"resourceName": "frontend-build",
		"resourceNamespace": "team-alpha"
	},
	"links": [
		{
			"id": "abb548f7-16a8-465e-94d4-6ec06fcfdef9",
			"title": "Pipeline Run Logs",
			"url": "https://your-ci.com/logs/run-456",
			"issueId": "58c96080-aba8-40c3-a9e9-585a5fd64692"
		}
	],
	"relatedFrom": [],
	"relatedTo": [],
	"createdAt": "2025-06-17T19:36:07.861097Z",
	"updatedAt": "2025-06-17T19:42:06.214236Z"
}
```

Note that both records have the same scope:
```json
...
	"scope": {
		"id": "390e6c1d-7078-4581-9d35-8c49bce42301",
		"resourceType": "pipelinerun",
		"resourceName": "frontend-build",
		"resourceNamespace": "team-alpha"
	},
...
```

**NOTE:** Because this webhook endpoint is customized for pipeline runs, the `resourceType` is automatically set to `pipelinerun`. Only the **pipeline name** and **namespace** is needed in the payload.

When the pipeline with the name `frontend-build` passes, both these issues should get resolved:

**Request Payload**
```json
// POST /api/v1/webhooks/pipeline-success
{
  "pipelineName": "frontend-build",
  "namespace": "team-alpha"
}
```

**Response**
```json
{
	"message": "Resolved 2 issue(s) for pipeline frontend-build",
	"status": "success"
}
```

**Updated issues**:
```json
// Payload 1
{
	"id": "4a99b4a0-fbde-4afd-9bfc-5307178ababb",
	"title": "Pipeline run failed: frontend-build",
	"description": "The pipeline run frontend-build failed with reason: Dependency conflict with React version",
	"severity": "major",
	"issueType": "pipelinerun",
	"state": "RESOLVED",
	"detectedAt": "2025-06-17T19:49:31.248829Z",
	"resolvedAt": "2025-06-17T19:57:19.634641Z",
	"namespace": "team-alpha",
	"scopeId": "abb27ca3-fd9d-48e1-8da7-0a776a4915d1",
	"scope": {
		"id": "abb27ca3-fd9d-48e1-8da7-0a776a4915d1",
		"resourceType": "pipelinerun",
		"resourceName": "frontend-build",
		"resourceNamespace": "team-alpha"
	},
	"links": [
		{
			"id": "8eb2802b-a3c4-4158-8df2-73396c646572",
			"title": "Pipeline Run Logs",
			"url": "https://your-ci.com/logs/run-123",
			"issueId": "4a99b4a0-fbde-4afd-9bfc-5307178ababb"
		}
	],
	"relatedFrom": [],
	"relatedTo": [],
	"createdAt": "2025-06-17T19:49:31.249607Z",
	"updatedAt": "2025-06-17T19:57:19.634641Z"
},
// Payload 2
{
	"id": "58c96080-aba8-40c3-a9e9-585a5fd64692",
	"title": "Pipeline run failed: frontend-build",
	"description": "The pipeline run frontend-build failed with reason: Yarn version outdated, needs update",
	"severity": "major",
	"issueType": "pipelinerun",
	"state": "RESOLVED",
	"detectedAt": "2025-06-17T19:36:07.860624Z",
	"resolvedAt": "2025-06-17T19:57:19.634641Z",
	"namespace": "team-alpha",
	"scopeId": "390e6c1d-7078-4581-9d35-8c49bce42301",
	"scope": {
		"id": "390e6c1d-7078-4581-9d35-8c49bce42301",
		"resourceType": "pipelinerun",
		"resourceName": "frontend-build",
		"resourceNamespace": "team-alpha"
	},
	"links": [
		{
			"id": "abb548f7-16a8-465e-94d4-6ec06fcfdef9",
			"title": "Pipeline Run Logs",
			"url": "https://your-ci.com/logs/run-456",
			"issueId": "58c96080-aba8-40c3-a9e9-585a5fd64692"
		}
	],
	"relatedFrom": [],
	"relatedTo": [],
	"createdAt": "2025-06-17T19:36:07.861097Z",
	"updatedAt": "2025-06-17T19:57:19.634641Z"
}
```

### Requesting Custom Webhooks
To request a custom webhook endpoint for your service:
1. **Identify Your Workflow** - What events do you want to track?
2. **Define the Scope** - What resources are affected by these events?
3. **Specify the Data** - What information needs to be captured?
4. **Plan Resolution** - How will issues be automatically resolved?

Here is a template to help with the request:
```markdown
## Webhook Request: [Workflow Name]

**Purpose:** Track [type of issues] for [system/service]

**Failure Endpoint:** POST /api/v1/webhooks/[workflow-name]-failure
**Success Endpoint:** POST /api/v1/webhooks/[workflow-name]-success

**Failure Request Body:**
{
  "resourceName": "string",
  "namespace": "string",
  "failureReason": "string",
  // ... other relevant fields
}

**Success Request Body**:
{
  "resourceName": "string",
  "namespace": "string"
}

**Scope Mapping**:
- resourceType: "[your-resource-type]"
- resourceName: from request body
- resourceNamespace: from request body

**Issue Details**:
- issueType: "[build|test|release|dependency|pipeline]"
- severity: "[info|minor|major|critical]"
- title format: "[Your title template]"

```

### Practical examples
#### CI/CD Pipeline Integration:
```yaml
# Tekton Task example
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: notify-kite
spec:
  params:
    - name: pipeline-name
    - name: status
    - name: failure-reason
  steps:
    - name: notify
      image: curlimages/curl
      script: |
        if [ "$(params.status)" = "Failed" ]; then
          curl -X POST http://kite-api/api/v1/webhooks/pipeline-failure \
            -H "Content-Type: application/json" \
            -d '{
              "pipelineName": "$(params.pipeline-name)",
              "namespace": "$(context.taskRun.namespace)",
              "failureReason": "$(params.failure-reason)"
            }'
        else
          curl -X POST http://kite-api/api/v1/webhooks/pipeline-success \
            -H "Content-Type: application/json" \
            -d '{
              "pipelineName": "$(params.pipeline-name)",
              "namespace": "$(context.taskRun.namespace)"
            }'
        fi
```

#### Codebase integration

**Python**:
```python
import requests

def notify_kite_build_status(component_name, namespace, status, build_id, error_msg=None):
    base_url = "http://kite-api/api/v1/webhooks"

    if status == "failed":
        response = requests.post(f"{base_url}/build-failure", json={
            "componentName": component_name,
            "namespace": namespace,
            "buildId": build_id,
            "errorMessage": error_msg,
            "buildLogsUrl": f"https://builds.com/logs/{build_id}"
        })
    else:
        response = requests.post(f"{base_url}/build-success", json={
            "componentName": component_name,
            "namespace": namespace
        })

    return response.status_code == 200
```

**Go**:
```golang
// KiteClient handles integration with Kite API
type KiteClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewKiteClient creates a new Kite API client
func NewKiteClient(baseURL string) *KiteClient {
	return &KiteClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Build webhook request structures
type BuildFailureRequest struct {
	ComponentName string `json:"componentName"`
	Namespace     string `json:"namespace"`
	BuildID       string `json:"buildId"`
	ErrorMessage  string `json:"errorMessage"`
	BuildLogsURL  string `json:"buildLogsUrl,omitempty"`
}

type BuildSuccessRequest struct {
	ComponentName string `json:"componentName"`
	Namespace     string `json:"namespace"`
}

// Generic webhook response
type WebhookResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Issue   interface{} `json:"issue,omitempty"`
}

// ReportBuildFailure reports a build failure to Kite
func (k *KiteClient) ReportBuildFailure(ctx context.Context, req BuildFailureRequest) error {
	endpoint := fmt.Sprintf("%s/api/v1/webhooks/build-failure?namespace=%s", k.baseURL, req.Namespace)

	var response WebhookResponse
	if err := k.makeRequest(ctx, "POST", endpoint, req, &response); err != nil {
		return fmt.Errorf("failed to report build failure: %w", err)
	}

	log.Printf("Build failure reported successfully for component: %s", req.ComponentName)
	return nil
}

// ReportBuildSuccess reports a build success to Kite
func (k *KiteClient) ReportBuildSuccess(ctx context.Context, req BuildSuccessRequest) error {
	endpoint := fmt.Sprintf("%s/api/v1/webhooks/build-success?namespace=%s", k.baseURL, req.Namespace)

	var response WebhookResponse
	if err := k.makeRequest(ctx, "POST", endpoint, req, &response); err != nil {
		return fmt.Errorf("failed to report build success: %w", err)
	}

	log.Printf("Build success reported successfully for component: %s", req.ComponentName)
	return nil
}

// NotifyBuildStatus reports build status to Kite based on success/failure
func (k *KiteClient) NotifyBuildStatus(ctx context.Context, componentName, namespace, buildID string, success bool, errorMessage, logsURL string) error {
	if success {
		return k.ReportBuildSuccess(ctx, BuildSuccessRequest{
			ComponentName: componentName,
			Namespace:     namespace,
		})
	}

	return k.ReportBuildFailure(ctx, BuildFailureRequest{
		ComponentName: componentName,
		Namespace:     namespace,
		BuildID:       buildID,
		ErrorMessage:  errorMessage,
		BuildLogsURL:  logsURL,
	})
}
```