#!/bin/bash

# 安装bitnami redis（一主多从哨兵模式）

set -e

# 添加仓库
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
# 安装
helm install bitnami-redis bitnami/redis -n bitnami-redis --create-namespace -f values.yaml
