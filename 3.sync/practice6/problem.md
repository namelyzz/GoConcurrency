题目：单例配置加载器

要求：实现一个全局配置加载器 Config，使用 sync.Once 确保初始化只执行一次。

1. 多个 goroutine 同时调用 Load()，只执行一次初始化
2. 初始化后，Config 的 Name 字段固定为 "Default"
3. 避免重复初始化（如 if !loaded 逻辑）