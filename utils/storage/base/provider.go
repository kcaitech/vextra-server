package base

type Provider string

const (
	MINIO Provider = "minio"
	S3    Provider = "s3"
	OSS   Provider = "oss"
)
