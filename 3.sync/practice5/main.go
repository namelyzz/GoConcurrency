package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	readRound  = 4
	readGoCnt  = 3
	writeRound = 8
	writeGoCnt = 5
)

type Logger struct {
	logs  []string
	count int
	rwmu  sync.RWMutex
	wg    sync.WaitGroup
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) statLog(goID int) {
	defer l.wg.Done()

	for i := 1; i <= readRound; i++ {
		time.Sleep(100 * time.Millisecond)
		l.rwmu.RLock()
		logCnt := l.count
		latestLog := l.logs[logCnt-1]
		fmt.Printf("统计协程-%d: 当前日志总数: %d, 最新日志记录为: %s\n", goID, logCnt, latestLog)
		l.rwmu.RUnlock()
	}
}

func (l *Logger) writeLog(goID int) {
	defer l.wg.Done()

	for i := 1; i <= writeRound; i++ {
		l.rwmu.Lock()
		logMsg := fmt.Sprintf("goroutine-%d: 第 %d 条日志", goID, i)
		l.logs = append(l.logs, logMsg)
		l.count++
		l.rwmu.Unlock()
		time.Sleep(50 * time.Millisecond) // 模拟日志生成耗时
	}
}

func main() {
	logger := NewLogger()

	logger.wg.Add(writeGoCnt)
	for i := 1; i <= writeGoCnt; i++ {
		go logger.writeLog(i)
	}

	logger.wg.Add(readGoCnt)
	for i := 1; i <= readGoCnt; i++ {
		go logger.statLog(i)
	}

	logger.wg.Wait()

	// 最终统计
	logger.rwmu.RLock()
	defer logger.rwmu.RUnlock()

	fmt.Printf("所有日志条数: %d\n", logger.count)
}
