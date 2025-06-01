# Kite :kite:

Konflux Issue Tracking Engine

:construction: **Under Construction – Currently a Proof of Concept** :construction:

## About

Kite is a centralized service for tracking Konflux-related issues that may disrupt your ability to build and deploy applications.

These issues might include:

- PipelineRun failures  
- Release errors  
- MintMaker issues  
- Cluster-wide problems  

Kite **does not** monitor cluster resources to detect these issues.  
That responsibility is left to you, based on your specific situation and workflow.

Instead, Kite acts as a **central information store**, offering the following features:

## Features

- **Issue Tracking**: Track build/test failures, release problems, and more in a centralized, extendable service.  
- **CLI Integration**: Access and manage issues from your terminal or as a `kubectl` plugin.  
- **Namespace Isolation**: Issues are scoped to Kubernetes namespaces for better security.  
- **Automation-Friendly**: Supports webhooks for automatic issue creation and resolution.  
- **API Access**: RESTful API for integration with external tools.  

## Components

This monorepo is structured around two primary components:

- `packages/backend`: A Go-based `gin-gonic` server with a PostgreSQL database.  
- `packages/cli`: A Go-based CLI tool that can run standalone or as a `kubectl` plugin.  

## Prerequisites

To work with this project, ensure you have the following installed:

- [Docker](https://docs.docker.com/get-docker/) or [Podman](https://podman.io/docs/installation)  
- [Go](https://golang.org/doc/install) v1.23 or later  
- [Make](https://www.gnu.org/software/make/)  
- [Minikube](https://minikube.sigs.k8s.io/docs/start/) – for local development  
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) – for local development  

