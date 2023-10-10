#!/bin/bash

set -e

version_tag=$1
if [ -z "$version_tag" ]; then
  version_tag="latest"
fi

./apply.sh apigateway $version_tag
./apply.sh authservice $version_tag
./apply.sh userservice $version_tag
./apply.sh documentservice $version_tag

cd ./docop-server && ./apply.sh $version_tag
cd -

cd ./webapp && ./apply.sh $version_tag
cd -
