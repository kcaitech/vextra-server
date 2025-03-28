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
	option.SetReadPreference(readpref.PrimaryPreferred())
	option.SetWriteConcern(writeconcern.New(writeconcern.WMajority()))
	client, err := mongo.Connect(ctx, option)
	if err != nil {
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
