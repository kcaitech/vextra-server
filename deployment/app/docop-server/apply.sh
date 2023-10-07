#!/bin/bash

set -e

kubectl create ns kc || true
kubectl apply -f s3-secret.yaml

cp template.env .env
mysql_passwd=$(kubectl -n bitnami-mariadb-galera get secret bitnami-mariadb-galera -o jsonpath="{.data.mariadb-root-password}" | base64 -d)
mongodb_user=$(kubectl -n psmdb get secrets percona-mongodb-psmdb-secrets -o jsonpath="{.data.MONGODB_DATABASE_ADMIN_USER}" | base64 -d)
mongodb_passwd=$(kubectl -n psmdb get secrets percona-mongodb-psmdb-secrets -o jsonpath="{.data.MONGODB_DATABASE_ADMIN_PASSWORD}" | base64 -d)
s3_access_key_id=$(kubectl -n kc get secrets s3-secret -o jsonpath="{.data.S3_ACCESS_KEY_ID}" | base64 -d)
s3_secret_access_key=$(kubectl -n kc get secrets s3-secret -o jsonpath="{.data.S3_SECRET_ACCESS_KEY}" | base64 -d)
sed -i "s/\$mysql_passwd/$mysql_passwd/g" .env
sed -i "s/\$mongodb_user/$mongodb_user/g" .env
sed -i "s/\$mongodb_passwd/$mongodb_passwd/g" .env
sed -i "s/\$s3_access_key_id/$s3_access_key_id/g" .env
sed -i "s/\$s3_secret_access_key/$s3_secret_access_key/g" .env
../update-secret.sh docop-server-config .env .env

cd ../config && ./create-secret.sh
cd -

kubectl apply -f apply.yaml
