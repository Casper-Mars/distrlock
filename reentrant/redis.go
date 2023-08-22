package reentrant

import (
	"context"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

type redisLocker struct {
	cli redis.Cmdable
}

func RegisterWithRedisImpl(cli redis.Cmdable) {
	r := NewRedisLockService(cli)
	Register(r)
}

func NewRedisLockService(cli redis.Cmdable) Api {
	return &redisLocker{
		cli: cli,
	}
}

func (r *redisLocker) Lock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, ca string, err error) {
	for {
		select {
		case <-ctx.Done():
			return false, "", ctx.Err()
		default:
			// try lock
			isLocked, ca, err = r.TryLock(ctx, key, expireTs)
			if err != nil {
				return false, "", err
			}
			// if lock success, return
			if isLocked {
				return true, ca, nil
			}
		}
		time.Sleep(time.Second * 100)
	}
}

func (r *redisLocker) Unlock(ctx context.Context, key string, ca string) error {
	// TODO: only unlock when ca is matched, otherwise return ErrWrongCa
	err := r.cli.Del(ctx, key).Err()
	if err != nil {
		//log.Errorf("Unlock Del failed key %s err %v", key, err)
		return err
	}
	return nil
}

func (r *redisLocker) TryLock(ctx context.Context, key string, expireTs time.Duration) (isLocked bool, ca string, err error) {
	ca = uuid.NewString()
	result, err := r.cli.SetNX(ctx, key, ca, expireTs).Result()
	if err != nil {
		return false, "", err
	}
	if result {
		return true, ca, nil
	}
	return false, "", nil
}
