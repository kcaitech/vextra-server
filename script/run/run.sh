#!/bin/bash

# 启动服务
# 参数是服务名：mysql minio apigateway authservice documentservice userservice
# 输入参数一次可多个，无输入参数时全部启动

if [ $(basename "$PWD") != "kcserver" ]; then
  echo "请在kcserver目录下执行"
fi

originalPath=$PWD
baseDir=$PWD

function FindIndex() {
  local value=$1
  shift
  local array=("$@")
  for i in "${!array[@]}"; do
    if [ "${array[$i]}" = "$value" ]; then
      echo $i
      return
    fi
  done
  echo -1
}

function RunService() {
  local serviceList1=("mysql" "minio")
  local position=$(FindIndex "$1" "${serviceList1[@]}")
  if [ $position -ne -1 ]; then
    cd "$baseDir/docker_compose/$1"
    docker-compose up -d
    return
  fi

  local serviceList2=("apigateway" "authservice" "documentservice" "userservice")
  local position=$(FindIndex "$1" "${serviceList2[@]}")
  if [ $position -ne -1 ]; then
    cd "$baseDir/$1"
    docker-compose up -d --build --force-recreate
    return
  fi

  echo "服务名称错误：$1"
  exit 1
}

if [ $# -eq 0 ] || [ $1 = "" ]; then
  # "mysql" "minio"
  serviceList=("apigateway" "authservice" "documentservice" "userservice")
  for service in "${serviceList[@]}"; do
    RunService "$service"
  done
else
  for service in "$@"; do
    RunService "$service"
  done
fi
