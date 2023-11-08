# (crsc): Checkpoint/Restore in Kubernetes for Stateful Containers

crsc is a transparent checkpoint/restore tooling for stateful container in a Kubernetes cluster. It works by implementing a custom operator that manages the periodically checkpointing of an application and restoration after a fault.

## Kubernetes Operator

The Kubernetes Operator is composed of many controllers that each have a defined functionality.

### Deployment Controller

The Deployment Controller is responsible for monitoring new Deployment resources and changes to the resources. It must check if the Deployment has the `crsc.io/checkpoint-restore` annotation enabled before continuing to monitor changes to the Deployment.

As soon a Deployment is created we wait for the creation of the Pod of the deployment, we are assuming every Deployment runs one replicas and every Pod has only one image running. When the Pod is created we can add a sidecar container with our Interceptor, which will intercept requests to the Pod and send them to the monitored container later. When we add the sidecar we add a new annotation to the pod, `crsc.io/interceptor-sidecar-attached`, to indicate that we attached the sidecar Interceptor to the Pod and add the Pod information to a database so we can fetch it later. As we create the Interceptor we also create a new Checkpoint resource to checkpoint the application with the given `crsc.io/checkpoint-interval` annotation between checkpoints.

When we check that a new Deployment is created with the monitored annotation we add it to the database so our Checkpoint/Restore tooling can all see the current monitoring resources.

When the Deployment is deleted we must clear the Checkpoint resources.

### Pod Controller

When a Pod reachs a state of fault, like the health probe of the monitored container is failing, and the Pod is already attached by a Interceptor sidecar with the `crsc.io/interceptor-sidecar-attached` annotation we must issue a new Restore resource.

### Checkpoint Controller

The Checkpoint Controller check for changes in Checkpoints, when a new Checkpoint is created it must create a new CronJob for run in the given interval requesting the Interceptor to perform a checkpoint of the application.

### Restore Controller

When a Restore resource is added we must request for the State Manager to restore the given application. The State Manager will change the deployment to use the latest checkpoint of the application and restart the application.


