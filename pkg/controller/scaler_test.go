package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestScaler_ReconcileScalingAndBounds(t *testing.T) {
	var queueDepth atomic.Int64
	metricSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"queue_depth": %d, "avg_latency_ms": 12.5}`, queueDepth.Load())
	}))
	defer metricSrv.Close()

	initialReplicas := int32(2)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "worker-service",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &initialReplicas,
		},
	}

	client := fake.NewSimpleClientset(dep)
	scaler := NewScaler(client, "default", "worker-service", metricSrv.URL, time.Second)
	ctx := context.Background()

	// Scale up: 550 items / 100 per pod -> ceil(5.5) = 6 replicas
	queueDepth.Store(550)
	scaler.reconcile(ctx)
	got, err := client.AppsV1().Deployments("default").Get(ctx, "worker-service", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected Get error: %v", err)
	}
	if *got.Spec.Replicas != 6 {
		t.Fatalf("expected 6 replicas for queue depth 550, got %d", *got.Spec.Replicas)
	}

	// Clamp to maxReplicas (15): 5000 items -> 50 -> clamped to 15
	queueDepth.Store(5000)
	scaler.reconcile(ctx)
	got, err = client.AppsV1().Deployments("default").Get(ctx, "worker-service", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected Get error: %v", err)
	}
	if *got.Spec.Replicas != 15 {
		t.Fatalf("expected maxReplicas=15 for queue depth 5000, got %d", *got.Spec.Replicas)
	}

	// Clamp to minReplicas (2): 10 items -> 1 -> clamped to 2
	queueDepth.Store(10)
	scaler.reconcile(ctx)
	got, err = client.AppsV1().Deployments("default").Get(ctx, "worker-service", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("unexpected Get error: %v", err)
	}
	if *got.Spec.Replicas != 2 {
		t.Fatalf("expected minReplicas=2 for queue depth 10, got %d", *got.Spec.Replicas)
	}
}
