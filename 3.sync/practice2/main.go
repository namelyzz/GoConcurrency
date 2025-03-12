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
