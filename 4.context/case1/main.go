package main

import (
	"context"
	"fmt"
	"time"
)

func analyzeContext(ctx context.Context, name string) {
	// 1. 检查截至时间
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		fmt.Printf("%s: 截止时间 %v, 剩余 %v\n", name, deadline, remaining)
	} else {
		fmt.Printf("%s: 无截止时间\n", name)
	}

	// 2. 检查是否已完成
	select {
	case <-ctx.Done():
		fmt.Printf("%s: 已结束, 原因: %v\n", name, ctx.Err())
	default:
		fmt.Printf("%s: 仍在运行\n", name)
	}

	// 3. 检查值传递
	if value := ctx.Value("requestID"); value != nil {
		fmt.Printf("%s: requestID = %v\n", name, value)
	}

	fmt.Println("---")
}

func main() {
	// 基础Context
	bgCtx := context.Background()
	analyzeContext(bgCtx, "Background")

	// 带值的Context
	valueCtx := context.WithValue(bgCtx, "requestID", "12345")
	analyzeContext(valueCtx, "WithValue")

	// 带超时的Context
	timeoutCtx, cancel := context.WithTimeout(valueCtx, 2*time.Second)
	defer cancel()
	analyzeContext(timeoutCtx, "WithTimeout")

	time.Sleep(1 * time.Second)
	analyzeContext(timeoutCtx, "WithTimeout after 1s")

	time.Sleep(2 * time.Second)
	analyzeContext(timeoutCtx, "WithTimeout after 3s")
}
