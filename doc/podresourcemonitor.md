# podresourcemonitor: monitor package

The `monitor` package provides a utility to monitor resource usage of Kubernetes pods.

## Installation

Install the package using:

```bash
go get -u github.com/sfotiadis/k8s-resource-tracker/pkg/monitor
```

## Usage

Import the package and use the provided APIs to monitor resource usage of Kubernetes pods.

```go
package main

import (
    "log"

    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "path/to/your/monitor"
)

func main() {
    // Create a Kubernetes client configuration
    config, err := rest.InClusterConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Create a Kubernetes clientset
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatal(err)
    }

    // Create a new PodResourceMonitor instance
    podResourceMonitor := monitor.NewPodResourceMonitor(clientset, "default", "app=myapp")

    // Start the monitoring process
    podResourceMonitor.Run()
}
```

## Documentation

### KubeClient Interface

`KubeClient` defines the interface for Kubernetes client operations.

```go
type KubeClient interface {
    GetPod(namespace, name string) (*coreV1.Pod, error)
    ListPods(namespace, labelSelector string) (*coreV1.PodList, error)
}
```

### PodResourceMonitor Type

`PodResourceMonitor` represents a resource monitor for Kubernetes pods.

```go
type PodResourceMonitor struct {
    // Fields
}

func New(clientset kubernetes.Interface, namespace, podLabel string) *PodResourceMonitor {
    // Creates a new PodResourceMonitor instance.
}
```

### CustomResourceEventHandler Type

`CustomResourceEventHandler` is a custom implementation of `cache.ResourceEventHandler`.

```go
type CustomResourceEventHandler struct {
    // Fields
}

func (reh *CustomResourceEventHandler) OnAdd(obj interface{}) {
    // Called when a new pod is added.
}
```

### Run Method

`Run` starts the pod resource monitoring process.

```go
func (prm *PodResourceMonitor) Run() {
    // Starts the monitoring process.
}
```

### GetKubeconfigPath Function

`GetKubeconfigPath` returns the path to the kubeconfig file.

```go
func GetKubeconfigPath() string {
    // Returns the path to the kubeconfig file.
}
```

## Example

See the example provided in the usage section.