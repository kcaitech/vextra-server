package redis

import (
	"context"
	"log"
	"testing"
	"time"

	"kcaitech.com/kcserver/common/redis/config"
)

func TestMain(m *testing.M) {
	if err := Init(&config.LoadConfig("config_test.yaml").Redis); err != nil {
		log.Fatalln("redis初始化失败：" + err.Error())
	}
	m.Run()
}

func TestGetSet(t *testing.T) {
	ctx := context.Background()
	if err := Client.Set(ctx, "test", "test", 0).Err(); err != nil {
		log.Fatalln("redis设置失败：" + err.Error())
	}
	val, err := Client.Get(ctx, "test").Result()
	if err != nil {
		log.Fatalln("redis获取失败：" + err.Error())
	}
	log.Println("redis获取成功：" + val)
}

func TestPublish(t *testing.T) {
	ctx := context.Background()
	if err := Client.Publish(ctx, "test", "test").Err(); err != nil {
		log.Fatalln("redis发布失败：" + err.Error())
	}
}

func TestSubscribe(t *testing.T) {
	ctx := context.Background()
	pubsub := Client.Subscribe(ctx, "143252945616519168")
	defer pubsub.Close()
	ch := pubsub.Channel()
	select {
	case msg := <-ch:
		log.Println("redis订阅成功：" + msg.Payload)
	case <-time.After(60 * time.Second):
		log.Fatal("redis订阅超时")
	}
}
