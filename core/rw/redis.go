package rw

import (
	"context"
	"github.com/Casper-Mars/distrlock/core"
	"github.com/redis/go-redis/v9"
)

type locker struct {
	cli redis.Cmdable
	key string
}

func NewLocker(cli redis.Cmdable, key string) core.Locker {
	return &locker{
		cli: cli,
		key: key,
	}
}

func (l *locker) Lock(ctx context.Context) (isLocked bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (l *locker) Unlock(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (l *locker) TryLock(ctx context.Context) (isLocked bool, err error) {
	//TODO implement me
	panic("implement me")
}
