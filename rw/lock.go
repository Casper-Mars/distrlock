package rw

import (
	"context"
	"time"
)

var lock Api

func Register(cService Api) {
	lock = cService
}

func GetLocker() Api {
	return lock
}

// RLock require read-lock
func RLock(ctx context.Context, key string, expireTs time.Duration, opts ...Option) (isLocked bool, err error) {
	return lock.RLock(ctx, key, expireTs, opts...)
}

// RUnlock release read-lock
func RUnlock(ctx context.Context, key string) error {
	return lock.RUnlock(ctx, key)
}

// TryRLock try to get read-lock, if failed, return false immediately.
func TryRLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, err error) {
	return lock.TryRLock(ctx, key, expireTs)
}

// WLock require write lock
func WLock(ctx context.Context, key string, expireTs time.Duration, opts ...Option) (isLocked bool, err error) {
	return lock.WLock(ctx, key, expireTs, opts...)
}

// WUnlock release write lock
func WUnlock(ctx context.Context, key string) error {
	return lock.WUnlock(ctx, key)
}

// TryWLock try to get write-lock, if failed, return false immediately.
func TryWLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, err error) {
	return lock.TryWLock(ctx, key, expireTs)
}
