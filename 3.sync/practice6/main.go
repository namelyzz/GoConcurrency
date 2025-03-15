package main

import (
	"fmt"
	"sync"
)

type Config struct {
	Name string
}

var (
	instance *Config
	once     sync.Once
)

func Load() *Config {
	once.Do(func() {
		fmt.Println("初始化配置...")
		instance = &Config{Name: "Default"}
	})
	return instance
}

func main() {
	var wg sync.WaitGroup

	wg.Add(10)
	for i := 1; i <= 10; i++ {
		go func(goID int) {
			defer wg.Done()

			for j := 1; j <= 3; j++ {
				cfg := Load()
				fmt.Printf("goroutine-%d 第%d次调用：实例地址=%p，Name=%s\n", goID, j, cfg, cfg.Name)
			}
		}(i)
	}

	wg.Wait()
}
