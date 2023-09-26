#!/bin/bash

# 安装pxc-mysql

set -e

# 安装
helm -n pxc install pxc-mysql percona/pxc-db -f values.yaml
kubectl apply -f pxc-mysql-nodeport-svc.yaml
