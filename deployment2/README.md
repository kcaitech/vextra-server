1. 需要先去构建svg2png,kcversion的docker镜像
2. ./run.sh up
3. http://localhost:9001/buckets 创建document,files bucket, 创建accesskey, 添加user
4. 设置minio中files bucket为所有人可读：
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