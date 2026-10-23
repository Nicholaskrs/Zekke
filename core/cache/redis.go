package cache

import (
	"context"
	"fmt"
	"template-go/util/config"

	"github.com/redis/go-redis/v9"
)

func NewClient(cfg *config.Config) (*redis.Client, error) {
	address := fmt.Sprintf("%s:%s", cfg.RedisAddr, cfg.RedisPort)
	rdb := redis.NewClient(&redis.Options{Addr: address, Password: cfg.RedisPassword, Username: cfg.RedisUsername})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}
