#!/bin/bash

# 下载安装包
echo "下载安装包"
curl https://get.helm.sh/helm-v3.12.3-linux-amd64.tar.gz -LO
# 解压并复制到/usr/local/bin
mkdir helm-v3.12.3-linux-amd64
tar -zxvf helm-v3.12.3-linux-amd64.tar.gz -C helm-v3.12.3-linux-amd64
cp helm-v3.12.3-linux-amd64/linux-amd64/helm /usr/local/bin/helm
# 添加repo
helm repo add aliyun  https://kubernetes.oss-cn-hangzhou.aliyuncs.com/charts
helm repo update
