# k8s-pod-autoscale

> Custom **Kubernetes Controller & Autoscaler** built with Go and `client-go` that scales Deployments dynamically based on external queue depths and custom Prometheus metrics.

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-client--go-326CE5?style=flat-square&logo=kubernetes)](https://github.com/kubernetes/client-go)
[![CI](https://img.shields.io/badge/CI-Passing-238636?style=flat-square&logo=githubactions)](https://github.com/txltedxgod/k8s-pod-autoscale/actions)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker)](https://docker.com)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)

`#kubernetes` `#k8s-controller` `#client-go` `#autoscaling` `#devops` `#cloud-native` `#golang`

---

## 🏛️ Controller Reconcile Loop

```mermaid
sequenceDiagram
    autonumber
    participant Queue as External Queue (SQS / Redis)
    participant Controller as K8s Autoscaler Controller
    participant K8sAPI as Kubernetes API Server
    participant Deployment as Target Deployment / Pods

    loop Every Poll Interval (15s)
        Controller->>Queue: Query Current Queue Depth / Lag
        Queue-->>Controller: Return Pending Message Count
        Controller->>Controller: Calculate Desired Replicas (QueueDepth / TargetValue)
        Controller->>K8sAPI: Get Current Deployment Scale Subresource
        K8sAPI-->>Controller: Current Replicas Count
        alt Desired != Current
            Controller->>K8sAPI: Update Scale Subresource (replicas = Desired)
            K8sAPI->>Deployment: Reconcile Pod Count
            Controller->>Controller: Enforce Cooldown Period (Prevent Thrashing)
        end
    end
```

---

## Features

- **Queue & Metric-Driven Scaling:** Scale workloads based on real processing backlogs (RabbitMQ, SQS, Redis Stream lag) rather than lagging CPU/Memory heuristics.
- **Custom In-Cluster Controller:** Implements `client-go` Informers, WorkQueues, and Leader Election for high-availability production clusters.
- **Flapping Prevention:** Configurable scale-up / scale-down stabilization windows and cooldown damping.
- **RBAC Manifests Included:** Production-ready `ClusterRole`, `ServiceAccount`, and `Deployment` configurations.

## Quick Start

### 1. Apply RBAC and CRD/Deployments

```bash
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deploy/deployment.yaml
```

### 2. Run Locally with Kubeconfig

```bash
go run main.go -kubeconfig=$HOME/.kube/config -namespace=default
```
