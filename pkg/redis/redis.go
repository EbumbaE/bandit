package redis

import (
	"context"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

var ErrNotFound = errors.New("key not found")

type Redis interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}) error
	ZGet(ctx context.Context, key string) ([]string, error)
	ZSet(ctx context.Context, key string, score float64, member interface{}) error
	Close() error
}

type redisClient struct {
	client *redis.Client
}

func NewRedis(ctx context.Context, addr string) (Redis, error) {
	opt, err := redis.ParseURL(addr)
	if err != nil {
		return nil, errors.Wrap(err, "redis connection parse error")
	}

	client := redis.NewClient(opt)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errors.Wrap(err, "redis connection failed")
	}

	return &redisClient{client: client}, nil
}

func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}

	return val, err
}

func (r *redisClient) Set(ctx context.Context, key string, value any) error {
	cmd := r.client.Set(ctx, key, value, 0)

	return cmd.Err()
}

func (r *redisClient) ZSet(ctx context.Context, key string, score float64, member any) error {
	cmd := r.client.ZAdd(ctx, key, redis.Z{Score: score, Member: member})

	return cmd.Err()
}

func (r *redisClient) ZGet(ctx context.Context, key string) ([]string, error) {
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if exists == 0 {
		return nil, ErrNotFound
	}

	return r.client.ZRange(ctx, key, 0, -1).Result()
}

func (r *redisClient) Close() error {
	return r.client.Close()
}
