package main

import (
    "fmt"
    "time"
)

func worker(id int) {
    fmt.Printf("Worker %d started\n", id)
    time.Sleep(2 * time.Second) // 模拟工作
    fmt.Printf("Worker %d completed\n", id)
}

func main() {
    fmt.Println("Main started")

    // 启动一个goroutine
    go worker(1)

    // 如果主程序不等待，worker可能没有机会完成
    // 取消下面这行注释看看会发生什么
    // time.Sleep(3 * time.Second)

    fmt.Println("Main ended")
    // 主程序结束，所有goroutine都会被终止
}
