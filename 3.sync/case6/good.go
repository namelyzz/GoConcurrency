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
