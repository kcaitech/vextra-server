#!/bin/bash

set -e

./apply.sh apigateway
./apply.sh authservice
./apply.sh userservice
./apply.sh documentservice

cd ./docop-server && ./apply.sh
cd -
