package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"kcaitech.com/kcserver/common/mongo/config"
	"time"
)

var Client *mongo.Client
var DB *mongo.Database
var IsDuplicateKeyError = mongo.IsDuplicateKeyError

func Init(filePath string) error {
	conf := config.LoadConfig(filePath)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	option := options.Client().ApplyURI(conf.Mongo.Uri)
	//option.SetReadConcern(readconcern.Majority())
	//option.SetReadConcern(readconcern.Snapshot())
	option.SetReadPreference(readpref.PrimaryPreferred())
	option.SetWriteConcern(writeconcern.New(writeconcern.WMajority()))
	if Client, err = mongo.Connect(ctx, option); err != nil {
		return err
	}
	DB = Client.Database(conf.Mongo.Db)
	return nil
}

type SessionContext = mongo.SessionContext

func WithSession(ctx context.Context, sess mongo.Session, fn func(mongo.SessionContext) error) error {
	return mongo.WithSession(ctx, sess, fn)
}

func UseSession(ctx context.Context, fn func(mongo.SessionContext) error) error {
	return Client.UseSession(ctx, fn)
}
