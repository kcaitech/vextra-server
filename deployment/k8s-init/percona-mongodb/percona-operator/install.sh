#!/bin/bash

# 安装percona-operator

# 添加仓库
helm repo add percona https://percona.github.io/percona-helm-charts/
helm repo update
# 安装
helm install percona-operator percona/psmdb-operator -n psmdb --create-namespace -f values.yaml
