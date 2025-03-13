package main

import (
	"fmt"
	"sync"
)

const (
	round        = 10
	goroutineCnt = 10
)

type Counter struct {
	count int
	mu    sync.Mutex
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) AddOne() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.count++
}

func worker(id int, wg *sync.WaitGroup, counter *Counter) {
	defer wg.Done()

	fmt.Printf("worker-%d 开始执行 %d 轮加1操作\n", id, round)
	for i := 0; i < round; i++ {
		counter.AddOne()
	}
}

func main() {
	counter := NewCounter()
	var wg sync.WaitGroup
	wg.Add(goroutineCnt)
	for i := 1; i <= goroutineCnt; i++ {
		go worker(i, &wg, counter)
	}
	wg.Wait()
	fmt.Printf("Final counter: %d\n", counter.count)
}
