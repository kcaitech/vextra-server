#!/bin/bash

# 初始化k8s资源

set -e

kubectl apply -f mysql-cm.yaml
kubectl apply -f mysql-secert.yaml
kubectl apply -f mysql-pv.yaml
kubectl apply -f mysql-statefulset.yaml
kubectl apply -f mysql-svc.yaml
