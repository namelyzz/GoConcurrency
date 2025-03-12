题目：线程安全的计数器

要求：

实现一个线程安全的计数器，启动 10 个 goroutine，每个 goroutine 对计数器执行 10 次 +1 操作。
使用 sync.Mutex 保护共享变量，最终输出计数器值应为 100。

输出示例： `Final counter: 100`