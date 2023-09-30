#!/bin/bash

set -e

kubectl create ns kc || true

cp template.env .env
mysql_passwd=$(kubectl -n bitnami-mariadb-galera get secret bitnami-mariadb-galera -o jsonpath="{.data.mariadb-root-password}" | base64 -d)
mongodb_user=$(kubectl -n psmdb get secrets percona-mongodb-psmdb-secrets -o jsonpath="{.data.MONGODB_DATABASE_ADMIN_USER}" | base64 -d)
mongodb_passwd=$(kubectl -n psmdb get secrets percona-mongodb-psmdb-secrets -o jsonpath="{.data.MONGODB_DATABASE_ADMIN_PASSWORD}" | base64 -d)
sed -i "s/\$mysql_passwd/$mysql_passwd/g" .env
sed -i "s/\$mongodb_user/$mongodb_user/g" .env
sed -i "s/\$mongodb_passwd/$mongodb_passwd/g" .env
../update-secret.sh docop-server-config .env .env

cd ../config && ./create-secret.sh
cd -

kubectl apply -f apply.yaml
