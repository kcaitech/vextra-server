#!/bin/bash
VERSION_TAG=$1
if ["${VERSION_TAG}" = ""] (
    VERSION_TAG='latest'
)

./build-image.bat apigateway ${VERSION_TAG}
./build-image.bat authservice ${VERSION_TAG}
./build-image.bat documentservice ${VERSION_TAG}
./build-image.bat userservice ${VERSION_TAG}