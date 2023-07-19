package redis

import (
	"github.com/redis/go-redis/v9"
	"protodesign.cn/kcserver/common/redis/config"
)

var Client *redis.Client

func Init(filePath string) error {
	conf := config.LoadConfig(filePath)
	Client = redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Addr,
		Password: conf.Redis.Password,
		DB:       conf.Redis.Db,
	})
	return nil
}
