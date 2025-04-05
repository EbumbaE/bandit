package redis

import (
	"context"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
}

func NewRedis(ctx context.Context, addr string) (*Client, error) {
	opt, err := redis.ParseURL(addr)
	if err != nil {
		return nil, errors.Wrap(err, "redis connection parse error")
	}

	client := redis.NewClient(opt)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errors.Wrap(err, "redis connection failed")
	}

	return &Client{client}, nil
}
