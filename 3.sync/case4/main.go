package main

import (
	"fmt"
	"sync"
	"time"
)

type DBConn struct {
	url string
}

var (
	instance *DBConn
	once     sync.Once
)

func getDBConn() *DBConn {
	once.Do(func() {
		fmt.Println("创建数据库连接...")
		// 模拟耗时的连接建立
		time.Sleep(2 * time.Second)
		instance = &DBConn{url: "postgres://localhost:5432/mydb"}
		fmt.Println("数据库连接已创建!")
	})
	return instance
}

func main() {
	var wg sync.WaitGroup

	// 多个 goroutine 获取连接
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Goroutine %d: 获取数据库连接...\n", id)
			conn := getDBConn()
			fmt.Printf("Goroutine %d: 连接到 %s\n", id, conn.url)
		}(i)
	}

	wg.Wait()
}
