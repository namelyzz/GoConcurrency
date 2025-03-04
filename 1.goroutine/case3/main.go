package main

import (
    "fmt"
    "github.com/namelyzz/GoConcurrency/1.goroutine/case3/examples"
    "time"
)

func main() {
    examples.Bad()

    time.Sleep(3 * time.Second)
    fmt.Println("--- 正确的做法 ---")

    examples.Good()
    time.Sleep(3 * time.Second)
}
