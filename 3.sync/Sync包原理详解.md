# Sync包的概述与设计哲学

在 Go 语言中，`sync` 包提供了一些强大的同步原语，用于协调多个 goroutine 之间的并发执行。相比于 `channel`，`sync` 包更侧重于**低级别的内存访问同步**，它提供了更精细的控制手段。

本篇文章将详细讲解 `sync` 包中的常用同步原语，包括 `Mutex`（互斥锁）、`RWMutex`（读写锁）、`WaitGroup`、`Once` 和 `Cond`（条件变量），并通过多个案例帮助你理解它们的使用场景和实现原理。

在 Go 语言的并发模型中，`sync` 包提供了内存同步的基本工具。与 `channel` 不同，`sync` 包适用于**内存访问同步**和**状态保护**，它帮助确保多个 goroutine 之间对共享资源的安全访问。

| **机制** | **适用场景** | **设计思想** |
| --- | --- | --- |
| Channel | 数据传递、工作协调 | "通过通信来共享内存" |
| sync包 | 内存访问同步、状态保护 | "通过同步来安全共享内存" |

# Mutex（互斥锁）解析

`Mutex`（互斥锁）是 Go 中最常见的同步原语，它用于确保在同一时刻只有一个 goroutine 能访问共享资源，从而避免了并发问题。

`Mutex` 在 `sync` 包中提供，它通过两个主要的操作来管理锁：

- `Lock()`：用于加锁。当一个 goroutine 调用 `Lock()` 时，如果该锁没有被其他 goroutine 占用，它就会成功地获取锁。如果已经被占用，调用 `Lock()` 的 goroutine 会被阻塞，直到锁被释放。
- `Unlock()`：用于解锁。当一个 goroutine 完成对共享资源的访问后，它调用 `Unlock()` 释放锁，使其他等待的 goroutine 可以获取该锁。

## 基本使用案例

代码实现了一个银行账户的模拟，在多 goroutine 并发操作存款和取款时使用了 `Mutex` 来保护账户余额的访问。

```go
package main

import (
	"fmt"
	"sync"
)

type BankAccount struct {
	balance int
	mu      sync.Mutex
}

func (b *BankAccount) Deposit(amount int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	fmt.Printf("存款前余额: %d, 存款: %d\n", b.balance, amount)
	b.balance += amount
	fmt.Printf("存款后余额: %d\n", b.balance)
}

func (b *BankAccount) Withdraw(amount int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.balance >= amount {
		fmt.Printf("取款前余额: %d, 取款: %d\n", b.balance, amount)
		b.balance -= amount
		fmt.Printf("取款后余额: %d\n", b.balance)
		return true
	}
	return false
}

func (b *BankAccount) Balance() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.balance
}

func main() {
	acct := &BankAccount{balance: 1000}
	var wg sync.WaitGroup

	// 并发存款
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(amount int) {
			defer wg.Done()
			acct.Deposit(amount)
		}(100)
	}

	// 并发取款
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(amount int) {
			defer wg.Done()
			acct.Withdraw(amount)
		}(200)
	}

	wg.Wait()
	fmt.Printf("最终余额: %d\n", acct.Balance())
}

```

说明：

- **`Deposit` 方法**：用于存款，接收存款金额，先加锁，再修改余额，最后解锁。
- **`Withdraw` 方法**：用于取款，接收取款金额，先加锁，检查余额是否足够，若足够则取款，最后解锁。
- **`Balance` 方法**：获取当前余额，获取余额时加锁确保线程安全。

在每个方法中（`Deposit`、`Withdraw` 和 `Balance`），都通过 `b.mu.Lock()` 和 `b.mu.Unlock()` 来确保同一时刻只有一个 goroutine 能访问和修改余额。通过 `defer` 关键字，确保即使函数退出时也能释放锁，从而避免死锁。

试想，如果没有 `Mutex`，在并发存款和取款时，可能会出现多个 goroutine 同时访问和修改 `balance` 的情况，从而导致数据不一致的结果。

输出：

输出的过程是按照 goroutine 的执行顺序生成的，可能会有一些差异，但大体的逻辑和操作顺序是一样的。

```go
存款前余额: 1000, 存款: 100
存款后余额: 1100
存款前余额: 1100, 存款: 100
存款后余额: 1200
存款前余额: 1200, 存款: 100
存款后余额: 1300
取款前余额: 1300, 取款: 200
取款后余额: 1100
存款前余额: 1100, 存款: 100
存款后余额: 1200
取款前余额: 1200, 取款: 200
取款后余额: 1000
存款前余额: 1000, 存款: 100
存款后余额: 1100
取款前余额: 1100, 取款: 200
取款后余额: 900
最终余额: 900
```

如果没有使用 `Mutex`，多个 goroutine 可能会并发修改余额，导致不一致的结果，比如存款可能丢失或者取款时余额不足的判断错误。因此，`Mutex` 在这种情况下非常重要。

## 底层原理解析

### 内部结构解析

在 Go 中，`sync.Mutex` 是通过以下方式实现的：

```go
package sync

// Mutex represents a mutual exclusion lock.
type Mutex struct {
	state int32 // 互斥锁的状态
	sema  uint32 // 信号量，用于 goroutine 阻塞和唤醒
}
```

- **state**：`Mutex` 的状态通常是一个 `int32` 类型，用于表示锁的当前状态。其值为 `0` 表示锁是“解锁”的状态，非零值表示锁被占用。
- **sema**：这是一个信号量，它是一个用于阻塞和唤醒 goroutine 的机制。通过 `sema`，Go 能够将请求锁的 goroutine 阻塞，直到它可以获取锁。

所以，`sync.Mutex` 的实现采用了自旋锁和信号量相结合的方式。具体来说，它有以下的几种状态：

- 锁是“未被占用”的状态，状态值为 `0`，此时任何 goroutine 都可以获取锁。
- 锁是“已占用”的状态，状态值非 `0`，此时其他 goroutine 必须等待，直到当前持有锁的 goroutine 调用 `Unlock()`。
- 如果锁被占用，Go 会通过 `sema` 阻塞等待的 goroutine，直到锁被释放。

### 加解锁过程解析

- 加锁（`Lock()`）过程：
    1. **检查锁的状态**：首先，尝试通过原子操作（`atomic.CompareAndSwapInt32`）来检查锁的状态是否为“未占用”状态（`0`）。如果是，锁定该 mutex，并将其状态修改为“已占用”。
    2. **阻塞等待**：如果状态不是“未占用”，则表明锁已经被其他 goroutine 占用，当前 goroutine 需要等待。在这种情况下，`Mutex` 会通过一个信号量 `sema` 来阻塞当前的 goroutine，直到其他 goroutine 释放锁。
    3. **自旋和阻塞结合**：Go 的 `Mutex` 实现通常采用了“自旋锁”与“阻塞”结合的策略。当 goroutine 发现锁被占用时，首先会尝试进行自旋，看看锁是否能被快速释放。如果自旋失败，goroutine 就会进入阻塞状态，直到可以获取锁。
- 解锁（`Unlock()`）过程：
    1. **释放锁**：当前 goroutine 调用 `Unlock()` 时，会将锁的状态从“已占用”改为“未占用”。
    2. **唤醒其他等待的 goroutine**：当锁释放后，如果有其他 goroutine 正在等待锁，Go 会通过信号量唤醒一个等待的 goroutine，让它获取锁。

## 死锁与饥饿

死锁发生在两个或多个 goroutine 相互等待对方释放锁的情况下。

在 Go 中，`Mutex` 的实现并没有直接避免死锁的机制。死锁是由用户的并发设计问题引起的，而非 `Mutex` 本身的缺陷。

因此，为了避免死锁，我们需要遵循以下原则：

- **锁的顺序**：当多个 goroutine 需要同时获取多个锁时，应确保按照固定的顺序来请求锁，避免循环等待。
- **避免持有锁太久**：**不要在持有锁的情况下进行耗时操作**，如 I/O 操作或者网络请求。
- **使用超时机制**：如果某个 goroutine 长时间未能获取锁，可以考虑实现超时机制，防止死锁的发生。

而饥饿指的是某些 goroutine 长时间无法获取到所需要的资源，尤其是锁，导致这些 goroutine 始终无法执行，甚至可能被永久阻塞。

饥饿通常发生在多个 goroutine 争夺有限的资源时，某些 goroutine 一直得不到资源的分配，导致它们永远无法得到执行机会。

Go 语言中的 `sync.Mutex` 采用的是 **互斥锁**（Mutual Exclusion），它是一个简单的锁机制，主要通过两种操作来控制并发：

- `Lock()`：申请获取锁。
- `Unlock()`：释放锁。

这在前文我们已经知道，但是，`Mutex` 本身并没有内建的机制来优先保证某个 goroutine 能够按顺序获得锁。所以，如果某个 goroutine 获取锁后长时间不释放，而其他 goroutine 在等待时没有合适的机制来公平地轮换锁的占用，可能会导致“饥饿”。以下两种情况是“饥饿”经常出现的：

- 长时间持锁的 goroutine：获取锁后，执行了一些非常耗时的任务
- 优先级不公平：Go 的调度器并没有特别的优先级控制。在某些情况下，频繁请求锁的 goroutine 可能会反复成功获取锁，而等待时间较长的 goroutine 一直被饿死

为了避免 Go 中出现饥饿现象，Go 提供了另一种锁机制 —— `sync.RWMutex`

> 题外话：当然除了 sync.RWMutex，避免饥饿的方式还有实现优先级调度、优化代码避免持有锁的时间过长、设计合理的锁粒度等等。
>

# RWMutex（读写锁）解析

## 什么是 sync.RWMutex

`sync.RWMutex` 是 Go 提供的一个锁，它分为 **读锁** 和 **写锁**。它可以同时允许多个 goroutine 读取共享数据，但只有一个 goroutine 可以对数据进行写操作。这个机制提高了并发性能，尤其是在读操作远远多于写操作的场景中。

我们来看看这两种锁的行为：

- **读锁（`RLock`）**：允许多个 goroutine 同时获得锁，这意味着多个 goroutine 可以同时读取共享数据。当有 goroutine 获得读锁时，其他 goroutine 也可以继续获得读锁，直到没有任何 goroutine 获取写锁。
- **写锁（`Lock`）**：写锁是独占的，意味着在有 goroutine 获得写锁时，所有其他 goroutine（无论是读锁还是写锁）都不能访问共享数据。

## 基本使用案例

在 Go 中，`sync.RWMutex` 主要的方法如下：

- **`RLock()`**：获取读锁，允许多个 goroutine 并发读取数据。
- **`RUnlock()`**：释放读锁。
- **`Lock()`**：获取写锁，只有一个 goroutine 能获取写锁。
- **`Unlock()`**：释放写锁。

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type ConfigManager struct {
	config map[string]string
	rw     sync.RWMutex
}

func (c *ConfigManager) Get(key string) string {
	c.rw.RLock()
	defer c.rw.RUnlock()

	time.Sleep(10 * time.Millisecond) // 这里模拟读取的耗时
	return c.config[key]
}

func (c *ConfigManager) Set(key, value string) {
	c.rw.Lock()
	defer c.rw.Unlock()

	time.Sleep(30 * time.Millisecond) // 这里模拟写入的耗时
	c.config[key] = value
	fmt.Printf("配置更新: %s = %s\n", key, value)
}

func main() {
	cfg := ConfigManager{
		config: map[string]string{"version": "1.0.1", "mode": "debug", "author": "nameless"},
	}

	var wg sync.WaitGroup

	for i := 0; i < 15; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ver := cfg.Get("version")
			author := cfg.Get("author")
			fmt.Printf("Reader %d: version = %s author = %s\n", id, ver, author)
		}(i)
	}

	// 启动一个写 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		cfg.Set("version", "1.0.2")
		cfg.Set("author", "无名客")
	}()

	wg.Wait()
}

```

输出

```go
Reader 0: version = 1.0.1 author = nameless
配置更新: version = 1.0.2
Reader 2: version = 1.0.1 author = nameless
Reader 1: version = 1.0.1 author = nameless
Reader 4: version = 1.0.1 author = nameless
Reader 5: version = 1.0.1 author = nameless
Reader 6: version = 1.0.1 author = nameless
Reader 7: version = 1.0.1 author = nameless
Reader 8: version = 1.0.1 author = nameless
Reader 9: version = 1.0.1 author = nameless
Reader 10: version = 1.0.1 author = nameless
Reader 12: version = 1.0.1 author = nameless
配置更新: author = 无名客
Reader 13: version = 1.0.2 author = 无名客
Reader 11: version = 1.0.2 author = 无名客
Reader 14: version = 1.0.2 author = 无名客
Reader 3: version = 1.0.2 author = 无名客
```

## 什么时候使用 RWMutex？

- **读多写少的场景**：`RWMutex` 特别适用于读操作频繁、写操作较少的场景。通过允许多个 goroutine 同时进行读操作，可以极大地提高并发性能。例如，缓存、配置读取等操作通常是典型的使用场景。
- **避免不必要的锁**：如果读操作和写操作几乎是等量的，或者写操作非常频繁，使用 `RWMutex` 可能会带来额外的开销。这时使用普通的 `Mutex` 更为简单高效。

## 实现原理解析

### 内部原理

`sync.RWMutex` 的底层实现是由 Go 的运行时（runtime）库提供的。在源代码中，`RWMutex` 的结构大致如下：

```go
type RWMutex struct {
    w           Mutex        // 内部互斥锁，用于序列化写者访问
    writerSem   uint32       // 写者信号量：写者等待所有读者完成
    readerSem   uint32       // 读者信号量：读者等待写者完成
    readerCount atomic.Int32 // 读者计数器：正值表示活跃读者数，负值表示写锁持有
    readerWait  atomic.Int32 // 等待读者计数器：写者等待的剩余读者数
}
```

这个结构中包含了几个重要的字段：

- **w (Mutex)**：这是一个标准的 sync.Mutex，用于确保多个写者不会同时竞争锁。它充当写者的“入口关卡”，防止并发写操作。
- **readerCount (atomic.Int32)**：核心计数器，使用原子操作（如 Add、Load、CompareAndSwap）来更新。正常情况下，它记录活跃读者的数量（正整数）。当写锁被获取时，它会被设置为负值（具体为 - (1 << 30) 加剩余读者数），以此信号通知新读者需要等待。
- **readerWait (atomic.Int32)**：辅助计数器，仅在写锁获取时使用。它记录写者需要等待的读者数量（即当前活跃读者数）。每当一个读者释放锁时，这个值会递减；当它降至零时，写者被唤醒。
- **writerSem 和 readerSem (uint32)**：这些是信号量（semaphore）的句柄，由 Go runtime 实现。它们不直接存储计数值，而是用于阻塞和唤醒 goroutine。信号量通过 runtime_Semacquire（等待）和 runtime_Semrelease（释放）操作。

### 读锁（RLock 和 RUnlock）的机制

读锁的设计强调并发性：只要没有写锁，多个读者可以同时进入。

- **`RLock()`**：
    1. 使用原子操作 `readerCount.Add(1)` 递增计数器。
    2. 如果递增后 `readerCount < 0`，说明有写锁持有或等待，新读者必须阻塞：调用 `runtime_SemacquireRWMutexR(&readerSem, ...)`，将 goroutine 置于等待队列，直到写锁释放。
    3. 如果 `>= 0`，立即成功获取读锁，支持并发读取。

  这确保了读者不会在写锁活跃时进入，但允许现有读者继续（除非写锁已信号通知）。

- **`RUnlock()`**：
    1. 原子操作 `readerCount.Add(-1)` 递减计数器。
    2. 如果递减后 `< 0`，说明有写锁等待，进入慢路径 `rUnlockSlow`：原子递减 `readerWait`。
    3. 如果 `readerWait` 降至 0，说明这是最后一个离开的读者，此时释放写者信号量 `runtime_Semrelease(&writerSem, ...)`，唤醒等待的写者。

这个过程避免了不必要的信号量操作，只有在有写者等待时才唤醒，从而优化性能。

### 写锁（Lock 和 Unlock）的机制

写锁强调排他性：它必须等待所有读者离开，并阻止新读者进入。

- **`Lock()`**：
    1. 先获取内部 w 互斥锁，确保只有一个写者能继续（其他写者阻塞在这里）。
    2. 原子操作 `readerCount.Add(-rwmutexMaxReaders)`，将计数器置为负值（信号写锁活跃），并计算当前活跃读者数 `r = readerCount + rwmutexMaxReaders`。
    3. 原子 `readerWait.Add(r)`，设置等待的读者数量。
    4. 如果 `r != 0`（有活跃读者），调用 `runtime_SemacquireRWMutex(&writerSem, ...)` 阻塞，直到所有读者释放并唤醒它。

  这个负值偏移是关键：它让后续 `RLock` 检测到写锁并阻塞，防止新读者介入，从而避免写者饥饿。

- **Unlock()**：
    1. 原子 `readerCount.Add(rwmutexMaxReaders)`，恢复计数器到正值（剩余读者数）。
    2. 对于每个等待的读者（基于之前计算的 r），循环调用 `runtime_Semrelease(&readerSem, ...)` 唤醒它们。
    3. 最后释放内部 w 互斥锁，允许下一个写者尝试。

### 整体协调与性能优化

看完读写锁的机制，我们大概能明白 `RWMutex` 在读多写少的场景下远优于简单 Mutex，因为读者间无互斥开销。只有在争用时，才 fallback 到信号量阻塞。

首先，是 RWMutex 计数器与信号量的协作。计数器（readerCount 和 readerWait）提供快速、原子化的状态检查和更新，用于快路径（无阻塞时）。信号量仅在需要阻塞时介入，用于慢路径的等待和唤醒。这是一种“乐观并发”策略 —— 大多数情况下仅用原子操作，避免昂贵的系统调用。

其次就是这里提到的原子操作了，所有计数器更新都使用 sync/atomic 包，确保无锁并发安全。

然后就是我们在 Mutex 中谈到的饥饿问题，RWMutex 的解决方案就是当写锁请求到来时，它会阻塞新读者（通过负 readerCount），确保现有读者完成后写者能介入。这就解决了经典读者-写者问题中的写者饥饿。

# WaitGroup 机制解析

## WaitGroup 说明

WaitGroup用于等待一组goroutine完成。

## 基本使用案例

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("Worker %d: 开始工作\n", id)
	time.Sleep(time.Duration(id) * time.Second) // 模拟工作时间
	fmt.Printf("Worker %d: 工作完成\n", id)
}

func main() {
	var wg sync.WaitGroup
	const workerCnt = 5

	fmt.Println("启动所有的 worker...")

	for i := 1; i <= workerCnt; i++ {
		wg.Add(1)
		go worker(i, &wg)
	}

	fmt.Println("等待所有worker完成...")
	wg.Wait() // 阻塞直到计数器归零
	fmt.Println("所有worker已完成!")
}
```

输出

```go
启动所有的 worker...
等待所有worker完成...
Worker 5: 开始工作
Worker 1: 开始工作
Worker 2: 开始工作
Worker 4: 开始工作
Worker 3: 开始工作
Worker 1: 工作完成
Worker 2: 工作完成
Worker 3: 工作完成
Worker 4: 工作完成
Worker 5: 工作完成
所有worker已完成!
```

## 相较于之前使用 time.Sleep 等待协程运行完毕，在 WaitGroup 中是怎样做的

## **WaitGroup内部原理**

# Once 实现原理

## Once 的作用

Once确保某个操作只执行一次。

## 基本使用案例

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type DBConn struct {
	url string
}

var (
	instance *DBConn
	once     sync.Once
)

func getDBConn() *DBConn {
	once.Do(func() {
		fmt.Println("创建数据库连接...")
		// 模拟耗时的连接建立
		time.Sleep(2 * time.Second)
		instance = &DBConn{url: "postgres://localhost:5432/mydb"}
		fmt.Println("数据库连接已创建!")
	})
	return instance
}

func main() {
	var wg sync.WaitGroup

	// 多个 goroutine 获取连接
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Goroutine %d: 获取数据库连接...\n", id)
			conn := getDBConn()
			fmt.Printf("Goroutine %d: 连接到 %s\n", id, conn.url)
		}(i)
	}

	wg.Wait()
}
```

输出

```go
Goroutine 0: 获取数据库连接...
Goroutine 4: 获取数据库连接...
创建数据库连接...
Goroutine 2: 获取数据库连接...
Goroutine 1: 获取数据库连接...
Goroutine 3: 获取数据库连接...
数据库连接已创建!
Goroutine 0: 连接到 postgres://localhost:5432/mydb
Goroutine 4: 连接到 postgres://localhost:5432/mydb
Goroutine 3: 连接到 postgres://localhost:5432/mydb
Goroutine 2: 连接到 postgres://localhost:5432/mydb
Goroutine 1: 连接到 postgres://localhost:5432/mydb
```

## **Once内部机制**

# 条件变量 Cond

## Cond 概述

Cond用于在特定条件下等待或通知goroutine。

## 基本使用示例

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

type TaskQueue struct {
	tasks []string
	cond  *sync.Cond
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		tasks: make([]string, 0),
		cond:  sync.NewCond(&sync.Mutex{}),
	}
}

func (q *TaskQueue) Add(task string) {
	q.cond.L.Lock()
	q.tasks = append(q.tasks, task)
	fmt.Printf("添加任务: %s\n", task)
	q.cond.L.Unlock()

	// 通知等待的消费者
	q.cond.Signal() // 通知一个等待者
	// q.cond.Broadcast() // 通知所有等待者
}

func (q *TaskQueue) Get() string {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	// 等待知道有任务可用
	for len(q.tasks) == 0 {
		fmt.Println("队列为空，等待任务...")
		q.cond.Wait() // 释放锁并等待，被唤醒时重新获取锁
	}

	task := q.tasks[0]
	q.tasks = q.tasks[1:]
	return task
}

func main() {
	queue := NewTaskQueue()
	var wg sync.WaitGroup

	// 消费者
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 3; i++ {
			task := queue.Get()
			fmt.Printf("处理任务: %s\n", task)
			time.Sleep(1 * time.Second)
		}
	}()

	// 生产者
	time.Sleep(200 * time.Millisecond) // 这里我们让生产者晚一点生产，让消费者先开始，来看看 Cond 的作用
	for i := 0; i < 3; i++ {
		queue.Add(fmt.Sprintf("task-%d", i))
		time.Sleep(500 * time.Millisecond)
	}

	wg.Wait()
}

```

输出：

```go
队列为空，等待任务...
添加任务: task-0
处理任务: task-0
添加任务: task-1
添加任务: task-2
处理任务: task-1
处理任务: task-2
```

# 同步原语的选择指南

| **场景** | **推荐原语** | **理由** |
| --- | --- | --- |
| 保护共享数据 | `Mutex` | 简单直接的互斥访问 |
| 读多写少 | `RWMutex` | 提高读并发性能 |
| 等待任务组完成 | `WaitGroup` | 简洁的任务协调 |
| 一次性初始化 | `Once` | 确保初始化只执行一次 |
| 复杂条件等待 | `Cond` | 基于条件的goroutine协调 |
| 数据传递 | `Channel` | goroutine间的通信 |

# 最佳实践与常见陷阱

最佳实践

```go
package case6

import "sync"

type StructExample struct {
	mu sync.Mutex
}

// 1. 总是使用 defer 解锁
func (s *StructExample) Method() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 这里操作共享数据
}

// 2. WatiGroup 使用模式
func example(n int) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1) // 在 goroutine 外调用
		go func() {
			// defer 关闭
			defer wg.Done()
			// 工作代码
		}()
	}
	wg.Wait()
}

```

坏例子

```go
package case6

import "sync"

// 陷阱1: 复制包含锁的结构体
type BadStruct struct {
	mu   sync.Mutex
	data int
}

func BadCopy() {
	s1 := BadStruct{}
	s2 := s1 // 复制了锁，会导致未定义行为!
}

// 陷阱2: 重入锁（Go的Mutex不可重入）
type ReentrantExample struct {
	mu sync.Mutex
}

func (r *ReentrantExample) A() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.B() // 死锁! 在A已经持有锁的情况下调用B
}

func (r *ReentrantExample) B() {
	r.mu.Lock() // 这里会阻塞，因为A已经持有锁
	defer r.mu.Unlock()
	// ...
}

```

# 课后思考

## 在什么情况下Mutex会从正常模式切换到饥饿模式？这种设计解决了什么问题？

## RWMutex在大量读操作和偶尔写操作的场景下性能更好，但如果写操作很频繁会怎样？

## 为什么WaitGroup的Add方法要在启动goroutine之前调用，而不是在goroutine内部调用？