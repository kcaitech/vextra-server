#!/bin/bash

set -e

../update-secret.sh jwt-config config.yaml jwt-config.yaml
../update-secret.sh snowflake-config config.yaml snowflake-config.yaml

cp redis-config-template.yaml redis-config.yaml
redis_password=$(kubectl -n bitnami-redis get secret bitnami-redis -o jsonpath="{.data.redis-password}" | base64 -d)
sed -i "s/\$redis_password/$redis_password/g" redis-config.yaml
../update-secret.sh redis-config config.yaml redis-config.yaml

cp mongo-config-template.yaml mongo-config.yaml
mongodb_user=$(kubectl -n psmdb get secrets percona-mongodb-psmdb-secrets -o jsonpath="{.data.MONGODB_DATABASE_ADMIN_USER}" | base64 --decode)
mongodb_password=$(kubectl -n psmdb get secrets percona-mongodb-psmdb-secrets -o jsonpath="{.data.MONGODB_DATABASE_ADMIN_PASSWORD}" | base64 --decode)
sed -i "s/\$mongodb_user/$mongodb_user/g" mongo-config.yaml
sed -i "s/\$mongodb_password/$mongodb_password/g" mongo-config.yaml
../update-secret.sh mongo-config config.yaml mongo-config.yaml

../update-secret.sh storage-config config.yaml storage-config.yaml
