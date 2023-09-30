#!/bin/bash

set -e

kubectl delete -f http-ingress.yaml
kubectl delete -f http-svc.yaml
kubectl delete -f http-dep.yaml
