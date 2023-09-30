#!/bin/bash

set -e

kubectl label nodes kc-master1 ingress-controller-ready=true
kubectl label nodes kc-master2 ingress-controller-ready=true
kubectl label nodes kc-master3 ingress-controller-ready=true

kubectl apply -f apply.yaml
