#!/bin/bash
version=$(cat package.yaml | grep 'version:' | sed 's/[^0-9.]*//g')
echo version: $version
echo '---'
docker build  -t kcserver:$version  -f Dockerfile .
docker build  -t kcserver-inner:$version  -f Dockerfile-inner .