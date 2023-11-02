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
	"time"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/infrastructure/operator/api/v1alpha1"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/config/interceptor"
	"github.com/containers/storage/pkg/reexec"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const DEPLOYMENT_MONITORING_ANNOTATION = "crsc.io/checkpoint-restore"
const DEPLOYMENT_CHECKPOINT_INTERVAL_ANNOTATION = "crsc.io/checkpoint-interval"
const INTERCEPTOR_PORT = 8001

func init() {
	reexec.Init()
}

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

		logger.Info("Getting Pod")
		pod := v1.Pod{}
		err = r.Get(ctx, req.NamespacedName, &pod)
		if err != nil {
			if _, ok := r.monitoredDeployments[deploymentIdentificator]; ok {
				logger.Info("Deployment was deleted, removing from monitored map.")
				delete(r.monitoredDeployments, deploymentIdentificator)

				// TODO: Delete all Checkpoint/Restore resources.
				logger.Info("Delete Checkpoint/Restore resources.")

				return ctrl.Result{}, nil
			}
		}

		logger.Info("Retrieved pod")
		podIP := pod.Status.PodIP
		logger.Info("Pod IP %s", podIP)
		checkpoint := v1alpha1.Checkpoint{}
		checkpointSelector := types.NamespacedName{
			Namespace: "default",
			Name:      "checkpoint-test",
		}
		if err := r.Get(ctx, checkpointSelector, &checkpoint); err == nil {
			logger.Info("Retrieved associated Pod Checkpoint")
			checkpoint.Spec.PodIP = podIP
			logger.Info("Updating Pod Checkpoint with Pod IP")
			if err := r.Update(ctx, &checkpoint); err != nil {
				logger.Error(err, "failed to update checkpoint")
			}
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

			logger.Info("Creating Checkpoint Resource.")
			if err := r.attachInterceptorToPod(deployment, checkpointInterval, logger, ctx); err != nil {
				return ctrl.Result{Requeue: true}, err
			}

			if err := r.createInterceptorService(deployment, logger, ctx); err != nil {
				return ctrl.Result{Requeue: true}, err
			}

			if err := r.createCheckpoint(deployment, checkpointInterval, logger, ctx); err != nil {
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
		Owns(&v1.Pod{}).
		Complete(r)
}

func (r *DeploymentReconciler) createInterceptorService(deployment appsv1.Deployment, logger logr.Logger, ctx context.Context) error {
	logger.Info("Creating Interceptor Service.")
	service := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"crsc.io/name": deployment.Name,
			},
			Ports: []v1.ServicePort{
				{
					Protocol:   v1.ProtocolTCP,
					Port:       INTERCEPTOR_PORT,
					TargetPort: intstr.FromInt(INTERCEPTOR_PORT),
				},
			},
		},
	}
	if err := r.Create(ctx, &service); err != nil {
		return err
	}
	logger.Info("Created Interceptor Service.")
	return nil
}

func (r *DeploymentReconciler) createCheckpoint(deployment appsv1.Deployment, checkpointInterval string, logger logr.Logger, ctx context.Context) error {
	logger.Info("Creating checkpoint cron job.")
	name := fmt.Sprintf("checkpoint-%s", deployment.Name)
	duration, err := time.ParseDuration(checkpointInterval)
	if err != nil {
		return err
	}
	checkpoint := v1alpha1.Checkpoint{
		Spec: v1alpha1.CheckpointSpec{
			Interval:      int(duration.Seconds()),
			ContainerName: deployment.Name,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: deployment.Namespace,
		},
	}
	if err := r.Create(ctx, &checkpoint); err != nil {
		return err
	}
	logger.Info("Created Checkpoint")
	return nil
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
			if containerPort == INTERCEPTOR_PORT {
				deployment.Spec.Template.Spec.Containers[0].Ports[0].HostPort = 8000
				containerPort = 8000
			}
			deploymentContainerPort = containerPort
		}
	}

	deployment.Annotations["crsc.io/name"] = deployment.Name
	deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, v1.Container{
		Name:  "interceptor",
		Image: "docker.io/gianaortiz/crsc-interceptor",
		Ports: []v1.ContainerPort{
			{
				ContainerPort: INTERCEPTOR_PORT,
				HostPort:      INTERCEPTOR_PORT,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "kubelet-certs",
				MountPath: "/var/run/secrets/kubelet-certs",
				ReadOnly:  true,
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
	deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, v1.Volume{
		Name: "kubelet-certs",
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: "kubelet-client-certs",
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
