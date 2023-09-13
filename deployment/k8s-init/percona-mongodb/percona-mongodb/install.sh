#!/bin/bash

# 安装percona-mongodb

# 安装
helm install percona-mongodb percona/psmdb-db -n psmdb -f values.yaml
kubectl apply -f mongos-nodeport-svc.yaml
