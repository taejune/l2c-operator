#!/bin/bash

cd "$(dirname "$0")"/.. || exit 1
source scripts/common.sh

operator-sdk generate k8s
operator-sdk generate crds

kubectl apply -f "./deploy/crds/tmax.io_${CRD_L2C}_crd.yaml"
kubectl apply -f "./deploy/crds/tmax.io_${CRD_L2CRUN}_crd.yaml"
kubectl apply -f "./deploy/crds/tmax.io_${CRD_SONARQUBE}_crd.yaml"
