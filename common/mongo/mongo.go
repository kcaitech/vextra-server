package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"kcaitech.com/kcserver/common/mongo/config"
)

var Client *mongo.Client
var DB *mongo.Database
var IsDuplicateKeyError = mongo.IsDuplicateKeyError

func Init(conf *config.MongoConf) error {
	// conf := config.LoadConfig(filePath)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	option := options.Client().ApplyURI(conf.Url)
	//option.SetReadConcern(readconcern.Majority())
	//option.SetReadConcern(readconcern.Snapshot())
	option.SetReadPreference(readpref.PrimaryPreferred())
	option.SetWriteConcern(writeconcern.New(writeconcern.WMajority()))
	if Client, err = mongo.Connect(ctx, option); err != nil {
		return err
	}
	DB = Client.Database(conf.Db)
	return nil
}

type SessionContext = mongo.SessionContext

func WithSession(ctx context.Context, sess mongo.Session, fn func(mongo.SessionContext) error) error {
	return mongo.WithSession(ctx, sess, fn)
}

func UseSession(ctx context.Context, fn func(mongo.SessionContext) error) error {
	return Client.UseSession(ctx, fn)
}
