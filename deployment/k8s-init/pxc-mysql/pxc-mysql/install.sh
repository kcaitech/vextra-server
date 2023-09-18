#!/bin/bash

# 安装pxc-mysql

set -e

# 安装
helm install pxc-mysql percona/pxc-db -n pxc -f values.yaml
kubectl apply -f pxc-mysql-nodeport-svc.yaml
