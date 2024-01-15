#!/bin/bash

set -e

kubectl create ns kc || true
kubectl apply -f service-account-apply.yaml
