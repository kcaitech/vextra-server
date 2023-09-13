#!/bin/bash

# 安装bitnami-mysql

# 添加仓库
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
# 安装
helm install bitnami-mysql bitnami/mysql -n bitnami-mysql --create-namespace -f values.yaml
