#!/bin/bash

# docker pull nginx:1.26.2-alpine

ymls='-f docker-compose.yaml'

if [ "$1" = "up" ] || [ "$1" = "reset" ]; then
    net=$(docker network ls | grep kcserver | awk '{print $2}')
    if [ "$net" != "kcserver" ]; then
        docker network create kcserver
    fi
    docker compose ${ymls} up -d
fi

if [ "$1" = "down" ]; then
    docker compose ${ymls} down
fi
