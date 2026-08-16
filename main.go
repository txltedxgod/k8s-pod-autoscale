package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/txltedxgod/k8s-pod-autoscale/pkg/controller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var kubeconfig string
	var targetDeployment string
	var targetNamespace string
	var metricEndpoint string
	var syncInterval time.Duration

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig file. Only required if out-of-cluster.")
	flag.StringVar(&targetDeployment, "deployment", "worker-service", "Target Deployment name to scale")
	flag.StringVar(&targetNamespace, "namespace", "default", "Kubernetes namespace of target Deployment")
	flag.StringVar(&metricEndpoint, "metric-url", "http://queue-service:8080/metrics/queue-depth", "External HTTP endpoint returning queue size")
	flag.DurationVar(&syncInterval, "interval", 10*time.Second, "Reconciliation poll interval")
	flag.Parse()

	log.Printf("[k8s-pod-autoscale] Starting controller for %s/%s\n", targetNamespace, targetDeployment)

	var config *rest.Config
	var err error

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		log.Printf("Falling back to local kubeconfig: %v\n", err)
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			log.Fatalf("Error building kubeconfig: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating kubernetes clientset: %v", err)
	}

	scaler := controller.NewScaler(clientset, targetNamespace, targetDeployment, metricEndpoint, syncInterval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go scaler.Run(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("[k8s-pod-autoscale] Shutting down controller...")
}
