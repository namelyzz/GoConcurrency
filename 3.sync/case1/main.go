package main

import (
	"fmt"
	"sync"
)

type BankAccount struct {
	balance int
	mu      sync.Mutex
}

func (b *BankAccount) Deposit(amount int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	fmt.Printf("存款前余额: %d, 存款: %d\n", b.balance, amount)
	b.balance += amount
	fmt.Printf("存款后余额: %d\n", b.balance)
}

func (b *BankAccount) Withdraw(amount int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.balance >= amount {
		fmt.Printf("取款前余额: %d, 取款: %d\n", b.balance, amount)
		b.balance -= amount
		fmt.Printf("取款后余额: %d\n", b.balance)
		return true
	}
	return false
}

func (b *BankAccount) Balance() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.balance
}

func main() {
	acct := &BankAccount{balance: 1000}
	var wg sync.WaitGroup

	// 并发存款
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(amount int) {
			defer wg.Done()
			acct.Deposit(amount)
		}(100)
	}

	// 并发取款
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(amount int) {
			defer wg.Done()
			acct.Withdraw(amount)
		}(200)
	}

	wg.Wait()
	fmt.Printf("最终余额: %d\n", acct.Balance())
}
