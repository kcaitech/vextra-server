#!/bin/bash

# 安装pxc-mysql

# 安装
helm install pxc-mysql percona/pxc-db -n pxc -f values.yaml
