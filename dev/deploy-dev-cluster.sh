#!/bin/bash

# Starts minikube with cri-o container runtime and a version of kuberenetes with support for
# checkpoint commands.
minikube start --container-runtime="cri-o" --kubernetes-version=v1.28.0 --driver=docker
