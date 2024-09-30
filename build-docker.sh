#!/bin/bash
version=$(cat package.yaml | grep 'version:' | sed 's/[^0-9.]*//g')
echo version: $version
echo '---'
docker pull golang:1.22-alpine
docker pull alpine:3.17
docker build  -t kcserver:$version  -f Dockerfile .