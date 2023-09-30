#!/bin/bash

set -e

kubectl create ns kc || true
kubectl apply -f docker-registry-auth-apply.yaml
