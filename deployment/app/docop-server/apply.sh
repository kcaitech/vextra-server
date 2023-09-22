#!/bin/bash

set -e

kubectl create ns kc || true

cp template.env .env
mysql_passwd=$(kubectl -n pxc get secrets pxc-mysql-pxc-db-secrets -o jsonpath="{.data.root}" | base64 --decode)
mongodb_user=$(kubectl -n psmdb get secrets percona-mongodb-psmdb-secrets -o jsonpath="{.data.MONGODB_DATABASE_ADMIN_USER}" | base64 --decode)
mongodb_passwd=$(kubectl -n psmdb get secrets percona-mongodb-psmdb-secrets -o jsonpath="{.data.MONGODB_DATABASE_ADMIN_PASSWORD}" | base64 --decode)
sed -i "s/\$mysql_passwd/$mysql_passwd/g" .env
sed -i "s/\$mongodb_user/$mongodb_user/g" .env
sed -i "s/\$mongodb_passwd/$mongodb_passwd/g" .env
../update-secret.sh docop-server-config .env .env

cd ../config && ./create-secret.sh
cd -

kubectl apply -f apply.yaml
