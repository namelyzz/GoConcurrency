package case6

import "sync"

// 陷阱1: 复制包含锁的结构体
type BadStruct struct {
	mu   sync.Mutex
	data int
}

func BadCopy() {
	s1 := BadStruct{}
	s2 := s1 // 复制了锁，会导致未定义行为!
}

// 陷阱2: 重入锁（Go的Mutex不可重入）
type ReentrantExample struct {
	mu sync.Mutex
}

func (r *ReentrantExample) A() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.B() // 死锁! 在A已经持有锁的情况下调用B
}

func (r *ReentrantExample) B() {
	r.mu.Lock() // 这里会阻塞，因为A已经持有锁
	defer r.mu.Unlock()
	// ...
}
