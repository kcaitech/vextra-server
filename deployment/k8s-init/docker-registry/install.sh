#!/bin/bash

# 创建认证证书

set -e

docker run \
  --entrypoint htpasswd \
  httpd:2 -Bbn kcai kcai1212 > docker-registry-auth-secret
# 创建namespace
kubectl create namespace docker-registry
# 创建secret
kubectl -n docker-registry create secret generic docker-registry-auth-secret \
  --from-file=htpasswd=docker-registry-auth-secret

# 安装
kubectl apply -f deployment.yaml
