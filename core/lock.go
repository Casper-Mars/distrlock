package core

import (
	"context"
)

type Locker interface {
	// Lock key. Blocked
	Lock(ctx context.Context) (isLocked bool, err error)
	// Unlock key
	Unlock(ctx context.Context) error
	// TryLock If lock failed, return false
	TryLock(ctx context.Context) (isLocked bool, err error)
}
