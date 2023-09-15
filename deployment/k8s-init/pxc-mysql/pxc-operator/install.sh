#!/bin/bash

# 安装pxc-operator

set -e

# 添加仓库
helm repo add percona https://percona.github.io/percona-helm-charts/
helm repo update
# 安装
helm install pxc-operator percona/pxc-operator -n pxc --create-namespace -f values.yaml
