package main

import (
	"fmt"
	"math/rand"
	"time"
)

func worker(id int, jobs <-chan int, results chan<- int) {
	for job := range jobs {
		fmt.Printf("Worker %d: 开始处理任务 %d\n", id, job)
		time.Sleep(time.Duration(rand.Intn(3)) * time.Second) // 模拟处理时间
		result := job * 2
		fmt.Printf("Worker %d: 任务 %d 完成，结果 %d\n", id, job, result)
		results <- result
	}
	fmt.Printf("Worker %d: 所有任务处理完成\n", id)
}

func main() {
	const numJobs = 10
	const numWorkers = 3

	job := make(chan int, numJobs)
	results := make(chan int, numJobs)

	for i := 1; i <= numWorkers; i++ {
		go worker(i, job, results)
	}

	for i := 1; i <= numJobs; i++ {
		job <- i
	}
	close(job)

	for i := 1; i <= numJobs; i++ {
		result := <-results
		fmt.Printf("主程序: 收到结果 %d\n", result)
	}

	fmt.Println("所有任务处理完成")
}
