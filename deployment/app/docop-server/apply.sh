#!/bin/bash

set -e

kubectl create ns kc || true

s3_access_key_id=$(grep 'S3_ACCESS_KEY_ID' passwd | awk -F= '{print $2}')
s3_secret_access_key=$(grep 'S3_SECRET_ACCESS_KEY' passwd | awk -F= '{print $2}')
sed -i "s/\$s3_access_key_id/$s3_access_key_id/g" s3-secret.yaml
sed -i "s/\$s3_secret_access_key/$s3_secret_access_key/g" s3-secret.yaml
kubectl apply -f s3-secret.yaml

cp template.env .env
mysql_passwd=$(kubectl -n bitnami-mariadb-galera get secret bitnami-mariadb-galera -o jsonpath="{.data.mariadb-root-password}" | base64 -d)
mongodb_user=$(kubectl -n psmdb get secrets percona-mongodb-psmdb-secrets -o jsonpath="{.data.MONGODB_DATABASE_ADMIN_USER}" | base64 -d)
mongodb_passwd=$(kubectl -n psmdb get secrets percona-mongodb-psmdb-secrets -o jsonpath="{.data.MONGODB_DATABASE_ADMIN_PASSWORD}" | base64 -d)
redis_passwd=$(kubectl -n bitnami-redis get secret bitnami-redis -o jsonpath="{.data.redis-password}" | base64 -d)
sed -i "s/\$mysql_passwd/$mysql_passwd/g" .env
sed -i "s/\$mongodb_user/$mongodb_user/g" .env
sed -i "s/\$mongodb_passwd/$mongodb_passwd/g" .env
sed -i "s/\$s3_access_key_id/$s3_access_key_id/g" .env
sed -i "s/\$s3_secret_access_key/$s3_secret_access_key/g" .env
sed -i "s/\$redis_passwd/$redis_passwd/g" .env
../update-secret.sh docop-server-config .env .env

cd ../config && ./create-secret.sh
cd -

version_tag=$1
if [ -z "$version_tag" ]; then
  version_tag="latest"
fi
cp apply-statefulset.template.yaml apply.yaml
sed -i "s/\$version_tag/$version_tag/g" apply.yaml

kubectl apply -f apply.yaml
