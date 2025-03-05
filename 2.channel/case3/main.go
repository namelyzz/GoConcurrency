package main

import (
	"fmt"
	"time"
)

func main() {
	// 创建一个有缓冲通道，容量为3
	ch := make(chan int, 3)

	// 生产者
	go func() {
		for i := 1; i <= 5; i++ {
			ch <- i
			fmt.Printf("生产者：发送 %d(缓冲区长度: %d/%d)\n", i, len(ch), cap(ch))
			time.Sleep(500 * time.Millisecond)
		}
		close(ch) // 关闭了 channel，表示不会再有数据发送
	}()

	// 消费者
	for i := 1; i <= 5; i++ {
		time.Sleep(2 * time.Second)
		if val, ok := <-ch; ok {
			fmt.Printf("消费者：接收 %d(缓冲区长度: %d/%d)\n", val, len(ch), cap(ch))
		}
	}
}
