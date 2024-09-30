#!/bin/bash

ymls='-f minio.yaml -f mongodb.yaml -f mysql.yaml -f redis.yaml -f kcsvg2png.yaml -f kcversion.yaml -f kcserver.yaml'

if [ "$1" = "up" ]; then
    docker network create kcserver
    docker compose ${ymls} up -d
fi

if [ "$1" = "down" ]; then
    docker compose ${ymls} down
fi
