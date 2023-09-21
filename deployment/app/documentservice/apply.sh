#!/bin/bash

set -e

kubectl create ns kc || true

../update-secret.sh documentservice-config config.yaml documentservice-config.yaml
cd ../config && ./create-secret.sh
cd -

kubectl apply -f apply.yaml
