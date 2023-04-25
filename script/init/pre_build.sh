#!/bin/bash
# 首次构建之前执行

if [ $(basename "$PWD") != "kcserver" ]; then
  echo "请在kcserver目录下执行"
fi

originalPath=$PWD
baseDir=$PWD

# 创建空日志文件
function CreateLog() {
  cd "$baseDir/$1"
  if [ ! -e "log/all.log" ]; then
    mkdir -p log
    touch log/all.log
  fi
}
CreateLog apigateway
CreateLog authservice
CreateLog documentservice
CreateLog userservice

# 创建docker network
docker network create --subnet=172.21.0.0/16 --gateway=172.21.0.1 db_net_1

cd $originalPath
