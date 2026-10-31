// Command reentrant 演示同一 reentrant.Locker 实例的重入与释放。
//
// 前置：本地 Redis 监听 localhost:6379。
// 运行：go run ./examples/reentrant
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Casper-Mars/distrlock/core/reentrant"
)

const (
	redisAddr = "localhost:6379"
	lockKey   = "distrlock:example:reentrant"
	lockWait  = 5 * time.Second
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "example reentrant failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cli := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer func() {
		_ = cli.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), lockWait)
	defer cancel()

	locker := reentrant.NewLocker(cli, lockKey)

	// 同一个实例连续加锁两次：第二次走 holderCnt 计数重入，不重复写 Redis。
	for i := 1; i <= 2; i++ {
		locked, err := locker.TryLock(ctx)
		if err != nil {
			return fmt.Errorf("try lock #%d: %w", i, err)
		}
		if !locked {
			return fmt.Errorf("try lock #%d: not locked", i)
		}
		fmt.Printf("try lock #%d: ok (重入计数 %d)\n", i, i)
	}

	// 连续释放两次：前一次只减计数，最后一次才真正 DEL 掉 key。
	for i := 1; i <= 2; i++ {
		if err := locker.Unlock(ctx); err != nil {
			return fmt.Errorf("unlock #%d: %w", i, err)
		}
		fmt.Printf("unlock #%d: ok\n", i)
	}

	fmt.Println("重入 2 次、释放 2 次")
	return nil
}
