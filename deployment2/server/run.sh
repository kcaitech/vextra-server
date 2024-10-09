#!/bin/bash

ymls='-f kcsvg2png.yaml -f kcversion.yaml -f kcserver.yaml'

if [ "$1" = "up" ]; then
    net=$(docker network ls | grep kcserver | awk '{print $2}')
    if [ "$net" != "kcserver" ]; then
        docker network create kcserver
    fi
    docker compose ${ymls} up -d
fi

if [ "$1" = "down" ]; then
    docker compose ${ymls} down
fi
