package examples

import "fmt"

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
