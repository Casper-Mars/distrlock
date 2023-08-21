package lock

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisLocker struct {
	cli redis.Cmdable
}

func RegisterWithRedisImpl(cli redis.Cmdable) {
	r := NewRedisLockService(cli)
	Register(r)
}

func NewRedisLockService(cli redis.Cmdable) SimpleLock {
	return &redisLocker{
		cli: cli,
	}
}

func (r *redisLocker) Lock(ctx context.Context, key string, expireTs time.Duration, opts ...Option) bool {
	// 初始化配置
	o := &lockOptions{
		waitTime: time.Second,
	}
	for _, opt := range opts {
		opt(o)
	}
	// 尝试获取锁
	success := r.TryLock(ctx, key, expireTs)
	// 获取锁成功，则直接返回
	if success {
		return true
	}
	// 获取锁失败，并且配置了重试次数，则重试获取锁
	if o.retry != 0 {
		for i := 0; i < o.retry; i++ {
			time.Sleep(o.waitTime)
			success = r.TryLock(ctx, key, expireTs)
			if success {
				return true
			}
		}
	}
	return false
}

func (r *redisLocker) Unlock(ctx context.Context, key string) error {
	err := r.cli.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *redisLocker) TryLock(ctx context.Context, key string, expireTs time.Duration) bool {
	result, err := r.cli.SetNX(ctx, key, "1", expireTs).Result()
	if err != nil {
		return false
	}
	if result {
		return true
	}
	return false
}
