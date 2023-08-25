package reentrant

import (
    "context"
    "errors"
    "github.com/Casper-Mars/distrlock/core"
    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"
    "sync"
    "time"
)

var (
    ErrNotBelong = errors.New("cannot unlock a lock that does not belong to this context")
)

type Counter struct {
    mu    sync.Mutex
    count int
}

// 增加计数
func (c *Counter) incr() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

// 减少计数，并返回计数
func (c *Counter) decrWithRet() int {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.count >= 0 {
        c.count--
    }
    count := c.count
    return count
}

// 初始化count为1
func (c *Counter) initCount() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count = 1
}

type locker struct {
    cli       redis.Cmdable
    key       string
    expire    time.Duration
    ca        string
    holderCnt *Counter // 持有锁计数
}

func NewLocker(cli redis.Cmdable, key string) core.Locker {
    return &locker{
        cli:       cli,
        key:       key,
        ca:        uuid.NewString(),
        holderCnt: &Counter{},
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
                l.holderCnt.initCount()
                return true, nil
            }

            val, err := l.cli.Get(ctx, l.key).Result()
            if err != nil {
                return false, err
            }

            if val == l.ca {
                // 当前锁的所有者，锁持有者数量+1
                l.holderCnt.incr()
                return true, nil
            }

            return false, nil
        }
    }
}

func (l *locker) Unlock(ctx context.Context) error {

    // 当锁最后一个持有者释放锁时，真正释放锁
    if l.holderCnt.decrWithRet() == 0 {
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
    result, err := l.cli.SetNX(ctx, l.key, l.ca, l.expire).Result()
    if err != nil {
        return false, err
    }

    return result, nil
}
