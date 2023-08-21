package reentrant_lock

import (
	"context"
	"time"
)

var service ReentrantLock

func Register(cService ReentrantLock) {
	service = cService
}

func GetLocker() ReentrantLock {
	return service
}

// Lock key with expireTs and return ca. If lock failed, return false. If retry is set, it will retry to lock until success or retry times used up
func Lock(ctx context.Context, key string, expireTs time.Duration, opts ...Option) (isLocked bool, ca string, err error) {
	return service.Lock(ctx, key, expireTs, opts...)
}

// Unlock key with ca. If ca is wrong, return ErrWrongCa
func Unlock(ctx context.Context, key string, ca string) error {
	return service.Unlock(ctx, key, ca)
}

// TryLock key with expireTs and return ca. If lock failed, return false
func TryLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, ca string, err error) {
	return service.TryLock(ctx, key, expireTs)
}
