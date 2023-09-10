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

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/config/interceptor"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

	monitoredDeployments map[string]appsv1.Deployment
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
	logger := log.FromContext(ctx)

	deploymentIdentificator := req.NamespacedName.Name

	// Retrieve the deployment from the cluster info.
	logger.Info("Getting deployment", "namespacedName", req.NamespacedName)
	deployment := appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, &deployment)
	if err != nil {
		logger.Info("Failed to get deployment")

		if _, ok := r.monitoredDeployments[deploymentIdentificator]; ok {
			logger.Info("Deployment was deleted, removing from monitored map.")
			delete(r.monitoredDeployments, deploymentIdentificator)

			// TODO: Delete all Checkpoint/Restore resources.
			logger.Info("Delete Checkpoint/Restore resources.")

			return ctrl.Result{}, nil
		}

		return ctrl.Result{Requeue: false}, err
	}

	// Deployment was created and not yet monitored, attach Interceptor to the Pod and create Checkpoint resource.
	if _, ok := r.monitoredDeployments[deploymentIdentificator]; !ok {
		logger.Info("Deployment is not monitored yet.")

		// We assume every deployment just have 1 replica.
		if deployment.Status.ReadyReplicas == 1 {
			r.monitoredDeployments[deploymentIdentificator] = deployment

			// Verify if the deployment has the annotation to monitor it.
			deploymentMonitoringAnnotation, ok := deployment.Annotations[DEPLOYMENT_MONITORING_ANNOTATION]
			if !ok || deploymentMonitoringAnnotation != "true" {
				// Deployment does not request monitoring.
				logger.Info("Deployment does not requires monitoring.")
				return ctrl.Result{}, nil
			}

			checkpointInterval := "10min"
			deploymentCheckpointInterval, ok := deployment.Annotations[DEPLOYMENT_CHECKPOINT_INTERVAL_ANNOTATION]
			if ok {
				checkpointInterval = deploymentCheckpointInterval
			}

			logger.Info("Deployment checkpoint interval resolved.", "checkpointInterval", checkpointInterval)

			// TODO: Create Checkpoint resource.
			logger.Info("Creating Checkpoint Resource.")

			if err := r.attachInterceptorToPod(deployment, checkpointInterval, logger, ctx); err != nil {
				return ctrl.Result{Requeue: true}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.monitoredDeployments = make(map[string]appsv1.Deployment)
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(r)
}

func (r *DeploymentReconciler) attachInterceptorToPod(deployment appsv1.Deployment, checkpointInterval string, logger logr.Logger, ctx context.Context) error {
	logger.Info("Attaching interceptor to the Pod.")
	var deploymentContainerPort int32 = 8000
	containers := deployment.Spec.Template.Spec.Containers
	if len(containers) > 0 {
		monitoredContainer := containers[0]
		containerPorts := monitoredContainer.Ports
		if len(containerPorts) > 0 {
			containerPort := containerPorts[0].HostPort
			if containerPort == 8001 {
				deployment.Spec.Template.Spec.Containers[0].Ports[0].HostPort = 8000
				containerPort = 8000
			}
			deploymentContainerPort = containerPort
		}
	}

	deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, v1.Container{
		Name:  "interceptor",
		Image: "docker.io/gianaortiz/crsc-interceptor",
		Ports: []v1.ContainerPort{
			{
				ContainerPort: 8001,
				HostPort:      8001,
			},
		},
		Env: []v1.EnvVar{
			{
				Name:  interceptor.CHECKPOINT_INTERVAL_VARIABLE_KEY,
				Value: checkpointInterval,
			},
			{
				Name:  interceptor.CONTAINER_NAME_VARIABLE_KEY,
				Value: deployment.Name,
			},
			{
				Name:  interceptor.CONTAINER_URL_VARIABLE_KEY,
				Value: fmt.Sprintf("http://localhost:%d", deploymentContainerPort),
			},
			{
				Name:  interceptor.STATE_MANAGER_URL_VARIABLE_KEY,
				Value: "http://localhost:5000", // TODO: resolve to correct value.
			},
			{
				Name:  interceptor.ENVIRONMENT_VARIABLE_KEY,
				Value: interceptor.KUBERNETES_ENVIRONMENT,
			},
			{
				Name: interceptor.KUBERNETES_NODE_IP_VARIABLE_KEY,
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: "status.hostIP",
					},
				},
			},
		},
	})
	if err := r.Update(ctx, &deployment); err != nil {
		logger.Error(err, "Failed to attach Interceptor to Pod.")
		return err
	}
	logger.Info("Attached interceptor to the Pod.")
	return nil
}
