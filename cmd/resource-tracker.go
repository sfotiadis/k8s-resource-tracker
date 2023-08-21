// Package main provides a command-line tool for starting a resource monitor for Kubernetes pods.
//
// This tool connects to a Kubernetes cluster and monitors resource usage (CPU and memory) of pods based on a given label selector.
// It uses the client-go library to interact with the Kubernetes API and the k8sresourcetracker/pkg/monitor package for monitoring.
// The monitored data is logged and can be used for resource tracking and optimization.
//
// Usage:
//
//	podresourcemonitor [flags]
//
// Flags:
//
//	kubeconfig string
//	      Path to the kubeconfig file (default is in-cluster configuration)
//	namespace string
//	      Kubernetes namespace to monitor (default is "default")
//	pod-label string
//	      Label selector for pods to monitor (e.g., "app=myapp")
//
// Example:
//
//	To start monitoring pods labeled with "app=myapp" in the "mynamespace" namespace:
//	podresourcemonitor kubeconfig=/path/to/kubeconfig namespace=mynamespace pod-label=app=myapp
package main

import (
	"flag"
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8sresourcetracker/pkg/monitor"
)

var (
	kubeconfig string
	namespace  string
	podLabel   string
)

func init() {
	// Configure command-line flags
	flag.StringVar(&kubeconfig, "kubeconfig", monitor.GetKubeconfigPath(), "Path to kubeconfig file")
	flag.StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	flag.StringVar(&podLabel, "pod-label", "", "Label selector for pods")
}

func main() {
	flag.Parse()

	// Load in-cluster Kubernetes configuration
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create a Kubernetes clientset using the configuration
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new PodResourceMonitor instance
	podResourceMonitor := monitor.New(clientset, namespace, podLabel)

	// Start the monitoring process
	podResourceMonitor.Run()
}
