package rw

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

func RegisterWithRedisImpl(cli redis.Cmdable) {
	r := NewRedisImpl(cli)
	Register(r)
}

type redisImpl struct {
	cli redis.Cmdable
}

func NewRedisImpl(cli redis.Cmdable) Api {
	return &redisImpl{
		cli: cli,
	}
}

func (r *redisImpl) RLock(ctx context.Context, key string, expireTs time.Duration, opts ...Option) (isLocked bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (r *redisImpl) RUnlock(ctx context.Context, key string) error {
	//TODO implement me
	panic("implement me")
}

func (r *redisImpl) TryRLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (r *redisImpl) WLock(ctx context.Context, key string, expireTs time.Duration, opts ...Option) (isLocked bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (r *redisImpl) WUnlock(ctx context.Context, key string) error {
	//TODO implement me
	panic("implement me")
}

func (r *redisImpl) TryWLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, err error) {
	//TODO implement me
	panic("implement me")
}
