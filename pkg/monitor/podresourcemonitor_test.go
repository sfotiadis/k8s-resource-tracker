package monitor

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	"log"

	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestCustomResourceEventHandler_OnAdd(t *testing.T) {
	// Create a mock PodResourceMonitor instance for testing
	mockPRM := &PodResourceMonitor{
		clientset: nil,
		Namespace: "default",
		PodLabel:  "app=myapp",
	}

	type fields struct {
		prm *PodResourceMonitor
	}
	type args struct {
		obj interface{}
	}
	tests := []struct {
		args   args
		fields fields
		name   string
	}{
		{
			name: "Matching Pod Label",
			fields: fields{
				prm: mockPRM,
			},
			args: args{
				obj: &coreV1.Pod{
					ObjectMeta: metaV1.ObjectMeta{
						Name:   "test-pod",
						Labels: map[string]string{"app": "myapp"},
					},
				},
			},
		},
		{
			name: "Non-Matching Pod Label",
			fields: fields{
				prm: mockPRM,
			},
			args: args{
				obj: &coreV1.Pod{
					ObjectMeta: metaV1.ObjectMeta{
						Name:   "test-pod",
						Labels: map[string]string{"app": "otherapp"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reh := &CustomResourceEventHandler{
				prm: tt.fields.prm,
			}
			reh.OnAdd(tt.args.obj)
		})
	}
}

func TestPodResourceMonitor_Run(t *testing.T) {
	// Create a fake clientset for testing
	clientset := fake.NewSimpleClientset()

	const (
		phaseRunning    = coreV1.PodRunning
		phaseNotRunning = coreV1.PodFailed
	)
	// Configure the expected list of pods
	expectedPods := &coreV1.PodList{
		Items: []coreV1.Pod{
			{
				ObjectMeta: metaV1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
				},
				Spec: coreV1.PodSpec{
					// Configure the pod spec as needed
					Containers: []coreV1.Container{
						{
							Name:  "container-1",
							Image: "nginx:latest",
						},
					},
				},
				Status: coreV1.PodStatus{
					Phase: coreV1.PodRunning, // Set the initial phase to "Running"
				},
			},
			{
				ObjectMeta: metaV1.ObjectMeta{
					Name:      "pod-2",
					Namespace: "default",
				},
				Spec: coreV1.PodSpec{
					// Configure the pod spec for the second pod
					Containers: []coreV1.Container{
						{
							Name:  "container-2",
							Image: "busybox:latest",
						},
					},
				},
				Status: coreV1.PodStatus{
					Phase: coreV1.PodRunning, // Set the initial phase to "Running"
				},
			},
		},
	}

	// Prepend the reactor to the fake clientset
	clientset.Fake.PrependReactor("list", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, expectedPods, nil
	})

	type fields struct {
		clientset *fake.Clientset
		Namespace string
		PodLabel  string
	}
	type args struct {
		expectedPhase coreV1.PodPhase
	}
	tests := []struct {
		args    args
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Run PodResourceMonitor with Running Pods",
			args: args{
				expectedPhase: phaseRunning,
			},
			fields: fields{
				clientset: clientset,
				Namespace: "default",
				PodLabel:  "app=myapp",
			},
			wantErr: false,
		},
		{
			name: "Run PodResourceMonitor with Failed Pods",
			args: args{
				expectedPhase: phaseNotRunning,
			},
			fields: fields{
				clientset: clientset,
				Namespace: "default",
				PodLabel:  "app=myapp",
			},
			wantErr: true,
		},
	}

	timeout := 10 * time.Second // Adjust the timeout duration as needed

	// Run the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prm := New(tt.fields.clientset, tt.fields.Namespace, tt.fields.PodLabel)

			// Create a channel to signal the informer has started and completed
			startedCh := make(chan struct{})
			completedCh := make(chan struct{})

			// Run the informer in a goroutine
			go func() {
				defer close(completedCh)
				// Set up a stop channel for the informer
				prm.stopCh = make(chan struct{})
				defer close(prm.stopCh)

				// Signal the informer has started
				close(startedCh)

				// Start the informer
				prm.Run()

				// Wait for the informer to start
				<-startedCh

				// Simulate stopping the informer (you can replace this with the actual stopping mechanism)
				close(prm.stopCh)

				// Wait for the informer to complete
				<-prm.stopCh

				// Wait for the informer to complete
				<-completedCh
			}()

			// Wait for the informer to start
			select {
			case <-startedCh:
				// Informer started, proceed
			case <-time.After(timeout):
				t.Fatalf("Timeout waiting for informer to start")
			}

			// For example, you can assert that the list of expected pods was used
			actualPods, err := prm.clientset.CoreV1().Pods(tt.fields.Namespace).List(context.TODO(), metaV1.ListOptions{})
			if err != nil {
				log.Println(err)
			}

			if tt.wantErr {
				for _, actualPod := range actualPods.Items {
					assert.NotEqual(t, tt.args.expectedPhase, actualPod.Status.Phase, "Pod is not in running state")
				}
			} else {
				assert.Equal(t, expectedPods.Items, actualPods.Items, "Expected pods do not match actual pods")
				assert.Equal(t, len(expectedPods.Items), len(actualPods.Items), "Number of started pods does not match")
				for i, expectedPod := range expectedPods.Items {
					actualPod := actualPods.Items[i]
					assert.Equal(t, expectedPod.Spec.Containers[0].Image, actualPod.Spec.Containers[0].Image, "Container image does not match")
				}

				for i, expectedPod := range expectedPods.Items {
					actualPod := actualPods.Items[i]
					assert.Equal(t, expectedPod.ObjectMeta.Labels, actualPod.ObjectMeta.Labels, "Pod labels do not match")
				}

				for _, actualPod := range actualPods.Items {
					assert.Equal(t, tt.args.expectedPhase, actualPod.Status.Phase, "Pod is not in running state")
				}
			}

			// Signal the informer has completed
			close(completedCh)
		})
	}
}
