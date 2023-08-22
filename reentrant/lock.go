package reentrant

import (
	"context"
	"time"
)

var lock Api

func Register(clock Api) {
	lock = clock
}

func Get() Api {
	return lock
}

// Lock key with expireTs and return ca. Blocked
func Lock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, ca string, err error) {
	return lock.Lock(ctx, key, expireTs)
}

// Unlock key with ca. If ca is wrong, return ErrWrongCa
func Unlock(ctx context.Context, key string, ca string) error {
	return lock.Unlock(ctx, key, ca)
}

// TryLock key with expireTs and return ca. If lock failed, return false
func TryLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, ca string, err error) {
	return lock.TryLock(ctx, key, expireTs)
}
