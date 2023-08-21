package lock

import (
	"context"
	"time"
)

type SimpleLock interface {
	// Lock key with expireTs, if lock failed, return false. If retry is set, it will try to get lock until success or retry times is used up.
	Lock(ctx context.Context, key string, expireTs time.Duration, opts ...Option) (isLocked bool)
	// Unlock key
	Unlock(ctx context.Context, key string) error
	// TryLock try to get lock, if failed, return false immediately.
	TryLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool)
}
