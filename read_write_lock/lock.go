package read_write_lock

import (
	"context"
	"time"
)

var service RWLock

func Register(cService RWLock) {
	service = cService
}

func GetLocker() RWLock {
	return service
}

// RLock require read-lock
func RLock(ctx context.Context, key string, expireTs time.Duration, opts ...Option) (isLocked bool, err error) {
	return service.RLock(ctx, key, expireTs, opts...)
}

// RUnlock release read-lock
func RUnlock(ctx context.Context, key string) error {
	return service.RUnlock(ctx, key)
}

// TryRLock try to get read-lock, if failed, return false immediately.
func TryRLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, err error) {
	return service.TryRLock(ctx, key, expireTs)
}

// WLock require write lock
func WLock(ctx context.Context, key string, expireTs time.Duration, opts ...Option) (isLocked bool, err error) {
	return service.WLock(ctx, key, expireTs, opts...)
}

// WUnlock release write lock
func WUnlock(ctx context.Context, key string) error {
	return service.WUnlock(ctx, key)
}

// TryWLock try to get write-lock, if failed, return false immediately.
func TryWLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, err error) {
	return service.TryWLock(ctx, key, expireTs)
}
