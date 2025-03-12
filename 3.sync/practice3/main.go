package main

import (
	"fmt"
	"sync"
)

type Task struct {
	Index, Num int
}

type Result struct {
	Index, Square int
}

func worker(id int, taskCh <-chan Task, resCh chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range taskCh {
		square := task.Num * task.Num
		fmt.Printf("worker-%d: 处理任务 %d(值为 %d), 计算平方值为 %d\n", id, task.Index, task.Num, square)
		resCh <- Result{Index: task.Index, Square: square}
	}
}

func main() {
	tasks := []int{2, 4, 6, 8, 10}
	workCnt := 3
	taskCh := make(chan Task, len(tasks))
	resCh := make(chan Result, len(tasks))
	var wg sync.WaitGroup

	wg.Add(workCnt)
	for i := 1; i <= workCnt; i++ {
		go worker(i, taskCh, resCh, &wg)
	}

	for id, task := range tasks {
		taskCh <- Task{Index: id, Num: task}
	}

	close(taskCh)

	wg.Wait()

	close(resCh)

	fmt.Println("所有结果：")
	resSlice := make([]int, len(resCh))
	for res := range resCh {
		resSlice[res.Index] = res.Square
	}
	fmt.Println(resSlice)
}
