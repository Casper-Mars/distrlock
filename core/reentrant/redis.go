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
    cli       redis.Cmdable
    key       string
    expire    time.Duration
    ca        string
    holderCnt int // 持有锁计数
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
                // 首次获得锁
                return true, nil
            }

            // 检查自己是否是锁的持有者
            val, err := l.cli.Get(ctx, l.key).Result()
            if err != nil {
                return false, err
            }

            if val == l.ca {
                // 当前锁的所有者，锁持有者数量+1
                l.holderCnt += 1
                return true, nil
            }

            l.holderCnt = 0 // 防止一些由于锁自动过期，没有更新holderCnt的情况
            return false, nil
        }
        time.Sleep(time.Second * 100)
    }
}

func (l *locker) Unlock(ctx context.Context) error {
    if l.holderCnt > 0 {
        l.holderCnt -= 1
        return nil
    }

    // 如果持有锁的数量==0，真正释放锁
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
    result, err := l.cli.SetNX(ctx, l.key, l.ca, l.expire).Result()
    if err != nil {
        return false, err
    }

    return result, nil
}
