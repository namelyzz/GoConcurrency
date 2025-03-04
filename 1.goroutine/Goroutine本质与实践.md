# **什么是Goroutine？**

Goroutine是Go语言中的轻量级线程，由Go运行时管理。与操作系统线程相比，它更加轻量、创建和销毁成本更低。Go 语言的并发模型通过 Goroutine 和通道（Channel）实现，Goroutine 是执行并发操作的基本单位。与传统的线程模型相比，Goroutine 的启动和销毁成本要低得多，因此在高并发环境下特别适用。

## 基础代码示例

首先，让我们通过一个简单的代码示例来看看如何启动并发执行的 Goroutine。在这个示例中，我们定义了两个函数 `sayHello` 和 `sayWorld`，它们分别打印 "Hello" 和 "World" 的信息，并在每次打印后暂停一定的时间。我们通过在 `main` 函数中使用 `go` 关键字将这两个函数作为 Goroutine 并发执行。主 Goroutine 会等待 2 秒钟，以确保其他 Goroutine 有足够的时间执行。

```go
package main

import (
    "fmt"
    "time"
)

// 一个普通的函数
func sayHello() {
    for i := 0; i < 5; i++ {
        fmt.Printf("Hello %d\n", i)
        time.Sleep(100 * time.Millisecond)
    }
}

// 另一个函数
func sayWorld() {
    for i := 0; i < 5; i++ {
        fmt.Printf("World %d\n", i)
        time.Sleep(150 * time.Millisecond)
    }
}

func main() {
    // 顺序执行 - 注释掉这部分看看顺序执行的效果
    // sayHello()
    // sayWorld()
    
    // 并发执行 - 使用go关键字启动goroutine
    go sayHello()  // 这会在新的goroutine中运行
    go sayWorld()  // 这会在另一个新的goroutine中运行
    
    // 主goroutine等待一段时间，让其他goroutine有机会执行
    time.Sleep(2 * time.Second)
    fmt.Println("Main function ends")
}
```

顺序执行结果

如果注释掉 `go` 关键字的部分，程序会按照顺序执行，先打印完所有的 "Hello"，再打印 "World"。

```go
Hello 0
Hello 1
Hello 2
Hello 3
Hello 4
World 0
World 1
World 2
World 3
World 4
```

并发执行结果

而当我们使用 Goroutine 并发执行时，输出会变得交错，显示出两个函数是同时执行的。

从输出结果中可以看到，`sayHello` 和 `sayWorld` 是交替执行的。Go 运行时的调度器会将它们分配到不同的操作系统线程上并交替执行。

```go
Hello 0
World 0
Hello 1
World 1
Hello 2
World 2
Hello 3
Hello 4
World 3
World 4
Main function ends
```

## goroutine 的创建和使用

下面是一个关于如何启动多个 Goroutine 的示例：在这个例子中，我们启动了三个 Goroutine，每个 Goroutine 都在打印不同的数字。与此同时，主 Goroutine 也执行了一个简单的循环。通过 `time.Sleep` 给每个 Goroutine 足够的时间来完成任务。

```go
package main

import (
    "fmt"
    "time"
)

func printNumbers(name string) {
    for i := 1; i <= 3; i++ {
        fmt.Printf("%s: %d\n", name, i)
        time.Sleep(300 * time.Millisecond)
    }
}

func main() {
    // 启动多个goroutine
    go printNumbers("Goroutine-A")
    go printNumbers("Goroutine-B")
    go printNumbers("Goroutine-C")
    
    // 主goroutine也执行一些工作
    for i := 1; i <= 2; i++ {
        fmt.Printf("Main: %d\n", i)
        time.Sleep(400 * time.Millisecond)
    }
    
    // 等待足够长时间让所有goroutine完成
    time.Sleep(2 * time.Second)
}
```

输出：可以看到，三个 Goroutine 和主 Goroutine 是交替执行的，Go 运行时调度器负责将这些 Goroutine 分配到操作系统的多个线程中执行。Goroutine 的并发执行显著提高了程序的执行效率，特别是在执行 I/O 操作时，Goroutine 可以有效减少等待时间。

```go
Main: 1
Goroutine-A: 1
Goroutine-C: 1
Goroutine-B: 1
Goroutine-B: 2
Goroutine-C: 2
Goroutine-A: 2
Main: 2
Goroutine-B: 3
Goroutine-C: 3
Goroutine-A: 3
```

## 匿名形式的 goroutine

在 Go 中，匿名 Goroutine 是指没有名称的 Goroutine，通常通过使用 `go` 关键字与匿名函数配合实现。以下是一个错误的示例：

### 错误案例

```go
func Bad() {
    for i := 0; i < 3; i++ {
        // 注意：直接使用循环变量i会有问题！
        go func() {
            fmt.Printf("Problem: %d\n", i) // 这里会有问题！
        }()
    }
}
```

在此代码中，所有的 Goroutine 都共享循环变量 `i`，这会导致每个 Goroutine 打印的结果不确定。因为 `i` 是在循环外部声明的，且在所有 Goroutine 中都可能是同一个变量，它的值会在 Goroutine 执行之前发生变化。

### 正确写法

在这种情况下，我们通过传递参数或在循环内部创建新的变量，确保每个 Goroutine 使用的是正确的值，从而避免了潜在的竞态问题。

```go
func Good() {
    for i := 0; i < 3; i++ {
        // 正确做法1：传递参数
        go func(id int) {
            fmt.Printf("Correct 1: %d\n", id)
        }(i)

        // 正确做法2：在循环内创建新的变量
        id := i
        go func() {
            fmt.Printf("Correct 2: %d\n", id)
        }()
    }
}
```

运行：

```go
func main() {
    examples.Bad()

    time.Sleep(1 * time.Second)
    fmt.Println("--- 正确的做法 ---")

    examples.Good()
}
```

## **Goroutine与主程序的生命周期**

Go 语言中的 Goroutine 是由 Go 运行时管理的，但主程序的生命周期依然对 Goroutine 的执行有影响。以下是一个例子：

```go
package main

import (
    "fmt"
    "time"
)

func worker(id int) {
    fmt.Printf("Worker %d started\n", id)
    time.Sleep(2 * time.Second) // 模拟工作
    fmt.Printf("Worker %d completed\n", id)
}

func main() {
    fmt.Println("Main started")

    // 启动一个goroutine
    go worker(1)

    // 如果主程序不等待，worker可能没有机会完成
    // 取消下面这行注释看看会发生什么
    // time.Sleep(3 * time.Second)

    fmt.Println("Main ended")
    // 主程序结束，所有goroutine都会被终止
}

```

在此代码中，主程序启动了一个 Goroutine 来模拟一个工作任务。若没有让主程序等待足够的时间，Goroutine 可能在主程序结束前没有机会完成。主程序在执行完毕后会终止，因此所有 Goroutine 也会被强制终止。

本小节，我们在主程序中使用 `time.Sleep` 来等待 Goroutine 执行完毕。但是，这种方式并不优雅，没理由每一个 main 都设置一个 `time.Sleep` 吧。而且，有时候设置多长时间也是不好确定的。那么，随着我们深入学习 goroutine，这个问题的答案也会浮出水面。

## 本文配套代码

https://github.com/namelyzz/GoConcurrency/tree/main/1.goroutine