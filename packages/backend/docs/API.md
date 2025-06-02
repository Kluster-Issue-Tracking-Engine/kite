# Konflux Issues API Documentation

## Table of Contents
- [Overview](#overview)
- [Authentication & Authorization](#authentication--authorization)
- [Data Models](#data-models)
- [API Endpoints](#api-endpoints)
- [Webhooks](#webhooks)
- [Usage Examples](#usage-examples)

> **NOTE**: This API doc is temporary and will be auto-generated in the future.

## Overview

The Konflux Issues API is a RESTful service for managing and tracking issues within Konflux CI/CD pipelines and components.

It provides endpoints for creating, reading, updating, and deleting issues, as well as webhook endpoints for automated issue management.

The goal of this project is for this service to be the backend to the Konflux Issues Dashboard.

The Konflux Issues Dashboard will function like a car dashboard - a centralized place to view and monitor issues (Specifically issues related to building and shipping applications in Konflux).

**Base URL:** `http://localhost:3000/api/v1` (development)
**API Version:** v1
**Content-Type:** `application/json`

## Authentication & Authorization

The API will use Kubernetes RBAC for namespace-based access control. Users must have access to the Kubernetes namespace to interact with issues in that namespace.

## Data Models

### Issue

An **issue** represents a problem or concern in the system. This problem can be related to resources in the Konflux (failed PipelineRuns, builds, etc) or outside of the Konflux cluster (MintMaker).

```json
{
  "id": "uuid",
  "title": "string",
  "description": "string",
  "severity": "info|minor|major|critical",
  "issueType": "build|test|release|dependency|pipeline",
  "state": "ACTIVE|RESOLVED",
  "detectedAt": "2025-01-01T12:00:00Z",
  "resolvedAt": "2025-01-01T13:00:00Z",
  "namespace": "string",
  "scopeId": "uuid",
  "scope": {
    "id": "uuid",
    "resourceType": "string",
    "resourceName": "string",
    "resourceNamespace": "string"
  },
  "links": [
    {
      "id": "uuid",
      "title": "string",
      "url": "string",
      "issueId": "uuid"
    }
  ],
  "relatedFrom": [],
  "relatedTo": [],
  "createdAt": "2025-01-01T12:00:00Z",
  "updatedAt": "2025-01-01T12:00:00Z"
}
```

### Enums

**Severity:**
- `info` - Informational issues
- `minor` - Minor issues that don't block functionality
- `major` - Major issues that impact functionality
- `critical` - Critical issues that block functionality

**Issue Type:**
- `build` - Build-related issues
- `test` - Test-related issues
- `release` - Release and deployment issues
- `dependency` - Dependency-related issues
- `pipeline` - Pipeline execution issues

**State:**
- `ACTIVE` - Issue is currently active/unresolved
- `RESOLVED` - Issue has been resolved

## API Endpoints

### Health & System

#### GET /health
Returns service health status.

**Response:**
```json
{
  "status": "UP",
  "message": "Service is healthy",
  "timestamp": "2025-01-01T12:00:00Z"
}
```

#### GET /version
Returns service version information.

**Response:**
```json
{
  "version": "1.0.0",
  "name": "Konflux Issues API",
  "description": "API for managing issues in Konflux"
}
```

---

### Issues

#### GET /api/v1/issues
Retrieve a list of issues with optional filtering.

**Query Parameters:**
- `namespace` (required) - Kubernetes namespace
- `severity` (optional) - Filter by severity: `info|minor|major|critical`
- `issueType` (optional) - Filter by type: `build|test|release|dependency|pipeline`
- `state` (optional) - Filter by state: `ACTIVE|RESOLVED`
- `resourceType` (optional) - Filter by resource type
- `resourceName` (optional) - Filter by resource name
- `search` (optional) - Search in title and description
- `limit` (optional, default: 50) - Number of results to return
- `offset` (optional, default: 0) - Number of results to skip

**Example Request:**
```bash
GET /api/v1/issues?namespace=team-alpha&severity=critical&limit=10
```

**Response:**
```json
{
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Frontend build failed due to dependency conflict",
      "description": "The build process failed...",
      "severity": "major",
      "issueType": "build",
      "state": "ACTIVE",
      "detectedAt": "2025-01-01T12:00:00Z",
      "namespace": "team-alpha",
      "scope": {
        "resourceType": "component",
        "resourceName": "frontend-ui",
        "resourceNamespace": "team-alpha"
      },
      "links": [],
      "createdAt": "2025-01-01T12:00:00Z",
      "updatedAt": "2025-01-01T12:00:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

#### POST /api/v1/issues
Create a new issue.

**Request Body:**
```json
{
  "title": "string (required)",
  "description": "string (required)",
  "severity": "info|minor|major|critical (required)",
  "issueType": "build|test|release|dependency|pipeline (required)",
  "state": "ACTIVE|RESOLVED (optional, default: ACTIVE)",
  "namespace": "string (required)",
  "scope": {
    "resourceType": "string (required)",
    "resourceName": "string (required)",
    "resourceNamespace": "string (optional, defaults to namespace)"
  },
  "links": [
    {
      "title": "string (required)",
      "url": "string (required)"
    }
  ]
}
```

**Response:** `201 Created`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Frontend build failed",
  // ... full issue object
}
```

#### GET /api/v1/issues/:id
Retrieve a specific issue by ID.

**Path Parameters:**
- `id` (required) - Issue UUID

**Query Parameters:**
- `namespace` (optional) - Namespace for access control

**Response:** `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Frontend build failed",
  // ... full issue object with related issues and links
}
```

**Error Responses:**
- `404 Not Found` - Issue not found
- `403 Forbidden` - Access denied to namespace

#### PUT /api/v1/issues/:id
Update an existing issue.

**Path Parameters:**
- `id` (required) - Issue UUID

**Query Parameters:**
- `namespace` (optional) - Namespace for access control

**Request Body (all fields optional):**
```json
{
  "title": "string",
  "description": "string",
  "severity": "info|minor|major|critical",
  "issueType": "build|test|release|dependency|pipeline",
  "state": "ACTIVE|RESOLVED",
  "resolvedAt": "2025-01-01T13:00:00Z",
  "links": [
    {
      "title": "string (required)",
      "url": "string (required)"
    }
  ]
}
```

**Response:** `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  // ... updated issue object
}
```

#### DELETE /api/v1/issues/:id
Delete an issue and all related data.

**Path Parameters:**
- `id` (required) - Issue UUID

**Query Parameters:**
- `namespace` (optional) - Namespace for access control

**Response:** `204 No Content`

#### POST /api/v1/issues/:id/resolve
Mark an issue as resolved.

**Path Parameters:**
- `id` (required) - Issue UUID

**Query Parameters:**
- `namespace` (optional) - Namespace for access control

**Response:** `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "state": "RESOLVED",
  "resolvedAt": "2025-01-01T13:00:00Z",
  // ... full updated issue object
}
```

#### POST /api/v1/issues/:id/related
Create a relationship between two issues.

**Path Parameters:**
- `id` (required) - Source issue UUID

**Request Body:**
```json
{
  "relatedId": "uuid (required)"
}
```

**Response:** `201 Created`
```json
{
  "message": "Relationship created"
}
```

**Error Responses:**
- `404 Not Found` - One or both issues not found
- `409 Conflict` - Relationship already exists

#### DELETE /api/v1/issues/:id/related/:relatedId
Remove a relationship between issues.

**Path Parameters:**
- `id` (required) - Source issue UUID
- `relatedId` (required) - Target issue UUID

**Response:** `204 No Content`

---

## Webhooks

On top of the base API, custom webhook endpoints can be created for automatic issue creation and resolving based on your situation or workflow.

### Tekton PipelineRun Webhook endpoints

This webhook endpoint makes it easy to create and resolve issues based on a PipelineRun failure and success.

A [live demo](https://drive.google.com/file/d/19a5j_hOxEY0r2wcchGdbLAEDfatDKnXB/view?usp=sharing) for this webhook endpoint is available.

The file used in that demo can be found [here](../examples/issue-service-pr-example.yml)

#### POST /api/v1/webhooks/pipeline-failure
Handle pipeline failure events and create issues automatically.

**Query Parameters:**
- `namespace` (required) - Kubernetes namespace

**Request Body:**
```json
{
  "pipelineName": "string (required)",
  "namespace": "string (required)",
  "failureReason": "string (required)",
  "runId": "string (optional)",
  "logsUrl": "string (optional)"
}
```

**Response:** `201 Created`
```json
{
  "status": "success",
  "issue": {
    // ... created or updated issue object
  }
}
```

**Behavior:**
- Creates a new issue with `issueType: "pipeline"` and `severity: "major"`
- If a duplicate issue exists for the same pipeline/namespace, updates the existing issue
- Automatically generates links to pipeline logs

#### POST /api/v1/webhooks/pipeline-success
Handle pipeline success events and resolve related issues.

**Query Parameters:**
- `namespace` (required) - Kubernetes namespace

**Request Body:**
```json
{
  "pipelineName": "string (required)",
  "namespace": "string (required)"
}
```

**Response:** `200 OK`
```json
{
  "status": "success",
  "message": "Resolved x issue(s) for pipeline test-pipeline"
}
```

**Behavior:**
- Resolves all active pipeline issues for the specified pipeline and namespace
- Sets `state: "RESOLVED"` and `resolvedAt` timestamp

---

## Usage Examples

### Live Demo
In case you missed in above, there is a short, [live demo](https://drive.google.com/file/d/19a5j_hOxEY0r2wcchGdbLAEDfatDKnXB/view?usp=sharing) available for watching.
### Creating A Build Issue

```bash
curl -X POST http://localhost:3000/api/v1/issues \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Frontend build failed due to dependency conflict",
    "description": "React version mismatch causing build failures",
    "severity": "major",
    "issueType": "build",
    "namespace": "team-alpha",
    "scope": {
      "resourceType": "component",
      "resourceName": "frontend-ui"
    },
    "links": [
      {
        "title": "Build Logs",
        "url": "https://jenkins.example.com/job/frontend/123"
      }
    ]
  }'
```

### Searching for Critical Issues

```bash
curl "http://localhost:3000/api/v1/issues?namespace=team-alpha&severity=critical&state=ACTIVE"
```

### Pipeline Failure Webhook

```bash
curl -X POST "http://localhost:3000/api/v1/webhooks/pipeline-failure?namespace=team-alpha" \
  -H "Content-Type: application/json" \
  -d '{
    "pipelineName": "frontend-deploy",
    "namespace": "team-alpha",
    "failureReason": "Deployment timeout",
    "runId": "run-123",
    "logsUrl": "https://pipeline.example.com/runs/123"
  }'
```

---