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
	"fmt"
	"net/http"
	"time"

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

	PodsToRestore map[string]podToRestore
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	pod := corev1.Pod{}
	err := r.Get(ctx, req.NamespacedName, &pod)
	if err != nil {
		logger.Info("no pod found")
		return ctrl.Result{}, nil
	}

	logger.Info("Got Pod")

	monitoringAnnotation, ok := pod.Annotations[DEPLOYMENT_MONITORING_ANNOTATION]
	if !ok || monitoringAnnotation != "true" {
		logger.Info("Pod does not requires monitoring.")
		return ctrl.Result{}, nil
	}

	podIsAlreadyStoredForRestore := false
	for key, podToRestore := range r.PodsToRestore {
		if podToRestore.pod.Name == pod.Name {
			now := time.Now()
			logger.Info("Pod is scheduled to restore, restoring it...", "time", now.String())
			// Check for the container status to be running.
			for _, container := range pod.Status.ContainerStatuses {
				if container.Name == podToRestore.containerName {
					if container.State.Running != nil {
						podIsAlreadyStoredForRestore = true
						// Container already started.
						podIP := pod.Status.PodIP

						// Call interceptor to reproject all request.
						reprojectURL := fmt.Sprintf("http://%s:8001/reproject", podIP)
						res, err := http.Get(reprojectURL)
						if err != nil {
							logger.Error(err, "failed to reproject requests")
							continue
						}

						if res.StatusCode != http.StatusOK {
							logger.Error(fmt.Errorf("reproject sent status %d", res.StatusCode), "failed to reproject requests")
							continue
						}

						// Change interceptor state to serving.
						changeStateURL := fmt.Sprintf("http://%s:8001/state?state=Proxying", podIP)
						res, err = http.Post(changeStateURL, "application/x-www-form-urlencoded", nil)
						if err != nil {
							logger.Error(err, "failed to change state")
							continue
						}

						if res.StatusCode != http.StatusOK {
							logger.Error(fmt.Errorf("change state sent status %d", res.StatusCode), "failed to change state")
							continue
						}

						now = time.Now()
						logger.Info("Pod restored", "time", now.String())
						delete(r.PodsToRestore, key)
						break
					}
				}
				logger.Info("Pod restored")
			}
		}
	}

	if !podIsAlreadyStoredForRestore {
		logger.Info("Checking if pod has failed container")
		// Add pod to restore if there is a fail container.
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Terminated != nil {
				now := time.Now()
				logger.Info("Adding Pod to pods to restore", "time", now.String())
				// Container crashed, it must be restored.
				r.PodsToRestore[pod.Name] = podToRestore{
					pod:           pod,
					containerName: containerStatus.Name,
				}
				// Set the interceptor to wait status, so it cache all requests.
				podIP := pod.Status.PodIP
				changeStateURL := fmt.Sprintf("http://%s:8001/state?state=Caching", podIP)
				res, err := http.Post(changeStateURL, "application/x-www-form-urlencoded", nil)
				if err != nil {
					logger.Error(err, "failed to change state")
					continue
				}

				if res.StatusCode != http.StatusOK {
					logger.Error(fmt.Errorf("change state sent status %d", res.StatusCode), "failed to change state")
					continue
				}
				logger.Info("Added Pod to pods to restore")
				break
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.PodsToRestore = make(map[string]podToRestore)
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}
