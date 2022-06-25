package redis

import (
	"context"
	"fmt"

	redis "github.com/go-redis/redis/v8"
	"gitlab.com/mgdi/kongroo-c2/c2/config"
)

var RedisCl *Client

type Client struct {
	Client  *redis.Client
	Context context.Context
}

func NewClient() {
	redisCl := Client{}
	redisCl.Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Configs["redis.host"], config.Configs["redis.port"]),
		Password: config.Configs["redis.password"],
		DB:       0,
	})
	redisCl.Context = context.Background()

	RedisCl = &redisCl
}

func (cl *Client) Get(key string) (string, error) {
	val, err := RedisCl.Client.GetDel(RedisCl.Context, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (cl *Client) Set(key string, value string) error {
	err := RedisCl.Client.Set(RedisCl.Context, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}
