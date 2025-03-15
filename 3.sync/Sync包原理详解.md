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

在 Go 中，**`sync.WaitGroup`** 是一个用于同步等待多个 goroutine 完成的工具。它常常用来协调一组并发任务的执行，确保所有 goroutine 执行完毕后才继续执行主程序。`WaitGroup` 是一个高效、轻量级的并发工具，广泛应用于需要等待多个并发任务完成的场景中。

`WaitGroup` 结构体定义在 `sync` 包中，简单来说，它主要由以下几个方法组成：

- **`Add(delta int)`**：修改计数器的值。`delta` 可以是正数或负数，通常用于增加 goroutine 数量。
- **`Done()`**：将计数器减一，通常在每个 goroutine 执行完毕时调用。
- **`Wait()`**：阻塞当前线程直到计数器的值减为 0，表示所有 goroutine 都已经完成。

## 基本使用案例

让我们从一个简单的例子开始，了解如何使用 `WaitGroup` 等待多个 goroutine 执行完毕：

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

- `main` 函数中我们创建了一个 `WaitGroup` 对象 `wg`。
- 在启动每个 goroutine 前，调用 `wg.Add(1)` 来增加等待计数，表示我们正在等待一个 goroutine 完成。
- 每个 `worker` 函数在完成工作后调用 `wg.Done()`，减小等待计数。
- 最后，`wg.Wait()` 会阻塞直到所有的 goroutine 完成，计数器归零。

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

## 一些 WaitGroup 的注意事项

1. `Add` 必须在 `go` 语句之前调用

   `WaitGroup` 的计数器是用于跟踪未完成的 goroutine 数量。因此，我们必须确保在每个 goroutine 启动之前调用 `wg.Add(1)`。如果在启动 goroutine 后再调用 `Add(1)`，会导致 `WaitGroup` 在调用 `Wait()` 时阻塞，造成死锁。

2. 不要重复调用 `Done`

   `Done` 是用来减少计数器的，所以每个 goroutine 只能调用一次 `Done`。如果重复调用 `Done()`，会导致程序 panic。最佳实践就是在 go 函数中使用 defer 确保调用了一次 `Done`

    ```go
    go func() {
        defer wg.Done()  // 正确：调用一次 Done
        // goroutine 执行的代码
    }()
    ```

3. `Wait` 必须被调用一次

   在程序中至少有一个地方调用 `wg.Wait()`，否则主程序可能会提前退出，导致 goroutine 没有足够时间执行完成。


## 相较于之前使用 time.Sleep 等待协程运行完毕，在 WaitGroup 中是怎样做的

我们之前会使用 `time.Sleep` 来等待所有 goroutine 完成，尽管这种方法可以达到目的，但它存在一定的问题和局限性。相比之下，`sync.WaitGroup` 提供了一种更优雅、可靠且高效的方式来等待 goroutine 完成。

使用 `time.Sleep` 来等待 goroutine 完成似乎很简单，但它有几个明显的问题：

- **不精确**：`time.Sleep` 需要预估所有 goroutine 最长的执行时间。比如，我们假设最慢的 goroutine 会运行 5 秒，因此在 `main` 中使用了 `time.Sleep(6 * time.Second)` 来等待。但是，如果某些 goroutine 执行的时间远远小于 5 秒，使用 6 秒的等待时间就会显得不必要，浪费了执行时间。
- **不可靠**：如果你无法准确预测 goroutine 执行的最大时间，或者某个 goroutine 的执行时间发生变化，可能会导致等待时间不足，导致主程序提前退出，而 goroutine 仍未完成。这会导致程序的不确定性，甚至可能出现数据不一致的情况。
- **不灵活**：如果我们有更多的 goroutine，或者每个 goroutine 的执行时间变化不确定，那么 `time.Sleep` 就变得不那么灵活和可维护了。每次修改 goroutine 数量或执行时间时，都需要手动调整 `time.Sleep` 的等待时间。

使用 `sync.WaitGroup` 来等待 goroutine 完成，不仅更加精确，而且更具可维护性和灵活性。`WaitGroup` 会准确地等待每个 goroutine 完成，而不需要预估执行时间。主程序通过 `wg.Add(1)` 增加计数，等到所有的 goroutine 都调用 `wg.Done()`，计数器归零时，`wg.Wait()` 才会解锁，程序继续执行。这样，你不需要关心具体的时间长度，程序会自动等待所有 goroutine 完成。

`WaitGroup` 使得程序可以完全不依赖于时间，避免了因为人为估算 `time.Sleep` 的等待时间而带来的错误。它会根据 goroutine 的执行情况自动完成同步，避免了人为错误的干扰。

而且，即使你改变了 goroutine 的数量或者每个 goroutine 的执行时间，`WaitGroup` 也会自动调整等待的时机，确保程序的健壮性和可靠性。这使得你的程序更加灵活和可维护。

更重要的是，使用 `time.Sleep` 时，开发者必须手动设置合适的时间，而 `WaitGroup` 在底层通过计数器自动管理 goroutine 的同步，减少了因等待时间设置不当而导致的死锁或提前退出等问题。

通过 `WaitGroup`，我们能够有效地管理多个 goroutine 的并发执行，确保主程序在所有 goroutine 执行完毕后才继续执行。`WaitGroup` 是 Go 并发编程中的标准工具，现在，我们

## **WaitGroup内部原理**

### 内部数据结构

WaitGroup 的结构体设计简洁高效，强调原子操作和信号量（semaphore）以实现无锁（lock-free）或低争用同步。

```go
type WaitGroup struct {
    noCopy noCopy       // 防止拷贝，确保线程安全（通过 vet 工具静态检查）
    state  atomic.Uint64 // 64 位原子状态字段，编码计数器和等待者数量
    sema   uint32        // 信号量，用于阻塞和唤醒等待的 goroutine
}
```

state 是核心状态字段，使用 64 位原子整数存储两个关键值，避免单独字段的额外开销

- 高 32 位（bits 63-32）：任务计数器（counter），类型为 int32，表示剩余待完成的任务数。初始为 0，支持正值（Add 增加）和负值检查（防止负计数）。
- 低 32 位（bits 31-0）：等待者计数（waiter count），类型为 uint32，表示当前调用 Wait 并阻塞的 goroutine 数量。其中，bit 32（实际是低 32 位的 bit 0？源代码中是 bit 32 为 synctest bubble flag，用于测试框架）。

通过位移和掩码提取：`counter = int32(state >> 32)`，`waiters = uint32(state & 0x7fffffff)`。

WaitGroup 的三个主要方法（Add、Done、Wait）围绕 state 和 sema 协作。它们使用乐观并发控制：优先原子操作，快路径无阻塞；慢路径涉及信号量。

【TODO】添加 Add、Done、Wait 三个方法的实现细节

# Once 实现原理

## Once 的作用

在并发编程中，有时我们需要确保某个操作只会执行一次。无论多少 goroutine 试图执行该操作，我们都希望它仅执行一次，这种需求在很多情况下非常重要，特别是在单例模式（Singleton）设计中。Go 提供了 `sync.Once` 来解决这一问题。

`sync.Once` 是 Go 语言中的一个同步原语，它保证所执行的某个操作在程序生命周期内仅会执行一次。即便有多个 goroutine 尝试执行相同的操作，`Once` 也会确保该操作只会执行一次。

Once 的常见场景有：

- **单例模式（Singleton）**：确保某个资源（如数据库连接、配置文件等）只创建一次。
- **延迟初始化**：在并发环境下，确保某些资源的初始化操作是线程安全的，并且只执行一次。
- **并发任务执行前的初始化**：比如只初始化一次日志记录系统。

## 基本使用案例

单例模式的目标是确保某个类只有一个实例，并提供一个全局访问点。在多 goroutine 环境下，我们通常会使用 `sync.Once` 来实现。

以下是一个典型的使用 `sync.Once` 实现单例模式的例子，确保数据库连接只创建一次。在数据库连接管理中，我们希望在整个程序生命周期内，数据库连接只被创建一次。即使多个 goroutine 同时调用 `getDBConn()` 方法，数据库连接也只会创建一次。

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

确保某些操作只执行一次是常见需求，Once 结构正是为此而生：它提供了一种线程安全的机制，保证关联的函数只被调用一次，无论有多少 goroutine 尝试触发它。

```go
type Once struct {
	_ noCopy

	// done indicates whether the action has been performed.
	// It is first in the struct because it is used in the hot path.
	// The hot path is inlined at every call site.
	// Placing done first allows more compact instructions on some architectures (amd64/386),
	// and fewer instructions (to calculate offset) on other architectures.
	done atomic.Bool
	m    Mutex
}
```

- **done (atomic.Bool)**：状态标志，初始为 false，表示函数尚未执行；设置为 true 后，表示已完成。
  - 注：在较早版本（如 Go 1.21 及之前），done 字段使用 atomic.Uint32；而在 Go 1.22 及后续版本中，引入更语义化的 atomic.Bool 类型，以提升代码可读性和类型安全。这种变化不影响核心逻辑，只是优化了原子操作的表达方式。
- **m (Mutex)**：嵌入的 sync.Mutex，仅在首次执行或争用时使用。它确保只有一个 goroutine 执行函数，其他等待者阻塞在此。

```go
func (o *Once) Do(f func()) {
	// Note: Here is an incorrect implementation of Do:
	//
	//	if o.done.CompareAndSwap(0, 1) {
	//		f()
	//	}
	//
	// Do guarantees that when it returns, f has finished.
	// This implementation would not implement that guarantee:
	// given two simultaneous calls, the winner of the cas would
	// call f, and the second would return immediately, without
	// waiting for the first's call to f to complete.
	// This is why the slow path falls back to a mutex, and why
	// the o.done.Store must be delayed until after f returns.

	if !o.done.Load() {
		// Outlined slow-path to allow inlining of the fast-path.
		o.doSlow(f)
	}
}
```

`Do(f func())` 是 Once 的唯一公共方法，其逻辑采用优化后的双重检查锁定（double-checked locking）模式，确保高效性和安全性。源码精简，但蕴含深意：

`Do` 方法的核心职责是确保传入的函数 `f` 仅执行一次。为了解决并发冲突和保证正确性，Go 在 `Do` 方法中设计了一个慢路径（**slow path**）。如果 `o.done.Load()` 返回 false，说明 `f()` 尚未执行，Go 就会调用 `o.doSlow(f)` 来处理执行。

```go
func (o *Once) doSlow(f func()) {
	o.m.Lock()  // 获取锁，保证只有一个 goroutine 执行 f()
	defer o.m.Unlock()  // 释放锁，避免死锁
	if !o.done.Load() {  // 再次检查 done 标志
		defer o.done.Store(true) // 设置 done 为 true，表示已执行过 f()
		f() // 执行函数 f
	}
}
```

# 条件变量 Cond

## Cond 概述

Go 提供了 `sync.Cond`，即条件变量，用于在特定条件下通知和等待 goroutine 的执行。条件变量通过等待和通知机制有效地控制 goroutine 的执行顺序，避免了无谓的轮询和资源浪费。

它允许一个 goroutine 等待某个条件的满足，而另一个 goroutine 可以通过发信号的方式通知等待的 goroutine 条件已满足，继续执行。

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

Add 方法的作用是生产者向任务队列中添加任务，并通过 `Signal` 唤醒一个等待中的消费者。`Signal` 方法会通知一个 goroutine 继续执行，如果有多个消费者，它会随机唤醒一个。

Get 方法中，消费者通过 `Wait` 方法等待队列中有任务可取。`Wait` 会释放当前 goroutine 持有的锁，直到条件满足（即任务队列不为空），然后重新获取锁并执行任务。

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

## Cond的优势

**避免轮询**：传统的轮询方法会不断检查条件，浪费 CPU 资源。`sync.Cond` 通过阻塞和唤醒机制，避免了这种浪费。

**高效的等待通知机制**：通过 `Wait` 和 `Signal` / `Broadcast`，Go 程序可以高效地控制 goroutine 之间的协调，只有条件满足时，goroutine 才会继续执行。

**简化同步逻辑**：在复杂的并发模型中，`sync.Cond` 可以帮助简化同步逻辑，避免显式使用锁和条件变量的复杂代码。

常见的应用场景包括：

- **生产者-消费者模式**：一个或多个 goroutine 生产任务，另一些 goroutine 消费任务。消费者在没有任务时等待，生产者在有任务时通知消费者。
- **工作池模型**：多个工作 goroutine 从一个任务队列中获取任务，当队列为空时，它们进入等待状态，直到有新任务加入。

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

先说说两种模式的核心差异——关键是**锁的抢占规则**

| **模式** | **核心规则（抢占方式）** | **优点** | **缺点** |
| --- | --- | --- | --- |
| 正常模式（默认） | 「非公平抢占」：新到达的 goroutine 可以「插队」抢锁（直接 CAS 尝试获取），等待队列中的 goroutine 需排队竞争 | 吞吐率极高（减少上下文切换） | 可能导致「饥饿」：等待久的 goroutine 一直抢不到锁 |
| 饥饿模式 | 「公平排队」：锁只能交给「等待队列队首」的 goroutine，新到达的 goroutine 不能插队，只能加入队尾等待 | 保证公平性，避免饥饿 | 吞吐率下降（无法插队，需严格排队） |

当且仅当以下两个条件同时满足时，Mutex 会自动从正常模式切换为饥饿模式：

1. **等待队列中有 goroutine 等待锁的时间超过 1 毫秒**（这是核心触发阈值，由 Go 源码硬编码）；
2. **当前持有锁的 goroutine 是等待队列中的“最后一个节点”，或等待队列长度 ≥ 1**（确保有 goroutine 在长期等待）。

通俗来讲：

- 多个 goroutine 竞争锁，其中一个 goroutine A 因为一直被新到的 goroutine 插队，等待时间超过了 1ms；
- 此时，当当前持有锁的 goroutine 释放锁时，会检测到有长期等待的 goroutine A；
- 为了避免 A 继续饥饿，Mutex 会切换到饥饿模式；
- 后续新到达的 goroutine 不能再插队，只能乖乖加入等待队列尾部，锁会优先交给队首的 A（以及后续排队的 goroutine）。

在高并发场景下，锁的公平性是一个至关重要的设计考量。传统的互斥锁在高竞争环境中容易导致"长尾等待"问题——某些 Goroutine 可能因为竞争激烈而长期甚至永远无法获得锁，从而形成事实上的"饿死"现象。这种不公平性不仅影响系统整体的吞吐量，更会导致用户体验的极端分化。通过实现公平的锁机制，我们能够确保每个等待的 Goroutine 都有平等的机会获取锁，从而避免某些请求被无限期推迟。这种公平性设计在实际系统中尤为重要，比如数据库查询和缓存系统等场景，它能保证所有用户请求都能得到及时处理，而不是让少数请求永远处于等待状态，从而构建更加健壮和可靠的分布式系统。

## RWMutex在大量读操作和偶尔写操作的场景下性能更好，但如果写操作很频繁会怎样？

当写操作**非常频繁**时，`RWMutex` 的性能会**显著下降**，甚至可能不如普通的 `Mutex`（互斥锁），核心原因是其「读写互斥」的设计在写密集场景下会放大阻塞开销，失去原本“读并发”的优势。

要想解决写频繁场景的问题，不建议用 `RWMutex`，可以选择以下方案：

1. 改换 Mutex
2. 优化写操作的持有锁的时间
3. 读写分离架构（分布式场景， 读请求走从库（多个从库支持并发读），写请求走主库（主从同步数据），彻底避免单机内的读写锁竞争。）

## 为什么WaitGroup的Add方法要在启动goroutine之前调用，而不是在goroutine内部调用？

回顾一下前文提到的 WaitGroup 核心工作逻辑：

1. `Add(n)`：给计数器加 `n`（告诉 WaitGroup 要等待 `n` 个 goroutine 完成）；
2. `Done()`：给计数器减 `1`（goroutine 完成后调用，通知 WaitGroup 完成了）；
3. `Wait()`：主 goroutine 阻塞，直到计数器减为 `0`（所有等待的 goroutine 都完成）。

所以，**`Wait()` 必须要知道所有 `Add()` 调用的计数**，否则会导致同步失效。

我们之前刚刚学习 goroutine 的时候，知道主 goroutine 和子 goroutine 的执行顺序没有任何保障，所以我们常常需要在主 goroutine 中加一个较长的 sleep。wait 也是同理，如果 Add 写在子 goroutine 内，主 goroutine 先执行 `Wait()`，子 goroutine 还没来得及执行 `Add()`，导致同步失效。而在启动前就 Add，子 goroutine 即使启动慢，也不影响计数器的初始值，`Wait()` 始终能“看到”需要等待的 goroutine 总数。

# 练习题

## 练习题1：顺序打印数字

要求：编写一个程序，启动两个 goroutine，分别打印 "A" 和 "B"，要求它们交替打印，总共打印 5 轮，可能的输出如下：

```go
A
B
A
B
A
B
A
B
A
B
```

`提示：使用两个 channel 实现交替控制。`

### 解法1：标准的使用两个 channel 实现交替控制

```go
func ans1() {
	// 创建两个控制用的 channel
	chA := make(chan bool)
	chB := make(chan bool)

	var wg sync.WaitGroup

	// 添加两个协程
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < round; i++ {
			<-chA // 等待 A 的打印信号
			fmt.Println("A")
			chB <- true // 通知 B 进行打印
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < round; i++ {
			<-chB
			fmt.Println("B")

			// 注意！最后一次打印后不需要再通知 A，此时 A 已经打印完了，再发送就会阻塞了
			if i < 4 {
				chA <- true
			}
		}
	}()

	// 发送开始打印 A 的信号
	chA <- true

	wg.Wait()
}
```

### 解法2：使用一个 channel 实现的简介做法

```go
func ans2() {
	ch := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < round; i++ {
			// 等待从 channel 接收信号
			// 这会阻塞直到另一个 goroutine 发送数据
			<-ch

			fmt.Println("A")

			// 向 channel 发送信号，通知另一个 goroutine 可以继续
			ch <- true
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < round; i++ {
			// 向 channel 发送信号，通知第一个 goroutine 可以开始
			// 注意：这个 goroutine 先发送信号，所以它控制着启动流程
			ch <- true

			fmt.Println("B")

			// 等待从 channel 接收信号
			// 这会阻塞直到第一个 goroutine 发送数据
			<-ch
		}
	}()

	wg.Wait()
}
```

过程详解：

1. 初始化
  - 创建了一个无缓冲 channel `ch`
  - 创建了 `WaitGroup` 并设置为等待 2 个 goroutine
  - 启动了两个 goroutine
2. 执行顺序说明
  - 先看首轮，第 1 轮循环：
    - goroutine B 执行 ch <- true（发送信号）
    - goroutine A 执行 <-ch（接收信号，解除阻塞）
    - goroutine A 打印 "A"
    - goroutine A 执行 ch <- true（发送信号）
    - goroutine B 执行 <-ch（接收信号，解除阻塞）
    - goroutine B 打印 "B"
  - 后续循环，重复上述步骤 2 ~ 6 共 4 次(注意，这里并不是严格按照 A 开始打印的，有可能先打印 B)

这里利用的是两个 goroutine 的操作顺序不同：

- goroutine A 的操作顺序：`接收 → 打印 → 发送`
- goroutine B 的操作顺序：`发送 → 打印 → 接收`

回顾本节的知识，无缓冲 channel 的特点是：**发送操作会阻塞直到有接收操作准备好，反之亦然**

我们利用了这个特性来确保两个 goroutine 严格交替执行，每个 goroutine 在完成自己的工作后，都会等待对方的信号才能继续，这种精确同步的通信模式也保证了无死锁风险

### 解法3：select + context 实现更标准的控制流程

涉及后续的知识，这里只做了解即可

```go
func ans3() {
	chA := make(chan bool, 1)
	chB := make(chan bool, 1)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		count := 0
		for {
			select {
			case <-chA:
				fmt.Println("A")
				count++
				chB <- true // 无论是否到次数，都给B发信号（确保B能处理最后一次）
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		count := 0
		for {
			select {
			case <-chB:
				fmt.Println("B")
				count++
				if count < round {
					chA <- true // 没到次数，继续给A发信号
				} else {
					cancel() // B完成最后一次打印，终止流程
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	chA <- true

	select {
	case <-time.After(2 * time.Second):
		fmt.Println("Timeout!")
	case <-ctx.Done():
		fmt.Println("Completed!")
	}

	wg.Wait()
}
```

## 练习题2：生产者-消费者模型(使用有缓冲通道)

要求：

创建一个生产者 goroutine，它向 channel 中发送数字 1 到 10；
创建一个消费者 goroutine，从该 channel 中读取并打印这些数字。

程序应在所有数字处理完毕后正常退出。

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	ch := make(chan int, 3)

	var wg sync.WaitGroup
	wg.Add(2)

	// 生产者
	go func() {
		defer wg.Done()
		defer close(ch)

		fmt.Println("生产者开始生产数字 1 - 10...")

		for i := 1; i <= 10; i++ {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("发送数字 %d\n", i)
			ch <- i
		}
	}()

	// 消费者
	go func() {
		defer wg.Done()

		fmt.Println("消费者开始接收数字...")

		time.Sleep(300 * time.Millisecond)
		for num := range ch {
			fmt.Printf("收到数字：%d\n", num)
			time.Sleep(300 * time.Millisecond)
		}

		fmt.Println("消费者已收到全部数字")
	}()

	wg.Wait()
}
```

主要是会使用 for 遍历 channel 即可

## 练习题3：多个 worker 并发处理任务

要求：

有一个任务列表 `[]int{2, 4, 6, 8, 10}`，每个任务表示需要“处理”的数值（比如计算其平方）。

启动若干个 worker goroutine（例如 3 个），从一个任务 channel 中读取任务，并将结果（平方值）发送到结果 channel。

主 goroutine 收集所有结果并打印。

思路：

1. 本质上这道题是一个基础的工作池实现，关键在于解决以下问题：
  1. 如何让 3 个 worker 安全共享任务、互不干扰？→ 用「任务通道」统一分发任务
  2. 如何收集所有 worker 的结果？→ 用「结果通道」统一接收结果
  3. 如何确保主 goroutine 不提前退出，且能拿到所有结果？→ 用 `sync.WaitGroup` 等待所有 worker 完成，再关闭结果通道
2. 回答了第一步的3个问题，代码就容易写了
  1. 初始化，创建任务通道、结果通道，以及用于同步 worker 的 WaitGroup。
  2. worker函数，执行工作，主要就是平方
  3. 启动 worker，每个 worker 循环从任务通道拿任务
  4. 提交任务，主 goroutine 将任务列表的数值逐一发送到任务通道，发送完后关闭任务通道（告诉 worker 没有新任务了）
  5. 收集结果，主 goroutine 等待所有 worker 完成任务后，关闭结果通道，再遍历结果通道打印所有结果。

```go
package main

import (
	"fmt"
	"sync"
)

func worker(id int, taskCh <-chan int, resCh chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()

	for num := range taskCh {
		square := num * num
		fmt.Printf("worker-%d: 处理任务 %d, 计算平方值为 %d\n", id, num, square)
		resCh <- square
	}
}

func main() {
	tasks := []int{2, 4, 6, 8, 10}
	workCnt := 3
	taskCh := make(chan int, len(tasks))
	resCh := make(chan int, len(tasks))
	var wg sync.WaitGroup

	wg.Add(workCnt)
	for i := 1; i <= workCnt; i++ {
		go worker(i, taskCh, resCh, &wg)
	}

	for _, task := range tasks {
		taskCh <- task
	}

	close(taskCh)

	wg.Wait()

	close(resCh)

	fmt.Println("所有结果：")
	for res := range resCh {
		fmt.Printf("%d ", res)
	}
}
```

注意，本题没有要求输出顺序，只要求结果，如果想让结果顺序和任务顺序一致（比如输入 [2,4,6,8,10]，输出 [4,16,36,64,100]），该怎么改？

```go
package main

import (
	"fmt"
	"sync"
)

type Task struct {
	Index, Num int
}

type Result struct {
	Index, Square int
}

func worker(id int, taskCh <-chan Task, resCh chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range taskCh {
		square := task.Num * task.Num
		fmt.Printf("worker-%d: 处理任务 %d(值为 %d), 计算平方值为 %d\n", id, task.Index, task.Num, square)
		resCh <- Result{Index: task.Index, Square: square}
	}
}

func main() {
	tasks := []int{2, 4, 6, 8, 10}
	workCnt := 3
	taskCh := make(chan Task, len(tasks))
	resCh := make(chan Result, len(tasks))
	var wg sync.WaitGroup

	wg.Add(workCnt)
	for i := 1; i <= workCnt; i++ {
		go worker(i, taskCh, resCh, &wg)
	}

	for id, task := range tasks {
		taskCh <- Task{Index: id, Num: task}
	}

	close(taskCh)

	wg.Wait()

	close(resCh)

	fmt.Println("所有结果：")
	resSlice := make([]int, len(resCh))
	for res := range resCh {
		resSlice[res.Index] = res.Square
	}
	fmt.Println(resSlice)
}

```

## 练习题4：线程安全的计数器

要求：

实现一个线程安全的计数器，启动 10 个 goroutine，每个 goroutine 对计数器执行 10 次 +1 操作。
使用 sync.Mutex 保护共享变量，最终输出计数器值应为 100。

输出示例： `Final counter: 100`

思路：

考察的是用 `sync.Mutex` 解决「并发写冲突」，首先计数器是多个 goroutine 共享的变量，在每个 goroutine 循环 10 次加一操作时，每次操作都应该加锁，操作过后再解锁

```go
package main

import (
	"fmt"
	"sync"
)

const (
	round        = 10
	goroutineCnt = 10
)

type Counter struct {
	count int
	mu    sync.Mutex
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) AddOne() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.count++
}

func worker(id int, wg *sync.WaitGroup, counter *Counter) {
	defer wg.Done()

	fmt.Printf("worker-%d 开始执行 %d 轮加1操作\n", id, round)
	for i := 0; i < round; i++ {
		counter.AddOne()
	}
}

func main() {
	counter := NewCounter()
	var wg sync.WaitGroup
	wg.Add(goroutineCnt)
	for i := 1; i <= goroutineCnt; i++ {
		go worker(i, &wg, counter)
	}
	wg.Wait()
	fmt.Printf("Final counter: %d\n", counter.count)
}

```

## 练习题5：多协程日志统计

要求：

1. 日志存储：用 []string 切片存储日志内容，同时维护一个 int 记录日志总条数（两个都是共享资源）；
2. 写日志协程：启动 5 个「日志写入协程」，每个协程循环 8 次，每次生成一条日志（格式："goroutine-X: 第Y条日志"，比如 "goroutine-1: 第3条日志"），将日志添加到切片中，并更新 logCount；
3. 读统计协程：启动 3 个「日志统计协程」，每个协程循环 4 次，每次等待 100 毫秒后，读取当前日志总条数 logCount 和最新一条日志内容，打印统计结果（格式："统计协程A：当前日志总数=Z，最新日志=XXX"）；
4. 并发安全：使用 sync.RWMutex 保护共享资源，必须体现「多个读协程可同时执行，写协程执行时会阻塞所有读 / 写协程」；
5. 主协程：等待所有写协程和读协程执行完毕后，打印「所有日志条数：XXX」并退出。

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	readRound  = 4
	readGoCnt  = 3
	writeRound = 8
	writeGoCnt = 5
)

type Logger struct {
	logs  []string
	count int
	rwmu  sync.RWMutex
	wg    sync.WaitGroup
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) statLog(goID int) {
	defer l.wg.Done()

	for i := 1; i <= readRound; i++ {
		time.Sleep(100 * time.Millisecond)
		l.rwmu.RLock()
		logCnt := l.count
		latestLog := l.logs[logCnt-1]
		fmt.Printf("统计协程-%d: 当前日志总数: %d, 最新日志记录为: %s\n", goID, logCnt, latestLog)
		l.rwmu.RUnlock()
	}
}

func (l *Logger) writeLog(goID int) {
	defer l.wg.Done()

	for i := 1; i <= writeRound; i++ {
		l.rwmu.Lock()
		logMsg := fmt.Sprintf("goroutine-%d: 第 %d 条日志", goID, i)
		l.logs = append(l.logs, logMsg)
		l.count++
		l.rwmu.Unlock()
		time.Sleep(50 * time.Millisecond) // 模拟日志生成耗时
	}
}

func main() {
	logger := NewLogger()

	logger.wg.Add(writeGoCnt)
	for i := 1; i <= writeGoCnt; i++ {
		go logger.writeLog(i)
	}

	logger.wg.Add(readGoCnt)
	for i := 1; i <= readGoCnt; i++ {
		go logger.statLog(i)
	}

	logger.wg.Wait()

	// 最终统计
	logger.rwmu.RLock()
	defer logger.rwmu.RUnlock()

	fmt.Printf("所有日志条数: %d\n", logger.count)
}
```

## 练习题6：

1. 无论多少 goroutine 同时调用 `Load()`，初始化逻辑仅执行一次；
2. 所有调用返回的 `Config` 实例地址完全一致，是真正的单例；
3. `Name` 字段始终为 "Default"，初始化结果正确。

```go
package main

import (
	"fmt"
	"sync"
)

type Config struct {
	Name string
}

var (
	instance *Config
	once     sync.Once
)

func Load() *Config {
	once.Do(func() {
		fmt.Println("初始化配置...")
		instance = &Config{Name: "Default"}
	})
	return instance
}

func main() {
	var wg sync.WaitGroup

	wg.Add(10)
	for i := 1; i <= 10; i++ {
		go func(goID int) {
			defer wg.Done()

			for j := 1; j <= 3; j++ {
				cfg := Load()
				fmt.Printf("goroutine-%d 第%d次调用：实例地址=%p，Name=%s\n", goID, j, cfg, cfg.Name)
			}
		}(i)
	}

	wg.Wait()
}

```

Once 的这道题和上文 Once 的例子相似，比较简单，但是 Once 还是需要重点避坑一些错误：

1. 把 Once 当局部变量

    ```go
    func Load() *Config {
        var once sync.Once
        once.Do(...)
        return instance
    }
    ```

   **错误！每个 Load() 调用会创建新的 once，导致多次初始化**

2. 使用 flag + mutex 代替 once

    ```go
    // 不推荐！代码繁琐，且不如 once 高效
    var (
        instance *Config
        loaded   bool
        mu       sync.Mutex
    )
    func Load() *Config {
        if !loaded {
            mu.Lock()
            if !loaded { // 双重检查锁定
                instance = &Config{Name: "Default"}
                loaded = true
            }
            mu.Unlock()
        }
        return instance
    }
    ```

3. 在初始化中对 instance 做其他操作

    ```go
    func Load() *Config {
        once.Do(func() {
            instance = &Config{Name: "Default"}
        })
        instance.Name = "Other" // 禁止！破坏单例的不可变性
        return instance
    }
    ```

   错误！可能导致多次修改 instance。单例配置初始化后应视为只读，避免并发修改导致数据错乱