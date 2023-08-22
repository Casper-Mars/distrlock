package reentrant

import (
	"context"
	"errors"
	"github.com/Casper-Mars/distrlock/core"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	ErrNotBelong = errors.New("cannot unlock a lock that does not belong to this context")
)

type locker struct {
	cli redis.Cmdable
	key string
	ca  string
}

func NewLocker(cli redis.Cmdable, key string) core.Locker {
	return &locker{
		cli: cli,
		key: key,
		ca:  uuid.NewString(),
	}
}

func (l *locker) Lock(ctx context.Context) (isLocked bool, err error) {
	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			locked, err := l.TryLock(ctx)
			if err != nil {
				return false, err
			}
			if locked {
				return true, nil
			}
		}
		time.Sleep(time.Second * 100)
	}
}

func (l *locker) Unlock(ctx context.Context) error {
	script := `
		local lock_key = KEYS[1]
		local lock_val = ARGV[1]
		local current_val = redis.call('GET', lock_key)
		if current_val == lock_val then
			redis.call('DEL', lock_key)
			return 1
		else
			return 0
		end
	`
	result, err := l.cli.Eval(ctx, script, []string{l.key}, l.ca).Result()
	if err != nil {
		return err
	}

	if result != int64(1) {
		return ErrNotBelong
	}

	return nil
}

func (l *locker) TryLock(ctx context.Context) (isLocked bool, err error) {
	script := `
		local lock_key = KEYS[1]
		local lock_val = ARGV[1]
		local current_val = redis.call('GET', lock_key)
		if current_val == lock_val then
			return 1
		elseif not current_val then
			redis.call('SET', lock_key, lock_val)
			return 1
		else
			return 0
		end
	`
	result, err := l.cli.Eval(ctx, script, []string{l.key}, l.ca).Result()
	if err != nil {
		return false, err
	}

	return result == int64(1), nil
}
