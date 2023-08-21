package reentrant_lock

import "time"

type Option func(o *lockOptions)

type lockOptions struct {
	// 等待锁的时间，默认1s，如果第一次获取锁失败，则会等待该时间后再次尝试获取锁
	waitTime time.Duration
	// 尝试获取锁的次数
	retry int
}

// WithWaitTime 设置等待时间
func WithWaitTime(waitTime time.Duration) Option {
	return func(o *lockOptions) {
		o.waitTime = waitTime
	}
}

// WithRetry 设置尝试获取锁的次数
func WithRetry(retry int) Option {
	return func(o *lockOptions) {
		o.retry = retry
	}
}
