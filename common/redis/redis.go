package redis

import (
	"github.com/redis/go-redis/v9"
	"protodesign.cn/kcserver/common/redis/config"
)

var Client *redis.Client

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
	return nil
}
