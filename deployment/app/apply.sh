#!/bin/bash

set -e

# 验证参数数量，必须大于等于1个
if [ $# -lt 1 ]; then
    echo "参数错误：$0 <service名称>"
    exit 1
fi

SERVICE_NAME=$1

kubectl create ns kc || true

cd "$SERVICE_NAME"
cp config-template.yaml config.yaml
mysql_passwd=$(kubectl -n bitnami-mariadb-galera get secret bitnami-mariadb-galera -o jsonpath="{.data.mariadb-root-password}" | base64 -d)
sed -i "s/\$mysql_passwd/$mysql_passwd/g" config.yaml
../update-secret.sh "$SERVICE_NAME-config" config.yaml config.yaml
cd ../config && ./create-secret.sh
cd -

version_tag=$2
if [ -z "$version_tag" ]; then
  version_tag="latest"
fi
cp apply.template.yaml apply.yaml
sed -i "s/\$version_tag/$version_tag/g" apply.yaml

kubectl apply -f apply.yaml
