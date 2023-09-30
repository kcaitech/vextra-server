#!/bin/bash

# 安装bitnami-mariadb-galera

set -e

# 添加仓库
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
# 安装
helm install bitnami-mariadb-galera bitnami/mariadb-galera -n bitnami-mariadb-galera --create-namespace -f values.yaml
