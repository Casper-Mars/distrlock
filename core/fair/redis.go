package fair

import (
	"context"

	"github.com/Casper-Mars/distrlock/core"

	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const zaddNxSuccess = 1

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
	//TODO implement me
	panic("implement me")
}

func (l *locker) Unlock(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (l *locker) TryLock(ctx context.Context) (isLocked bool, err error) {
	//TODO implement me
	panic("implement me")
}

type DistributedMutex struct {
	client       *redis.Client
	lockKey      string
	lockDuration time.Duration
}

func NewDistributedMutex(client *redis.Client, lockKey string, lockDuration time.Duration) *DistributedMutex {
	return &DistributedMutex{
		client:       client,
		lockKey:      lockKey,
		lockDuration: lockDuration,
	}
}

// unlock时需要mem
func (m *DistributedMutex) Lock(ctx context.Context, mems ...string) (mem string, err error) {
	// mem使用uuid，score使用时间戳，确保有序
	if len(mems) <= 0 {
		mem = uuid.New().String()
	} else {
		mem = mems[0]
	}

	score := time.Now().UnixMicro()

	// 尝试获取锁，并排队等待
	result, err := m.client.ZAddNX(ctx, m.lockKey, &redis.Z{
		Score:  float64(score),
		Member: mem,
	}).Result()
	// 发生错误 或者 该成员已存在
	if err != nil || result != zaddNxSuccess {
		return "", fmt.Errorf("failed to ZAddNX lock, key:%s, mem:%s, result:%d, err: %w", m.lockKey, mem, result, err)
	}

	firstLog := true
	// 阻塞等待 前面排队的锁完成
	for {
		rank, err := m.client.ZRank(ctx, m.lockKey, mem).Result()
		// key超时被移除，此处为redis nil 也应return 避免该时刻多个key同时获得锁
		if err != nil {
			return "", fmt.Errorf("failed to ZRank lock, key:%s, mem:%s, err: %w", m.lockKey, mem, err)
		}
		if firstLog {
			log.Printf("try lock, key:%s, mem:%s, score:%d, rank:%d", m.lockKey, mem, score, rank)
			firstLog = false
		}
		if rank == 0 {
			log.Printf("lock success, key:%s, mem:%s", m.lockKey, mem)
			break
		}
		time.Sleep(300 * time.Millisecond) // 等待300毫秒后重试
	}

	// 设置锁的过期时间，防止其他请求长时间等待
	err = m.client.Expire(ctx, m.lockKey, m.lockDuration).Err()
	if err != nil {
		return "", fmt.Errorf("failed to acquire lock, key:%s, mem:%s, err: %w", m.lockKey, mem, err)
	}

	return mem, nil
}

func (m *DistributedMutex) Unlock(ctx context.Context, mem string) error {
	_, err := m.client.ZRem(ctx, m.lockKey, mem).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock, key:%s, mem:%s, err: %w", m.lockKey, mem, err)
	}
	log.Printf("Unlock finish, key:%s, mem:%s\n", m.lockKey, mem)
	return nil
}
