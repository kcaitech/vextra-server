package redis

import (
	"time"

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

	// 基础选项配置
	baseOptions := &redis.Options{
		// 连接池设置
		PoolSize:     50, // 连接池最大连接数
		MinIdleConns: 10, // 最小空闲连接数

		// 超时设置
		DialTimeout:  5 * time.Second, // 建立连接超时
		ReadTimeout:  3 * time.Second, // 读取超时
		WriteTimeout: 3 * time.Second, // 写入超时
		PoolTimeout:  4 * time.Second, // 当连接池达到最大时，获取连接的等待时间

		// 重试设置
		MaxRetries:      3, // 命令执行失败时的重试次数
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	}

	if conf.Sentinel {
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			SentinelAddrs:    conf.SentinelAddrs,
			MasterName:       conf.MasterName,
			Password:         conf.Password,
			SentinelPassword: conf.Password,
			DB:               conf.Db,

			// 继承基础配置
			PoolSize:        baseOptions.PoolSize,
			MinIdleConns:    baseOptions.MinIdleConns,
			DialTimeout:     baseOptions.DialTimeout,
			ReadTimeout:     baseOptions.ReadTimeout,
			WriteTimeout:    baseOptions.WriteTimeout,
			PoolTimeout:     baseOptions.PoolTimeout,
			MaxRetries:      baseOptions.MaxRetries,
			MinRetryBackoff: baseOptions.MinRetryBackoff,
			MaxRetryBackoff: baseOptions.MaxRetryBackoff,
		})
	} else {
		baseOptions.Addr = conf.Addr
		baseOptions.Password = conf.Password
		baseOptions.DB = conf.Db
		client = redis.NewClient(baseOptions)
	}

	redSync := redsync.New(goredis.NewPool(client))
	return &RedisDB{
		Client:  client,
		RedSync: redSync,
	}, nil
}

const Nil = redis.Nil
