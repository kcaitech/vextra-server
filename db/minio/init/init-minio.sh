#!/bin/bash

# 等待 MinIO 服务启动
until curl -s http://localhost:9000/minio/health/live > /dev/null; do
    echo "Waiting for MinIO to start..."
    sleep 1
done

# 设置 MinIO 客户端别名
mc alias set myminio http://localhost:9000 $MINIO_ROOT_USER $MINIO_ROOT_PASSWORD

# 创建 document bucket（如果不存在）
if ! mc ls myminio/document > /dev/null 2>&1; then
    mc mb myminio/document
    echo "Created document bucket"
else
    echo "document bucket already exists"
fi

# 创建 attatch bucket（如果不存在）
# if ! mc ls myminio/attatch > /dev/null 2>&1; then
#     mc mb myminio/attatch
#     echo "Created attatch bucket"
# else
#     echo "attatch bucket already exists"
# fi

# 设置 attatch bucket 的自定义策略
# cat > /tmp/policy.json << EOF
# {
#     "Version": "2012-10-17",
#     "Statement": [
#         {
#             "Effect": "Allow",
#             "Principal": {
#                 "AWS": ["*"]
#             },
#             "Action": [
#                 "s3:GetBucketLocation",
#                 "s3:GetObject"
#             ],
#             "Resource": [
#                 "arn:aws:s3:::*"
#             ]
#         }
#     ]
# }
# EOF

# 创建自定义策略并应用到 attatch bucket
# mc admin policy create myminio custom /tmp/policy.json
# mc policy set myminio/attatch /tmp/policy.json
# mc anonymous set-json /tmp/policy.json myminio/attatch
# rm -f /tmp/policy.json

# 创建用户和访问密钥（如果不存在）
# if ! mc admin user info myminio user > /dev/null 2>&1; then
#     mc admin user add myminio user GmwAzetY1L0jU4pUAqpO7RJGIsG7k3gozVH6whoJ
#     mc admin policy attach myminio readonly --user user
#     echo "Created user and attached policy"
# else
#     echo "User already exists"
# fi

# 尝试创建服务账户，忽略错误
mc admin user svcacct add myminio $MINIO_ROOT_USER \
    --access-key "59L3pDWuk9oLQWIgytCI" \
    --secret-key "GmwAzetY1L0jU4pUAqpO7RJGIsG7k3gozVH6whoJ" 2>/dev/null || true

echo "MinIO initialization completed!" 