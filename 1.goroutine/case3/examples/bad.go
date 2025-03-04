package examples

import (
    "fmt"
)

func Bad() {
    for i := 0; i < 3; i++ {
        // 注意：直接使用循环变量i会有问题！
        go func() {
            fmt.Printf("Problem: %d\n", i) // 这里会有问题！
        }()
    }
}
