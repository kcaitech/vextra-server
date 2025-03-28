package redis

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

type RedisConf struct {
	Addr          string   `yaml:"addr" json:"addr"`
	Password      string   `yaml:"password" json:"password"`
	Db            int      `yaml:"db" json:"db"`
	Sentinel      bool     `yaml:"sentinel" json:"sentinel"`
	SentinelAddrs []string `yaml:"sentinelAddrs" json:"sentinelAddrs"`
	MasterName    string   `yaml:"masterName" json:"masterName"`
}

type RedisDB struct {
	Client  *redis.Client
	RedSync *redsync.Redsync
}

func NewRedisDB(conf *RedisConf) (*RedisDB, error) {
	var client *redis.Client
	if conf.Sentinel {
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			SentinelAddrs:    conf.SentinelAddrs,
			MasterName:       conf.MasterName,
			Password:         conf.Password,
			SentinelPassword: conf.Password,
			DB:               conf.Db,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:     conf.Addr,
			Password: conf.Password,
			DB:       conf.Db,
		})
	}

	redSync := redsync.New(goredis.NewPool(client))
	return &RedisDB{
		Client:  client,
		RedSync: redSync,
	}, nil
}

const Nil = redis.Nil
