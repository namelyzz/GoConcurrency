package main

import (
	"fmt"
	"sync"
	"time"
)

type ConfigManager struct {
	config map[string]string
	rw     sync.RWMutex
}

func (c *ConfigManager) Get(key string) string {
	c.rw.Lock()
	defer c.rw.Unlock()

	time.Sleep(10 * time.Millisecond) // 这里模拟读取的耗时
	return c.config[key]
}

func (c *ConfigManager) Set(key, value string) {
	c.rw.Lock()
	defer c.rw.Unlock()

	time.Sleep(30 * time.Millisecond) // 这里模拟写入的耗时
	c.config[key] = value
	fmt.Printf("配置更新: %s = %s\n", key, value)
}

func main() {
	cfg := ConfigManager{
		config: map[string]string{"version": "1.0.1", "mode": "debug", "author": "nameless"},
	}

	var wg sync.WaitGroup

	for i := 0; i < 15; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ver := cfg.Get("version")
			author := cfg.Get("author")
			fmt.Printf("Reader %d: version = %s author = %s\n", id, ver, author)
		}(i)
	}

	// 启动一个写 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		cfg.Set("version", "1.0.2")
		cfg.Set("author", "无名客")
	}()

	wg.Wait()
}
