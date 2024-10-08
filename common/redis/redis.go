package redis

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"kcaitech.com/kcserver/common/redis/config"
)

var Client *redis.Client
var RedSync *redsync.Redsync

func Init(conf *config.RedisConf) error {
	// conf := config.LoadConfig(filePath)
	if conf.Sentinel {
		Client = redis.NewFailoverClient(&redis.FailoverOptions{
			SentinelAddrs:    conf.SentinelAddrs,
			MasterName:       conf.MasterName,
			Password:         conf.Password,
			SentinelPassword: conf.Password,
			DB:               conf.Db,
		})
	} else {
		Client = redis.NewClient(&redis.Options{
			Addr:     conf.Addr,
			Password: conf.Password,
			DB:       conf.Db,
		})
	}

	RedSync = redsync.New(goredis.NewPool(Client))

	return nil
}

const Nil = redis.Nil
