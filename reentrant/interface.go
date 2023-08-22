package reentrant

import (
	"context"
	"errors"
	"time"
)

var (
	ErrWrongCa = errors.New("wrong ca")
)

type Api interface {
	// Lock key with expireTs and return ca. Blocked
	Lock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, ca string, err error)
	// Unlock key with ca. If ca is wrong, return ErrWrongCa
	Unlock(ctx context.Context, key string, ca string) error
	// TryLock key with expireTs and return ca. If lock failed, return false
	TryLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, ca string, err error)
}
