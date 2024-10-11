#!/bin/sh
mc alias set myminio http://localhost:9000 ${MINIO_ROOT_USER} ${MINIO_ROOT_PASSWORD}
mc mb myminio/document
mc mb myminio/files
mc admin policy load myminio http://localhost:9000 --policy-file ./public-read.json
mc policy put myminio/files public #public-read #应该只要公共只读就行
# 创建accesskey
# 创建user
