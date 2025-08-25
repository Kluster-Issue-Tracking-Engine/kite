# Kite Bridge Operator
The **Kite Bridge Operator** automates the monitoring, creation and resolution of issues within Konflux.
It integrates with the [Kite Backend service](../backend/) to persist issue records.

## Description
This project implements the *bridge operator* pattern:
> A Kubernetes "bridge" operator connects a Kubernetes environment with external systems or resources not natively managed by Kubernetes. Its primary role is to extend the Kubernetes control plane so it can interact with external entities, typically through exposed APIs.

In this case, the operator monitors resources in the Konflux cluster. It detects successes or failures and reports failures back to the Kite backend.

You can run a [local demo](./docs/Demo.md) to see how this all works.

## Getting Started

### Prerequisites
- go version v1.24.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push the operator image:**

```sh
make docker-build docker-push IMG=<some-registry>/operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs:**

```sh
make install
```

**Deploy the operator:**

```sh
make deploy IMG=<some-registry>/operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Apply sample resources**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Remove sample resources:**

```sh
kubectl delete -k config/samples/
```

**Uninstall the CRDs:**

```sh
make uninstall
```

**Remove the operator:**

```sh
make undeploy
```

## Project Distribution

You can distribute this operator in two ways:

### Option 1: YAML bundle

1. Build an installer containing all Kubernetes resources:

```sh
make build-installer IMG=<some-registry>/operator:tag
```

This generates `dist/install.yaml`, a Kustomize-built manifest with all required resources.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/operator/<tag or branch>/dist/install.yaml
```

### Option 2: Helm Chart

1. Generate a Helm chart using the optional helm plugin

```sh
operator-sdk edit --plugins=helm/v1-alpha
```

2. The chart will be created under `dist/chart`. Users can install the operator from this chart.

**NOTE:** If you make changes to the project, regenerate the chart using the same command.
If you’ve created webhooks, use the `--force` flag and manually re-apply any custom configuration to `dist/chart/values.yaml` or `dist/chart/manager/manager.yaml`.

## Contributing
To contribute, please create a Jira issue for your controller include details about:
- The resources your controller watches
- How it interacts wtih the Kite backend service

**NOTE:** Run `make help` to see al available `make` targets.

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

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

