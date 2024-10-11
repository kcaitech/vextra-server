#!/bin/bash

ymls='-f minio.yaml -f mongodb.yaml -f mysql.yaml -f redis.yaml'

cmd=$1
if [ "$cmd" = "down" ] || [ "$cmd" = "reset" ]; then
    docker compose ${ymls} down
fi

if [ "$cmd" = "reset" ]; then
    rm -rf minio/data
    rm -rf mongodb/data
    rm -rf mysql/data
    rm -rf redis/data
    mkdir minio/data
    mkdir mongodb/data
    mkdir mysql/data
    mkdir redis/data
fi

if [ "$cmd" = "up" ] || [ "$cmd" = "reset" ]; then
    net=$(docker network ls | grep kcserver | awk '{print $2}')
    if [ "$net" != "kcserver" ]; then
        docker network create kcserver
    fi
    docker compose ${ymls} up -d
fi
