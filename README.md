# k8s-pod-autoscale

> Lightweight custom Kubernetes controller for dynamic queue and metric-based pod autoscaling written in **Go** with `client-go`.

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-client--go-326CE5?style=flat-square&logo=kubernetes)](https://kubernetes.io)
[![Distroless](https://img.shields.io/badge/Container-Distroless-4285F4?style=flat-square&logo=google)](https://github.com/GoogleContainerTools/distroless)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)

`#kubernetes` `#k8s-controller` `#client-go` `#autoscaling` `#devops` `#cloud-native` `#golang`

---

## Overview

While standard Kubernetes Horizontal Pod Autoscaler (HPA) primarily reacts to CPU and memory, `k8s-pod-autoscale` reconciles workloads against application-level external metrics (such as message queue backlogs, pending tasks, or external HTTP gauge endpoints) to preemptively scale worker deployments.

## Architecture

```
 ┌────────────────────────┐
 │ External Metric API    │
 │ (e.g. Queue Depth: 450)│
 └───────────┬────────────┘
             │
             ▼ Poll / Inspect
 ┌────────────────────────┐         Reconcile / Scale
 │  k8s-pod-autoscale     ├──────────────────────────►  ┌─────────────────────────┐
 │  Controller (Go)       │                             │ Worker Pods (Replicas)  │
 └────────────────────────┘                             │ Pod 1  Pod 2  Pod 3 ... │
                                                        └─────────────────────────┘
```

## Quick Start

### 1. Deploy RBAC and Controller

```bash
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deploy/deployment.yaml
```

### 2. Run Locally (with local kubeconfig)

```bash
go run main.go \
  --namespace=default \
  --deployment=worker-service \
  --metric-url=http://localhost:8080/metrics/queue \
  --interval=10s
```

## Configuration Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--namespace` | `default` | Target Kubernetes namespace |
| `--deployment` | `worker-service` | Target Deployment to scale |
| `--metric-url` | `http://...` | External HTTP JSON metric source |
| `--interval` | `10s` | Reconciliation loop tick interval |
| `--kubeconfig` | (in-cluster) | Path to explicit kubeconfig for dev |
