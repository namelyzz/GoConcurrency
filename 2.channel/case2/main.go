package main

import (
	"fmt"
	"time"
)

// worker 模拟一个任务，注意 ready 和 done 的类型，是不一样的
func worker(taskID int, ready <-chan bool, done chan<- bool) {
	// 等待信号开始
	<-ready

	fmt.Printf("Worker: %d 开始工作\n", taskID)
	time.Sleep(time.Duration(taskID) * time.Second) // 模拟工作耗时
	fmt.Printf("Worker: %d 工作完成\n", taskID)

	// 发送完成信号
	done <- true
}

func main() {
	ready := make(chan bool) // 无缓冲，用于同步开始
	done := make(chan bool)  // 无缓冲，用于同步完成

	// 启动3个worker
	for i := 1; i <= 3; i++ {
		go worker(i, ready, done)
	}

	fmt.Println("协调器: 所有worker已启动，等待2秒后开始...")
	time.Sleep(2 * time.Second)

	fmt.Println("协调器: 发送开始信号")
	// 同时启动所有worker
	for i := 1; i <= 3; i++ {
		go func() { ready <- true }()
	}

	// 等待所有worker完成
	for i := 1; i <= 3; i++ {
		<-done
		fmt.Printf("协调器: 收到第%d个完成信号\n", i)
	}

	fmt.Println("协调器: 所有任务完成")
}
