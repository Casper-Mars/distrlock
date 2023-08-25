package reentrant

import (
    "context"
    "errors"
    "github.com/Casper-Mars/distrlock/core"
    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"
    "sync/atomic"
    "time"
)

var (
    ErrNotBelong = errors.New("cannot unlock a lock that does not belong to this context")
)

type locker struct {
    cli       redis.Cmdable
    key       string
    ca        string
    holderCnt int32 // 持有锁计数
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
            // 尝试获取锁
            trySuccess, err := l.TryLock(ctx)
            if err != nil {
                return false, err
            }
            if trySuccess {
                return true, nil
            }
            time.Sleep(100 * time.Millisecond)
        }
    }
}

func (l *locker) Unlock(ctx context.Context) error {
    if atomic.LoadInt32(&l.holderCnt) <= 0 {
        return ErrNotBelong
    }

    // 锁持有者数量-1
    // 当锁最后一个持有者释放锁时，真正释放锁
    if atomic.AddInt32(&l.holderCnt, -1) == 0 {

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

        if result.(int64) != int64(1) {
            return ErrNotBelong
        }
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

    if result == int64(1) {
        atomic.AddInt32(&l.holderCnt, 1)
        return true, nil
    }

    // 非锁的持有者
    return false, nil
}
