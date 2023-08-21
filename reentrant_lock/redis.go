package reentrant_lock

import (
	"context"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

type redisLocker struct {
	cli redis.Cmdable
}

func RegisterWithRedisImpl(cli redis.Cmdable) {
	r := NewRedisLockService(cli)
	Register(r)
}

func NewRedisLockService(cli redis.Cmdable) ReentrantLock {
	return &redisLocker{
		cli: cli,
	}
}

func (r *redisLocker) Lock(ctx context.Context, key string, expireTs time.Duration, opts ...Option) (isLocked bool, ca string, err error) {
	// 初始化配置
	o := &lockOptions{
		waitTime: time.Second,
	}
	for _, opt := range opts {
		opt(o)
	}
	// 尝试获取锁
	success, ca, err := r.TryLock(ctx, key, expireTs)
	// 获取锁成功，则直接返回
	if success {
		return true, ca, nil
	}
	// 获取锁失败，并且配置了重试次数，则重试获取锁
	if o.retry != 0 {
		for i := 0; i < o.retry; i++ {
			time.Sleep(o.waitTime)
			success, ca, err = r.TryLock(ctx, key, expireTs)
			if success {
				return true, ca, nil
			}
		}
	}
	return false, "", err
}

func (r *redisLocker) Unlock(ctx context.Context, key string, ca string) error {
	err := r.cli.Del(ctx, key).Err()
	if err != nil {
		//log.Errorf("Unlock Del failed key %s err %v", key, err)
		return err
	}
	return nil
}

func (r *redisLocker) TryLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, ca string, err error) {
	ca = uuid.NewString()
	result, err := r.cli.SetNX(ctx, key, ca, expireTs).Result()
	if err != nil {
		return false, "", err
	}
	if result {
		return true, ca, nil
	}
	return false, "", nil
}
