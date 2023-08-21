// Package monitor provides a utility to monitor resource usage of Kubernetes pods.
package monitor

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"log"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/homedir"
)

// KubeClient defines the interface for Kubernetes client operations.
type KubeClient interface {
	GetPod(namespace, name string) (*coreV1.Pod, error)
	ListPods(namespace, labelSelector string) (*coreV1.PodList, error)
	// Add more methods as needed for your operations
}

// PodResourceMonitor represents a resource monitor for Kubernetes pods.
type PodResourceMonitor struct {
	clientset         kubernetes.Interface
	stopCh            chan struct{}
	cpuUsageMetric    *prometheus.GaugeVec
	memoryUsageMetric *prometheus.GaugeVec
	Namespace         string
	PodLabel          string
}

// New creates a new PodResourceMonitor instance.
func New(clientset kubernetes.Interface, namespace, podLabel string) *PodResourceMonitor {
	cpuUsageMetric := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pod_cpu_usage",
			Help: "CPU usage of pods",
		},
		[]string{"namespace", "pod_name", "container_name"},
	)

	memoryUsageMetric := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pod_memory_usage",
			Help: "Memory usage of pods",
		},
		[]string{"namespace", "pod_name", "container_name"},
	)

	return &PodResourceMonitor{
		clientset:         clientset,
		Namespace:         namespace,
		PodLabel:          podLabel,
		cpuUsageMetric:    cpuUsageMetric,
		memoryUsageMetric: memoryUsageMetric,
	}
}

// CustomResourceEventHandler is a custom implementation of cache.ResourceEventHandler.
type CustomResourceEventHandler struct {
	prm *PodResourceMonitor
}

// OnAdd is called when a new pod is added.
func (reh *CustomResourceEventHandler) OnAdd(obj interface{}) {
	pod := obj.(*coreV1.Pod)
	if reh.prm.PodLabel == "" || pod.Labels[reh.prm.PodLabel] != "" {
		go reh.prm.monitorPodResourceUsage(pod)
	}
}

// Run starts the pod resource monitoring process.
func (prm *PodResourceMonitor) Run() {
	// Create a ListWatcher for pods in the specified namespace
	listWatcher := &cache.ListWatch{
		ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
			return prm.clientset.CoreV1().Pods(prm.Namespace).List(context.TODO(), metaV1.ListOptions{})
		},
		WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
			return prm.clientset.CoreV1().Pods(prm.Namespace).Watch(context.Background(), options)
		},
	}

	// Create an informer to watch for pod changes
	informer := cache.NewSharedInformer(listWatcher, &coreV1.Pod{}, 0)

	// Set up a stop channel for the informer
	prm.stopCh = make(chan struct{}) // Initialize the stop channel
	defer close(prm.stopCh)

	// Start the informer
	go informer.Run(prm.stopCh) // Use prm.stopCh to stop the informer

	// Wait for the informer to stop
	<-prm.stopCh
}

// monitorPodResourceUsage monitors the resource usage of a pod.
func (prm *PodResourceMonitor) monitorPodResourceUsage(pod *coreV1.Pod) {
	// Create a ticker to monitor resource usage at regular intervals
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Get the latest pod information from the API server
			pod, err := prm.clientset.CoreV1().Pods(prm.Namespace).Get(context.TODO(), pod.Name, metaV1.GetOptions{})
			if err != nil {
				log.Println(err)
				continue
			}

			// Print resource usage information for each container in the pod
			for _, container := range pod.Spec.Containers {
				cpuUsage, _ := container.Resources.Requests.Cpu().AsInt64()
				memory, _ := container.Resources.Requests.Memory().AsInt64()
				prm.cpuUsageMetric.WithLabelValues(prm.Namespace, pod.Name, container.Name).Set(float64(cpuUsage))
				prm.memoryUsageMetric.WithLabelValues(prm.Namespace, pod.Name, container.Name).Set(float64(memory))
				log.Printf("Pod: %s, Container: %s, CPU: %s, Memory: %s\n", pod.Name, container.Name, container.Resources.Requests.Cpu(), container.Resources.Requests.Memory())
			}
		}
	}
}

// GetKubeconfigPath returns the path to the kubeconfig file.
func GetKubeconfigPath() string {
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}
	return ""
}

// Example usage:
//
// package main
//
// import (
//     "log"
//
//     "k8s.io/client-go/kubernetes"
//     "k8s.io/client-go/rest"
//     "path/to/your/monitor"
// )
//
// func main() {
//     // Create a Kubernetes client configuration
//     config, err := rest.InClusterConfig()
//     if err != nil {
//         log.Fatal(err)
//     }
//
//     // Create a Kubernetes clientset
//     clientset, err := kubernetes.NewForConfig(config)
//     if err != nil {
//         log.Fatal(err)
//     }
//
//     // Create a new PodResourceMonitor instance
//     podResourceMonitor := monitor.NewPodResourceMonitor(clientset, "default", "app=myapp")
//
//     // Start the monitoring process
//     podResourceMonitor.Run()
// }
