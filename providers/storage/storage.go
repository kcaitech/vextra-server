/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package storage

import (
	"errors"
)

type StorageConf struct {
	Provider Provider `yaml:"provider" json:"provider"`
	Minio    Config   `yaml:"minio" json:"minio"`
	S3       Config   `yaml:"s3" json:"s3"`
	Oss      Config   `yaml:"oss" json:"oss"`
}

type StorageClient struct {
	Client        Client
	Bucket        Bucket
	AttatchBucket Bucket
}

func NewStoraageClient(config *Config) (*StorageClient, error) {
	var client Client
	var err error
	var bucketConfig = config.BucketConfig
	switch config.Provider {
	case MINIO:
		client, err = NewMinioClient(&config.ClientConfig)
	case S3:
		client, err = NewS3Client(&config.ClientConfig)
	case OSS:
		client, err = NewOSSClient(&config.ClientConfig)
	default:
		return nil, errors.New("不支持的provider")
	}

	if err != nil {
		return nil, err
	}

	return &StorageClient{
		Client: client,
		Bucket: client.NewBucket(&BucketConfig{
			DocumentBucket: bucketConfig.DocumentBucket,
		}),
		AttatchBucket: client.NewBucket(&BucketConfig{
			DocumentBucket: bucketConfig.AttatchBucket,
		}),
	}, nil
}
