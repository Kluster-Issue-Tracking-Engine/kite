# Demo with Sample Resources
This demo shows how the Kite Bridge Operator can monitor, create and resolve issues as they occur in a cluster.

You'll run a sample `PipelineRun` that fails and the included Pipeline Run controller will report the failure to the [Kite backend](../backend/).

You'll then modify the Pipeline manifest so it succeeds and see the operator resolve the issue.

## Requirements
- [Tekton](https://tekton.dev/docs/installation/) and [Tekton CLI](https://tekton.dev/docs/cli/) installed
- A local instance of the [Kite backend service](../backend/README.md) running on `localhost:8080`

## Run the Demo
1. **Set the local development environment variables:**
```sh
export KITE_API_URL="http://localhost:8080"
export ENABLE_HTTP2=false
```

2. **Run the Operator locally (without deploying):**
```sh
make run
```

3. **Apply the sample manifests:**
```sh
kubectl apply -k config/samples/
```

4. **Run the sample Pipeline:**
```sh
tkn p start simple-pipeline -n tekton-pipelines
```

5. **Observe the Operator logs:**
You should see the operator detect the failure and report it to Kite.
```sh
{"level":"info","msg":"Processing failed PipelineRun","namespace":"tekton-pipelines","pipeline_run":"simple-pipeline-run-f4tk9","status":"failed","time":"2025-08-20T13:58:43-04:00"}
{"level":"info","msg":"Successfully sent request to KITE","operation":"pipeline-failure","status_code":201,"time":"2025-08-20T13:58:43-04:00"}
```

6. **Update the [sample manifest](../config/samples/tekton_v1_pipelinerun.yaml) so the PipelineRun succeeds:**
```yaml
apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: simple-pipeline
  namespace: tekton-pipelines
spec:
  tasks:
  - name: echo-task
    taskSpec:
      steps:
      - image: busybox
        name: echo-message
        script: |
          #!/bin/sh
          echo "Hello, Tekton!"
          exit 0 # <- Updated to succeed
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop: ["ALL"]
          runAsNonRoot: true
          runAsUser: 1000
          runAsGroup: 3000
          seccompProfile:
            type: "RuntimeDefault"
```

7. **Apply the update:**
```sh
kubectl apply -k config/samples
```

8. **Run the Pipeline again:**
```sh
tkn p start simple-pipeline -n tekton-pipelines
```

9. **Confirm success in the operator logs:**
```sh
{"level":"info","msg":"Processing successful PipelineRun","namespace":"tekton-pipelines","pipeline_run":"simple-pipeline-run-v4j9h","status":"succeeded","time":"2025-08-20T13:58:43-04:00"}
{"level":"info","msg":"Successfully sent request to KITE","operation":"pipeline-success","status_code":200,"time":"2025-08-20T13:58:43-04:00"}
{"id":"ed361d7b-852c-48c8-9f16-b4cb16d2b1a1","level":"info","msg":"Successfully reported pipeline success to KITE","operation":"pipeline-success","pipeline_run":"simple-pipeline-run-v4j9h","time":"2025-08-20T13:58:43-04:00"}
```

10. **Stop the operator:**
You can stop it with `Ctrl-C`.

11. **Delete the sample resources:**
```sh
kubectl delete -k config/samples/
```