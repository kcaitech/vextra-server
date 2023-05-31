package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"protodesign.cn/kcserver/common/mongo/config"
	"time"
)

var Client *mongo.Client
var DB *mongo.Database

func Init(filePath string) error {
	conf := config.LoadConfig(filePath)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	if Client, err = mongo.Connect(ctx, options.Client().ApplyURI(conf.Mongo.Uri)); err != nil {
		return err
	}
	DB = Client.Database(conf.Mongo.Db)
	return nil
}
