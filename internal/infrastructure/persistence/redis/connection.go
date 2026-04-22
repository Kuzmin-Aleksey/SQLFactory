package redis

import (
	"SQLFactory/internal/config"
	"context"
	"github.com/redis/go-redis/v9"
)

const (
	tokensDB = iota
	confirmEmailDB
)

func connect(cfg config.RedisConfig, dbId int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       dbId,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}

func NewTokensCache(cfg config.RedisConfig) (*Adapter, error) {
	redisClient, err := connect(cfg, tokensDB)
	if err != nil {
		return nil, err
	}
	return &Adapter{redisClient}, nil
}

func NewConfirmEmailCache(cfg config.RedisConfig) (*Adapter, error) {
	redisClient, err := connect(cfg, confirmEmailDB)
	if err != nil {
		return nil, err
	}
	return &Adapter{redisClient}, nil
}
