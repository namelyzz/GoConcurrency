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
