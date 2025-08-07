/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// var Client *mongo.Client
// var DB *mongo.Database
var IsDuplicateKeyError = mongo.IsDuplicateKeyError

type MongoConf struct {
	Url string `yaml:"url" json:"url"`
	Db  string `yaml:"db" json:"db"`
}

type MongoDB struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func NewMongoDB(conf *MongoConf) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	option := options.Client().ApplyURI(conf.Url)
	// 设置连接池
	option.SetMaxPoolSize(100)
	option.SetMinPoolSize(5)
	option.SetMaxConnecting(10)

	// 设置超时时间
	option.SetConnectTimeout(5 * time.Second)
	option.SetServerSelectionTimeout(5 * time.Second)
	option.SetSocketTimeout(10 * time.Second)

	// 设置心跳检测
	option.SetHeartbeatInterval(5 * time.Second)

	// 设置重试
	option.SetRetryWrites(true)
	option.SetRetryReads(true)

	// 读写偏好设置
	option.SetReadPreference(readpref.PrimaryPreferred())
	option.SetWriteConcern(writeconcern.New(writeconcern.WMajority()))

	client, err := mongo.Connect(ctx, option)
	if err != nil {
		return nil, err
	}

	// 验证连接
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	db := client.Database(conf.Db)
	return &MongoDB{
		Client: client,
		DB:     db,
	}, nil
}

func (m *MongoDB) Close() error {
	return m.Client.Disconnect(context.Background())
}

func (m *MongoDB) UseSession(ctx context.Context, fn func(mongo.SessionContext) error) error {
	return m.Client.UseSession(ctx, fn)
}

func (m *MongoDB) WithSession(ctx context.Context, sess mongo.Session, fn func(mongo.SessionContext) error) error {
	return mongo.WithSession(ctx, sess, fn)
}

type SessionContext = mongo.SessionContext

func WithSession(ctx context.Context, sess mongo.Session, fn func(mongo.SessionContext) error) error {
	return mongo.WithSession(ctx, sess, fn)
}
