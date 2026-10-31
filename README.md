# Distrlock

基于 Redis 的 Go 分布式锁工具包：提供 simple、reentrant、fair 三种锁实现，read-write 锁目前只有占位入口（调用会 panic，见下方矩阵）。

[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://golang.org/)
[![Build Status](https://github.com/Casper-Mars/distrlock/workflows/Build/badge.svg?branch=main)](https://github.com/Casper-Mars/distrlock/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/Casper-Mars/distrlock#1)](https://goreportcard.com/report/github.com/Casper-Mars/distrlock)

**前置条件**：Go 1.20+（见 `go.mod`）、Redis 6+（仓库内示例连接 `localhost:6379`）。

```bash
go get github.com/Casper-Mars/distrlock
```

## 快速开始

```bash
# 1. 起一个本地 Redis；已有可跳过
docker run -d -p 6379:6379 redis:7

# 2. 拿到示例代码
git clone https://github.com/Casper-Mars/distrlock.git
cd distrlock

# 3. 直接跑 simple 示例
go run ./examples/simple
```

期望输出（实际运行结果，退出码 0）：

```text
lock acquired
lock released
```

`examples/simple` 只有两个依赖：本包与 `github.com/redis/go-redis/v9`。想在自己的项目里用，把 `examples/simple/main.go` 抄过去即可。

## ⚠️ 先读：fair 锁有两个入口，其中一个会在调用时 panic

fair 锁对外暴露两个构造，行为完全不同：

| 构造 | 行为 |
|---|---|
| `fair.NewDistributedMutex(...)` | **推荐入口**，`Lock` / `Unlock` 有实现，示例见 `examples/fair` |
| `fair.NewLocker(...)` | ❌ **未实现**：构造能成功，调用 `Lock` / `Unlock` / `TryLock` 任意一个都会 `panic("implement me")` |

`fair.NewLocker` 和 `rw.NewLocker` 一样：构造返回一个编译期完全合法的 `core.Locker`，直到运行到第一个方法调用才崩。它比 `rw` 更容易被误选——同一个 `fair` 包里就放着一个能用的 `NewDistributedMutex`。**本文档推荐的 fair 用法只有 `fair.NewDistributedMutex`。**

## 功能矩阵

| 锁 | 入口 | 状态 |
|---|---|---|
| simple | `simple.NewLocker` | ✅ 可用 |
| reentrant | `reentrant.NewLocker` | ✅ 可用（仅进程内重入） |
| fair | `fair.NewDistributedMutex` | ⚠️ 部分实现：此入口可用 |
| fair | `fair.NewLocker` | ❌ 未实现，**调用即 panic** |
| rw | `rw.NewLocker` | ❌ 未实现，**调用即 panic** |

矩阵与源码的对应关系（行号以当前 `main` 计）：

- `simple.NewLocker`：`core/simple/redis.go:16`，`Lock` / `Unlock` / `TryLock` 三个方法均有真实实现。
- `reentrant.NewLocker`：`core/reentrant/redis.go:24`，三个方法均有实现，重入靠实例内的 `holderCnt` 计数。
- `fair.NewDistributedMutex`：`core/fair/redis.go:51`，返回 `*fair.DistributedMutex`，只有 `Lock` / `Unlock` 两个方法，**没有 `TryLock`**，因此不实现 `core.Locker`，所以标「部分实现」。
- `fair.NewLocker`：`core/fair/redis.go:23`，返回的 `core.Locker` 三个方法体是 `panic("implement me")`（`core/fair/redis.go:32`、`:37`、`:42`）。
- `rw.NewLocker`：`core/rw/redis.go:14`，三个方法体同样是 `panic("implement me")`（`core/rw/redis.go:23`、`:28`、`:33`）。

## 三类锁对照

| | simple | reentrant | fair（`NewDistributedMutex`） |
|---|---|---|---|
| 适用场景 | 单次互斥，抢不到就重试 | 同一逻辑内需要重复加锁（如递归调用） | 多个实例排队，希望先到先得 |
| 是否阻塞 | `Lock` 阻塞（100ms 轮询），`TryLock` 非阻塞 | `Lock` 阻塞（100ms 轮询），`TryLock` 非阻塞 | `Lock` 阻塞（300ms 轮询 `ZRank`），无 `TryLock` |
| 是否带 TTL | 否（`SETNX` 无过期时间） | 否（`SET` 无过期时间） | 是（拿到锁后 `EXPIRE lockDuration`） |
| 跨实例是否成立 | 是，基于 Redis 上同一个 key | 互斥成立；但**重入只在同一实例内生效**，跨实例重入不成立 | 是，基于 Redis ZSET 的 member 排序 |
| 当前状态 | ✅ 可用 | ✅ 可用（仅进程内重入） | ⚠️ 部分实现：此入口可用 |

## 什么时候会死锁

先用三个问题自查，任一命中就要往下读对应小节：

1. 你的锁 key 有没有过期时间？——**没有**的话，进程异常退出就会把锁永久占住（失效模式 1、2）。
2. 你的 `Unlock` 是否可能在「锁已经不属于自己」时执行？——**会**的话就是误释放（失效模式 3）。
3. 你是否指望跨进程「重入」？——**跨进程没有重入**，只有互斥（失效模式 4）。

每条都给出「成因 → 触发条件 → 复现命令 → 观察结果」，全部在本机实跑过。

> **复现口径**：本节所有 `redis-cli` 都用 `-n 15`（专用库），复现程序也连 db15。仓库里 `examples/` 的示例连的是**默认 db0**，两者不要混用。
> 示例拿到锁就立刻释放，来不及演示「进程被强杀」，所以复现需要一个能长时间持锁的最小程序，见下。

### 复现前的准备

把下面的程序存成 `repro/main.go`（约 65 行，只为复现，不属于库代码）：

```go
// repro/main.go —— 仅用于复现 README「什么时候会死锁」小节的失效模式，不属于库代码。
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Casper-Mars/distrlock/core"
	"github.com/Casper-Mars/distrlock/core/reentrant"
	"github.com/Casper-Mars/distrlock/core/simple"
	"github.com/redis/go-redis/v9"
)

// 专用库 db15，与本节 redis-cli -n 15 对应。
var cli = redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 15})

func main() {
	ctx := context.Background()
	action, kind := os.Args[1], os.Args[2]
	key := "distrlock:example:" + kind

	var l core.Locker
	if kind == "reentrant" {
		l = reentrant.NewLocker(cli, key)
	} else {
		l = simple.NewLocker(cli, key)
	}

	switch action {
	case "hold": // 拿到锁后一直持着，等外部 kill -9
		cli.Del(ctx, key) // 从干净起点开始，避免上一节残留的 key 让 Lock 一直阻塞
		if _, err := l.Lock(ctx); err != nil {
			panic(err)
		}
		fmt.Println("locked:", key)
		if ok, _ := l.TryLock(ctx); ok {
			fmt.Println("same instance TryLock -> true (in-process reentrant works)")
		} else {
			fmt.Println("same instance TryLock -> false")
		}
		time.Sleep(5 * time.Minute)
	case "try": // 另一个进程尝试加锁
		ok, err := l.TryLock(ctx)
		fmt.Printf("another process TryLock -> %v, err=%v\n", ok, err)
	case "misrelease": // A 加锁 -> 锁消失 -> B 加锁 -> A 释放
		cli.Del(ctx, key) // 同上，保证 A 能立刻拿到锁
		if _, err := l.Lock(ctx); err != nil {
			panic(err)
		}
		fmt.Println("A locked")
		cli.Del(ctx, key)
		fmt.Println("simulate A lock expired/deleted")
		b := simple.NewLocker(cli, key)
		if _, err := b.Lock(ctx); err != nil {
			panic(err)
		}
		fmt.Println("B locked")
		if err := l.Unlock(ctx); err != nil {
			panic(err)
		}
		fmt.Println("A.Unlock done, B lock still exists:", cli.Exists(ctx, key).Val() == 1)
	default:
		fmt.Fprintln(os.Stderr, "usage: repro <hold|try|misrelease> <simple|reentrant>")
		os.Exit(2)
	}
}
```

```bash
go build -o /tmp/repro-distrlock ./repro
```

### 失效模式 1：simple 的锁永不过期

**成因**：`core/simple/redis.go:50` 是 `SetNX(ctx, key, "locked", 0)`，go-redis 里最后一个参数 `0` 表示**不设过期时间**。

**触发条件**：持锁进程没走到 `Unlock` 就退出——`kill -9`、panic、容器 OOM、机器宕机。

**复现命令**：

```bash
/tmp/repro-distrlock hold simple & HOLD=$!        # 拿锁并持住（最多 5 分钟）
sleep 2
redis-cli -n 15 TTL distrlock:example:simple      # -1 表示没有过期时间
kill -9 "$HOLD"                                   # 模拟进程被强杀
sleep 1
redis-cli -n 15 TTL distrlock:example:simple      # 仍然是 -1
redis-cli -n 15 EXISTS distrlock:example:simple   # 1：锁还在
/tmp/repro-distrlock try simple                   # 另一个进程拿不到
```

**观察结果**：

```text
locked: distrlock:example:simple
same instance TryLock -> false
-1
-1
1
another process TryLock -> false, err=<nil>
```

**这就是死锁**：`TTL` 恒为 `-1`，持锁进程已经死了，key 却不会消失。除非有人手动 `DEL`，这把锁永远被一个不存在的进程占着。

### 失效模式 2：reentrant 的锁同样永不过期

**成因**：`core/reentrant/redis.go:85-97` 的 Lua 脚本执行 `redis.call('SET', key, lock_val)`（`:92`），没有 `EX` / `PX` 参数，写进去的 key 没有过期时间。

**触发条件**：同失效模式 1。PRD 早期只提到 simple，reentrant 同样成立。

**复现命令**：

```bash
/tmp/repro-distrlock hold reentrant & HOLD=$!
sleep 2
redis-cli -n 15 TTL distrlock:example:reentrant      # -1
kill -9 "$HOLD"
sleep 1
redis-cli -n 15 TTL distrlock:example:reentrant      # 仍然是 -1
redis-cli -n 15 EXISTS distrlock:example:reentrant   # 1
```

**观察结果**：

```text
locked: distrlock:example:reentrant
same instance TryLock -> true (in-process reentrant works)
-1
-1
1
```

### 失效模式 3：simple 的 Unlock 不校验持有者（误释放）

**成因**：`core/simple/redis.go:41-47` 的 `Unlock` 只有一句 `l.cli.Del(ctx, l.key)`（`:42`），不看 key 里的值、也不校验是谁加的锁。

**触发条件**：A 的锁已经消失（超时、被人手动删、Redis 重启），B 拿到了同一个 key 的锁；之后 A 的 `Unlock`（重试、`defer`、超时补偿）才执行，**删掉的是 B 的锁**。

**复现命令**（程序按顺序执行：A 加锁 → 模拟锁消失 → B 加锁 → A 释放）：

```bash
/tmp/repro-distrlock misrelease simple
```

**观察结果**：

```text
A locked
simulate A lock expired/deleted
B locked
A.Unlock done, B lock still exists: false
```

**危害**：B 以为自己持有锁，实际的互斥已经没了，两个进程可以同时进入临界区。这种「误释放」比死锁更难排查——不报错、不留残留 key，只是保护悄悄失效。

### 失效模式 4：reentrant 的重入只在进程内成立

**成因**：重入计数 `holderCnt int32` 是 `locker` 的**进程内字段**（`core/reentrant/redis.go:21`；`TryLock` 里 `+1` 见 `:104`，`Unlock` 里 `-1` 见 `:58`）。跨进程时只认 Redis 里存的实例标识 `ca`（`:28` 生成；`TryLock` 在 `:98`、`Unlock` 在 `:71` 把它作为参数传给 Lua 脚本，脚本用 `GET` 到的值与它比对），另一个进程的 `ca` 不同，`TryLock` 直接返回 false。

**触发条件**：期望「同一业务跨 goroutine / 跨服务重入」。跨进程只有互斥，没有重入语义。

**复现命令**：

```bash
/tmp/repro-distrlock hold reentrant & HOLD=$!   # 进程 A：拿锁，并在同一实例上再 TryLock 一次
sleep 2
/tmp/repro-distrlock try reentrant              # 进程 B：另一个实例
kill -9 "$HOLD"
sleep 1
```

**观察结果**：

```text
locked: distrlock:example:reentrant
same instance TryLock -> true (in-process reentrant works)
another process TryLock -> false, err=<nil>
```

同一实例第二次 `TryLock` 成功（`holderCnt` 变成 2），另一个实例拿不到——确认重入仅限进程内。

### 失效模式 5：fair 的 Unlock 不校验 mem（补充）

`core/fair/redis.go:108-115` 的 `Unlock` 只做 `ZRem(ctx, lockKey, mem)`（`:109`）：知道 `mem` 的任何进程都能把别人的排队项删掉，`Lock` 一侧也不校验调用者。

```bash
redis-cli -n 15 ZADD distrlock:example:fair 1 worker-1 2 worker-2
redis-cli -n 15 ZREM distrlock:example:fair worker-1   # 等价于 fair.Unlock(ctx, "worker-1")
redis-cli -n 15 ZRANGE distrlock:example:fair 0 -1 WITHSCORES
```

观察：`worker-1` 被删掉后，`worker-2` 直接排到第一位——删的是谁的排队项，`ZRem` 并不关心。

### 清理

```bash
redis-cli -n 15 --scan --pattern 'distrlock:*' | xargs -r redis-cli -n 15 del
```

如果你的 `xargs` 不支持 `-r`（部分 macOS 自带版本），用等价的循环：

```bash
redis-cli -n 15 --scan --pattern 'distrlock:*' | while read -r k; do redis-cli -n 15 del "$k"; done
```

## 示例

三个示例都只依赖本包和 `github.com/redis/go-redis/v9`：

| 目录 | 命令 | 演示内容 |
|---|---|---|
| [`examples/simple`](examples/simple) | `go run ./examples/simple` | 取锁、打印 `lock acquired`、释放、打印 `lock released` |
| [`examples/reentrant`](examples/reentrant) | `go run ./examples/reentrant` | 同一实例连续 `TryLock` 两次重入，再 `Unlock` 两次释放 |
| [`examples/fair`](examples/fair) | `go run ./examples/fair` | 两个 goroutine 用 `NewDistributedMutex` 排队取锁，打印 `lock success` 与 `rank`；排队位次见 `Lock` 自身日志 |

## 存储后端

- [x] Redis
- [ ] Etcd
- [ ] Zookeeper
