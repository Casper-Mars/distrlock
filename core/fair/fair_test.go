package fair

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v8"
)

func Test_FairLock(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 如果没有设置密码，则为空字符串
		DB:       0,  // 使用默认数据库
	})

	lockKey := "testwqs"
	lockDuration := 5 * time.Second

	mutex := NewDistributedMutex(client, lockKey, lockDuration)

	sg := sync.WaitGroup{}
	n := 4
	sg.Add(n)
	for i := 0; i < n; i++ {
		// 模拟多个协程同时请求锁
		go func() {
			defer sg.Done()
			ctx := context.Background()
			mem, err := mutex.Lock(ctx, "hhh")
			// 模拟耗时操作
			time.Sleep(time.Second * 3)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer mutex.Unlock(ctx, mem)
		}()
	}
	sg.Wait()
}
