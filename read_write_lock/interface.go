package read_write_lock

import (
	"context"
	"time"
)

type RWLock interface {
	// RLock require read-lock
	RLock(ctx context.Context, key string, expireTs time.Duration, opts ...Option) (isLocked bool, err error)
	// RUnlock release read-lock
	RUnlock(ctx context.Context, key string) error
	// TryRLock try to get read-lock, if failed, return false immediately.
	TryRLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, err error)
	// WLock require write lock
	WLock(ctx context.Context, key string, expireTs time.Duration, opts ...Option) (isLocked bool, err error)
	// WUnlock release write lock
	WUnlock(ctx context.Context, key string) error
	// TryWLock try to get write-lock, if failed, return false immediately.
	TryWLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, err error)
}
