package redis

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"kcaitech.com/kcserver/common/redis/config"
)

var Client *redis.Client
var RedSync *redsync.Redsync

func Init(filePath string) error {
	conf := config.LoadConfig(filePath)
	if conf.Redis.Sentinel {
		Client = redis.NewFailoverClient(&redis.FailoverOptions{
			SentinelAddrs:    conf.Redis.SentinelAddrs,
			MasterName:       conf.Redis.MasterName,
			Password:         conf.Redis.Password,
			SentinelPassword: conf.Redis.Password,
			DB:               conf.Redis.Db,
		})
	} else {
		Client = redis.NewClient(&redis.Options{
			Addr:     conf.Redis.Addr,
			Password: conf.Redis.Password,
			DB:       conf.Redis.Db,
		})
	}

	RedSync = redsync.New(goredis.NewPool(Client))

	return nil
}

const Nil = redis.Nil
