// Command simple 演示用 simple.NewLocker 在本机 Redis 上获取/释放分布式锁。
//
// 前置：本地 Redis 监听 localhost:6379（如 `docker run -p 6379:6379 redis:7-alpine`）。
// 运行：go run ./examples/simple
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Casper-Mars/distrlock/core/simple"
)

const (
	redisAddr = "localhost:6379"
	lockKey   = "distrlock:example:simple"
	lockWait  = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "example simple failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cli := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer func() {
		_ = cli.Close()
	}()

	// simple.NewLocker 的 Lock 是 100ms 轮询的阻塞循环，用超时兜底，
	// 避免无 Redis 或锁被占用时无限等待。
	ctx, cancel := context.WithTimeout(context.Background(), lockWait)
	defer cancel()

	locker := simple.NewLocker(cli, lockKey)

	locked, err := locker.Lock(ctx)
	if err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	if !locked {
		return fmt.Errorf("acquire lock: %s not locked", lockKey)
	}
	fmt.Println("lock acquired")

	if err := locker.Unlock(ctx); err != nil {
		return fmt.Errorf("release lock: %w", err)
	}
	fmt.Println("lock released")

	return nil
}
