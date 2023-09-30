#!/bin/bash

set -e

kubectl apply -f http-dep.yaml
kubectl apply -f http-svc.yaml
kubectl apply -f http-ingress.yaml
