#!/bin/bash

MINIKUBE=/tmp/minikube

K="kubectl --context=minikube"

wget -O $MINIKUBE https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 
chmod +x $MINIKUBE

$MINIKUBE start --bootstrapper=kubeadm

$K create serviceaccount --namespace kube-system tiller
$K create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller

helm init --service-account=$HELM_SA


