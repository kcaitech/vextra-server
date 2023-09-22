#!/bin/bash

set -e

# 验证参数数量
if [[ $# -ne 1 ]]; then
    echo "参数错误：$0 <service名称>"
    exit 1
fi

SERVICE_NAME=$1

kubectl create ns kc || true

cd "$SERVICE_NAME"
cp config-template.yaml config.yaml
mysql_passwd=$(kubectl -n pxc get secrets pxc-mysql-pxc-db-secrets -o jsonpath="{.data.root}" | base64 --decode)
sed -i "s/\$mysql_passwd/$mysql_passwd/g" config.yaml
../update-secret.sh "$SERVICE_NAME-config" config.yaml config.yaml
cd ../config && ./create-secret.sh
cd -

kubectl apply -f apply.yaml
