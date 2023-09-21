#!/bin/bash

set -e

../update-secret.sh jwt-config config.yaml jwt-config.yaml
../update-secret.sh snowflake-config config.yaml snowflake-config.yaml
../update-secret.sh storage-config config.yaml storage-config.yaml
../update-secret.sh mongo-config config.yaml mongo-config.yaml
../update-secret.sh redis-config config.yaml redis-config.yaml
