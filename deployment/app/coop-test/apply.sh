#!/bin/bash

set -e

kubectl create ns kc || true

version_tag=$1
if [ -z "$version_tag" ]; then
  version_tag="latest"
fi
cp apply.template.yaml apply.yaml
sed -i "s/\$version_tag/$version_tag/g" apply.yaml

kubectl apply -f apply.yaml
