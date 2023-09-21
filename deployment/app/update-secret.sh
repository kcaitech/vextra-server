#!/bin/bash

# 创建/更新secret
# 接收3个输入参数：secret名称、secret中的字段名、config文件名

# set -e

# 验证参数数量
if [[ $# -ne 3 ]]; then
    echo "参数错误：$0 <secret名称> <secret中的字段名> <config文件名>"
    exit 1
fi

SECRET_NAME=$1
DATA_FIELD_NAME=$2
FILE_PATH=$3

# 检测 secret 是否存在
kubectl -n kc get secret $SECRET_NAME &> /dev/null
if [[ $? -eq 0 ]]; then
    # 存在
    # 文件内容Base64编码
    ENCODED_VALUE=$(cat $FILE_PATH | base64 | tr -d '\n')
    kubectl -n kc patch secret $SECRET_NAME -p="{\"data\":{\"$DATA_FIELD_NAME\":\"$ENCODED_VALUE\"}}"
else
    # 不存在
    kubectl -n kc create secret generic $SECRET_NAME --from-file=$DATA_FIELD_NAME=$FILE_PATH
fi
