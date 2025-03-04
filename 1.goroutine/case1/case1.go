package main

import (
    "fmt"
    "time"
)

// 一个普通的函数
func sayHello() {
    for i := 0; i < 5; i++ {
        fmt.Printf("Hello %d\n", i)
        time.Sleep(100 * time.Millisecond)
    }
}

// 另一个函数
func sayWorld() {
    for i := 0; i < 5; i++ {
        fmt.Printf("World %d\n", i)
        time.Sleep(150 * time.Millisecond)
    }
}

func main() {
    // 顺序执行 - 注释掉这部分看看顺序执行的效果
    // sayHello()
    // sayWorld()

    // 并发执行 - 使用go关键字启动goroutine
    go sayHello() // 这会在新的goroutine中运行
    go sayWorld() // 这会在另一个新的goroutine中运行

    // 主goroutine等待一段时间，让其他goroutine有机会执行
    time.Sleep(2 * time.Second)
    fmt.Println("Main function ends")
}
