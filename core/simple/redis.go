package simple

import (
	"context"
	"fmt"
	"github.com/Casper-Mars/distrlock/core"
	"github.com/redis/go-redis/v9"
	"time"
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
	for {
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("redis lock failed, err:%v", ctx.Err())
		default:
			locked, err := l.TryLock(ctx)
			if err != nil {
				return false, err
			}
			if locked {
				return true, nil
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func (l *locker) Unlock(ctx context.Context) error {
	err := l.cli.Del(ctx, l.key).Err()
	if err != nil {
		return fmt.Errorf("redis del failed, err:%v", err)
	}
	return nil
}

func (l *locker) TryLock(ctx context.Context) (isLocked bool, err error) {
	result, err := l.cli.SetNX(ctx, l.key, "locked", 0).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx failed, err:%v", err)
	}
	if result {
		return true, nil
	}
	return false, nil
}
