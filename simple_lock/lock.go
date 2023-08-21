package lock

import (
	"context"
	"time"
)

var service Service

func Register(cService Service) {
	service = cService
}

func GetService() Service {
	return service
}

// Lock 加锁，如果加锁失败，则返回false。如果配置了重试参数，则会一直尝试获取锁，直到获取成功或者尝试次数用完
func Lock(ctx context.Context, key string, expireTs time.Duration, opts ...Option) (isLocked bool) {
	return service.Lock(ctx, key, expireTs, opts...)
}

// Unlock 释放锁
func Unlock(ctx context.Context, key string) error {
	return service.Unlock(ctx, key)
}

// TryLock 尝试获取锁，如果获取不到，则立即返回false
func TryLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool) {
	return service.TryLock(ctx, key, expireTs)
}
