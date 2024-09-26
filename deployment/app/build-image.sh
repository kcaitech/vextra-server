#!/bin/bash

SERVICE_NAME=$1
VERSION_TAG=$2

if ["${SERVICE_NAME}" = ""]; then
    echo "参数错误: build-image.sh [SERVICE_NAME] [VERSION_TAG]"
    exit -1
fi
if ["${VERSION_TAG}" = ""]; then
    VERSION_TAG='latest'
fi

# 构建builder镜像：
# docker build --target builder -t kcserver-builder_image:latest -f ../../${SERVICE_NAME}/Dockerfile ../../

docker build -t ${SERVICE_NAME}:${VERSION_TAG} -f ../../${SERVICE_NAME}/Dockerfile ../../
# docker tag ${SERVICE_NAME}:${VERSION_TAG} registry.kcaitech.com:36000/kcserver/${SERVICE_NAME}:${VERSION_TAG}
# docker login registry.kcaitech.com:36000 -u admin -p Kcai1212
# docker push registry.kcaitech.com:36000/kcserver/${SERVICE_NAME}:${VERSION_TAG}