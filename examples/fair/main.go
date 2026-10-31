// Command fair 演示 fair.NewDistributedMutex 的排队取锁与释放。
//
// 注意：core/fair 的 NewLocker 返回的 core.Locker 三个方法仍是
// panic("implement me")，所以这里刻意只用 NewDistributedMutex。
//
// 前置：本地 Redis 监听 localhost:6379。
// 运行：go run ./examples/fair
package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Casper-Mars/distrlock/core/fair"
)

const (
	redisAddr    = "localhost:6379"
	lockKey      = "distrlock:example:fair"
	lockDuration = 5 * time.Second
	totalTimeout = 3 * time.Second
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "example fair failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	client := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer func() {
		_ = client.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()

	// 必须用 NewDistributedMutex：fair.NewLocker 的方法全部 panic。
	mutex := fair.NewDistributedMutex(client, lockKey, lockDuration)

	start := time.Now()
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 1; i <= 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mem := fmt.Sprintf("worker-%d", i)
			if err := work(ctx, client, mutex, mem, start); err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			return err
		}
	}

	fmt.Println("2 个 worker 依次获得锁")
	return nil
}

func work(ctx context.Context, client *redis.Client, mutex *fair.DistributedMutex, mem string, start time.Time) error {
	if _, err := mutex.Lock(ctx, mem); err != nil {
		return fmt.Errorf("%s lock: %w", mem, err)
	}
	rank, err := client.ZRank(ctx, lockKey, mem).Result()
	if err != nil {
		return fmt.Errorf("%s rank: %w", mem, err)
	}
	fmt.Printf("%s lock success, rank=%d, 拿到锁耗时=%dms\n", mem, rank, time.Since(start).Milliseconds())

	time.Sleep(500 * time.Millisecond)

	if err := mutex.Unlock(ctx, mem); err != nil {
		return fmt.Errorf("%s unlock: %w", mem, err)
	}
	fmt.Printf("%s unlock ok, 总耗时=%dms\n", mem, time.Since(start).Milliseconds())
	return nil
}
