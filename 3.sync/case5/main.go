package main

import (
	"fmt"
	"sync"
	"time"
)

type TaskQueue struct {
	tasks []string
	cond  *sync.Cond
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		tasks: make([]string, 0),
		cond:  sync.NewCond(&sync.Mutex{}),
	}
}

func (q *TaskQueue) Add(task string) {
	q.cond.L.Lock()
	q.tasks = append(q.tasks, task)
	fmt.Printf("添加任务: %s\n", task)
	q.cond.L.Unlock()

	// 通知等待的消费者
	q.cond.Signal() // 通知一个等待者
	// q.cond.Broadcast() // 通知所有等待者
}

func (q *TaskQueue) Get() string {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	// 等待知道有任务可用
	for len(q.tasks) == 0 {
		fmt.Println("队列为空，等待任务...")
		q.cond.Wait() // 释放锁并等待，被唤醒时重新获取锁
	}

	task := q.tasks[0]
	q.tasks = q.tasks[1:]
	return task
}

func main() {
	queue := NewTaskQueue()
	var wg sync.WaitGroup

	// 消费者
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 3; i++ {
			task := queue.Get()
			fmt.Printf("处理任务: %s\n", task)
			time.Sleep(1 * time.Second)
		}
	}()

	// 生产者
	time.Sleep(200 * time.Millisecond) // 这里我们让生产者晚一点生产，让消费者先开始，来看看 Cond 的作用
	for i := 0; i < 3; i++ {
		queue.Add(fmt.Sprintf("task-%d", i))
		time.Sleep(500 * time.Millisecond)
	}

	wg.Wait()
}
