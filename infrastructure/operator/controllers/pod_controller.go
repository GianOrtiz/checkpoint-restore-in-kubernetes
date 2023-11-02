/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type podToRestore struct {
	pod           corev1.Pod
	containerName string
}

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	PodsToRestore []podToRestore
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	pod := corev1.Pod{}
	err := r.Get(ctx, req.NamespacedName, &pod)
	if err != nil {

	}

	monitoringAnnotation, ok := pod.Annotations[DEPLOYMENT_MONITORING_ANNOTATION]
	if !ok || monitoringAnnotation != "true" {
		logger.Info("Pod does not requires monitoring.")
		return ctrl.Result{}, nil
	}

	podIsAlreadyStoredForRestore := false
	for _, podToRestore := range r.PodsToRestore {
		if podToRestore.pod.Name == pod.Name {
			// Check for the container status to be running.
			for _, container := range pod.Status.ContainerStatuses {
				if container.Name == podToRestore.containerName {
					if container.State.Running != nil {
						// Container already started.

						// Call interceptor to reproject all request.

						// Change interceptor state to serving.
					}
				}
			}
		}
	}

	if !podIsAlreadyStoredForRestore {
		// Add pod to restore if there is a fail container.
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Terminated != nil {
				// Container crashed, it must be restored.
				r.PodsToRestore = append(r.PodsToRestore, podToRestore{
					pod:           pod,
					containerName: containerStatus.Name,
				})
				// Set the interceptor to wait status, so it cache all requests.
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}
