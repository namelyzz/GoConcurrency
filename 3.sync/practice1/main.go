package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const round = 5

func main() {
	//ans1()
	//ans2()
	ans3()
}

/*
ans1: 标准的使用两个 channel 实现交替控制
*/
func ans1() {
	// 创建两个控制用的 channel
	chA := make(chan bool)
	chB := make(chan bool)

	var wg sync.WaitGroup

	// 添加两个协程
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < round; i++ {
			<-chA // 等待 A 的打印信号
			fmt.Println("A")
			chB <- true // 通知 B 进行打印
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < round; i++ {
			<-chB
			fmt.Println("B")

			// 注意！最后一次打印后不需要再通知 A，此时 A 已经打印完了，再发送就会阻塞了
			if i < 4 {
				chA <- true
			}
		}
	}()

	// 发送开始打印 A 的信号
	chA <- true

	wg.Wait()
}

/*
ans2: 使用一个 channel 实现的简介做法
详细思路看文档
*/
func ans2() {
	ch := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < round; i++ {
			// 等待从 channel 接收信号
			// 这会阻塞直到另一个 goroutine 发送数据
			<-ch

			fmt.Println("A")

			// 向 channel 发送信号，通知另一个 goroutine 可以继续
			ch <- true
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < round; i++ {
			// 向 channel 发送信号，通知第一个 goroutine 可以开始
			// 注意：这个 goroutine 先发送信号，所以它控制着启动流程
			ch <- true

			fmt.Println("B")

			// 等待从 channel 接收信号
			// 这会阻塞直到第一个 goroutine 发送数据
			<-ch
		}
	}()

	wg.Wait()
}

func ans3() {
	chA := make(chan bool, 1)
	chB := make(chan bool, 1)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		count := 0
		for {
			select {
			case <-chA:
				fmt.Println("A")
				count++
				chB <- true // 无论是否到次数，都给B发信号（确保B能处理最后一次）
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		count := 0
		for {
			select {
			case <-chB:
				fmt.Println("B")
				count++
				if count < round {
					chA <- true // 没到次数，继续给A发信号
				} else {
					cancel() // B完成最后一次打印，终止流程
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	chA <- true

	select {
	case <-time.After(2 * time.Second):
		fmt.Println("Timeout!")
	case <-ctx.Done():
		fmt.Println("Completed!")
	}

	wg.Wait()
}
