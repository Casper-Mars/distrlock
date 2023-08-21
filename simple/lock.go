package simple

import (
	"context"
	"time"
)

var lock Api

func Register(cService Api) {
	lock = cService
}

func GetService() Api {
	return lock
}

// Lock 加锁，如果加锁失败，则返回false。如果配置了重试参数，则会一直尝试获取锁，直到获取成功或者尝试次数用完
func Lock(ctx context.Context, key string, expireTs time.Duration, opts ...Option) (isLocked bool) {
	return lock.Lock(ctx, key, expireTs, opts...)
}

// Unlock 释放锁
func Unlock(ctx context.Context, key string) error {
	return lock.Unlock(ctx, key)
}

// TryLock 尝试获取锁，如果获取不到，则立即返回false
func TryLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool) {
	return lock.TryLock(ctx, key, expireTs)
}
