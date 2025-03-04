package main

import (
    "fmt"
    "time"
)

func printNumbers(name string) {
    for i := 1; i <= 3; i++ {
        fmt.Printf("%s: %d\n", name, i)
        time.Sleep(300 * time.Millisecond)
    }
}

func main() {
    // 启动多个goroutine
    go printNumbers("Goroutine-A")
    go printNumbers("Goroutine-B")
    go printNumbers("Goroutine-C")

    // 主goroutine也执行一些工作
    for i := 1; i <= 2; i++ {
        fmt.Printf("Main: %d\n", i)
        time.Sleep(400 * time.Millisecond)
    }

    // 等待足够长时间让所有goroutine完成
    time.Sleep(2 * time.Second)
}
