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

# 安装docker-registry
kubectl apply -f docker-registry-apply.yaml

# 安装docker-registry-ui
docker_registry_ip=$(nslookup docker-registry.protodesign.cn | grep Address | tail -1 | awk '{print $2}')
echo "docker_registry_ip: $docker_registry_ip"
cp docker-registry-ui-apply-template.yaml docker-registry-ui-apply.yaml
sed -i "s/\$docker_registry_ip/$docker_registry_ip/g" docker-registry-ui-apply.yaml
kubectl apply -f docker-registry-ui-apply.yaml
