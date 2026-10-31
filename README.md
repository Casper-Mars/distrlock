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
