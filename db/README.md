

## minio直接使用缓存配置（需要手动操作）
1. http://localhost:9001/buckets  user: kcserver passwd: kcai1212
2. 创建 document,files bucket, files设置policy为custom:
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": [
                    "*"
                ]
            },
            "Action": [
                "s3:GetBucketLocation",
                "s3:GetObject"
            ],
            "Resource": [
                "arn:aws:s3:::*"
            ]
        }
    ]
}
3. 创建accesskey, 添加user
"accessKeyID": "59L3pDWuk9oLQWIgytCI",
"secretAccessKey": "GmwAzetY1L0jU4pUAqpO7RJGIsG7k3gozVH6whoJ",
"stsAccessKeyID": "user",
"stsSecretAccessKey": "GmwAzetY1L0jU4pUAqpO7RJGIsG7k3gozVH6whoJ",
user权限选择readonly
(已配置)

## mongodb需要设置{document_id, cmd_id}为unique（已配置）
1. 见db/mongodb/init

## mysql初始化（已配置）
1. 见db/mysql/init