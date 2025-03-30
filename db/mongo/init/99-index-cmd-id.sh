#!/bin/sh
# 设置环境变量
MONGO_URI="mongodb://${MONGO_INITDB_ROOT_USERNAME}:${MONGO_INITDB_ROOT_PASSWORD}@localhost:27017"
# 连接到MongoDB并执行初始化命令
mongosh --eval "
    db = db.getSiblingDB('${MONGO_INITDB_DATABASE}');
    db.${MONGO_INITDB_COLLECTION}.createIndex({ document_id: 1, cmd_id: 1 }, { unique: true });
    print('Index created successfully.');
" $MONGO_URI