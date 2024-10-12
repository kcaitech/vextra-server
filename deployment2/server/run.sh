#!/bin/bash

ymls='-f kcsvg2png.yaml -f kcversion.yaml -f kcserver.yaml'

if [ "$1" = "up" ] || [ "$1" = "reset" ]; then
    net=$(docker network ls | grep kcserver | awk '{print $2}')
    if [ "$net" != "kcserver" ]; then
        docker network create kcserver
    fi

    # 清除掉log
    rm -rf kcserver/log
    rm -rf kcversion/log
    rm -rf kcsvg2png/log

    docker compose ${ymls} up -d
fi

if [ "$1" = "down" ]; then
    docker compose ${ymls} down
fi
