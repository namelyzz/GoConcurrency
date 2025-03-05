package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan string) // 这是一个无缓冲的 channel

	// 生产者 goroutine
	go func() {
		fmt.Println("生产者 -> 准备发送数据")
		ch <- "Hello Nameless!" // 这个发送操作会阻塞，直到有人接收它
		fmt.Println("生产者发送数据完毕")
	}()

	// 让主 goroutine 等待一段时间
	time.Sleep(2 * time.Second)

	fmt.Println("主程序 -> 开始接收数据")
	msg := <-ch
	fmt.Printf("主程序已收到消息：%s\n", msg)

	time.Sleep(1 * time.Second)
}
