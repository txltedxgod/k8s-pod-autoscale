package controller

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/txltedxgod/k8s-pod-autoscale/pkg/metrics"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Scaler struct {
	client           kubernetes.Interface
	namespace        string
	deploymentName   string
	collector        *metrics.Collector
	interval         time.Duration
	minReplicas      int32
	maxReplicas      int32
	targetQueueDepth int
}

func NewScaler(
	client kubernetes.Interface,
	namespace string,
	deploymentName string,
	metricURL string,
	interval time.Duration,
) *Scaler {
	return &Scaler{
		client:           client,
		namespace:        namespace,
		deploymentName:   deploymentName,
		collector:        metrics.NewCollector(metricURL),
		interval:         interval,
		minReplicas:      2,
		maxReplicas:      15,
		targetQueueDepth: 100, // 1 pod per 100 items
	}
}

func (s *Scaler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.reconcile(ctx)
		}
	}
}

func (s *Scaler) reconcile(ctx context.Context) {
	queueDepth, err := s.collector.FetchCurrentMetric(ctx)
	if err != nil {
		log.Printf("[Scaler Error] Could not fetch metrics: %v\n", err)
		return
	}

	deployment, err := s.client.AppsV1().Deployments(s.namespace).Get(ctx, s.deploymentName, metav1.GetOptions{})
	if err != nil {
		log.Printf("[Scaler Error] Could not get Deployment %s/%s: %v\n", s.namespace, s.deploymentName, err)
		return
	}

	currentReplicas := int32(1)
	if deployment.Spec.Replicas != nil {
		currentReplicas = *deployment.Spec.Replicas
	}

	// Calculate desired replicas
	calculated := int32(math.Ceil(float64(queueDepth) / float64(s.targetQueueDepth)))
	if calculated < s.minReplicas {
		calculated = s.minReplicas
	}
	if calculated > s.maxReplicas {
		calculated = s.maxReplicas
	}

	if calculated != currentReplicas {
		log.Printf("[Autoscale Trigger] Queue Depth: %d | Scaling %s/%s from %d -> %d replicas\n",
			queueDepth, s.namespace, s.deploymentName, currentReplicas, calculated)

		deployment.Spec.Replicas = &calculated
		_, err := s.client.AppsV1().Deployments(s.namespace).Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			log.Printf("[Scaler Error] Failed to update replicas: %v\n", err)
		} else {
			log.Printf("[Scaler Success] Successfully scaled %s/%s to %d pods\n", s.namespace, s.deploymentName, calculated)
		}
	}
}
