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
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const DEPLOYMENT_MONITORING_ANNOTATION = "crsc.io/checkpoint-restore"
const DEPLOYMENT_CHECKPOINT_INTERVAL_ANNOTATION = "crsc.io/checkpoint-interval"

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	monitoredDeployments map[types.UID]appsv1.Deployment
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Retrieve the deployment from the cluster info.
	deployment := appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, &deployment)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Deployment was created and not yet monitored, attach Interceptor to the Pod and create Checkpoint resource.
	if _, ok := r.monitoredDeployments[deployment.UID]; !ok {
		r.monitoredDeployments[deployment.UID] = deployment

		// Verify if the deployment has the annotation to monitor it.
		deploymentMonitoringAnnotation, ok := deployment.Annotations[DEPLOYMENT_MONITORING_ANNOTATION]
		if !ok || deploymentMonitoringAnnotation != "true" {
			// Deployment does not request monitoring.
			return ctrl.Result{}, nil
		}

		checkpointInterval := time.Duration(10 * time.Minute)
		deploymentCheckpointInterval, ok := deployment.Annotations[DEPLOYMENT_CHECKPOINT_INTERVAL_ANNOTATION]
		if ok {
			deploymentCheckpointIntervalValue, err := strconv.Atoi(deploymentCheckpointInterval)
			if err == nil {
				checkpointInterval = time.Duration(time.Minute * time.Duration(deploymentCheckpointIntervalValue*1000000000))
			}
		}
		fmt.Println(checkpointInterval)

		// TODO: Create Checkpoint resource.

		// TODO: Attach Interceptor to the Pod.
	}

	if deployment.DeletionTimestamp != nil {
		delete(r.monitoredDeployments, deployment.UID)

		// TODO: Delete all Checkpoint/Restore resources.

		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(r)
}
