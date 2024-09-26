#!/bin/bash
docker network create kcserver
docker compose -f minio.yml -f mongodb.yml -f mysql.yml -f redis.yml -f kcserver.yml up -d