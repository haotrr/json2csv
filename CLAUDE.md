# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个高性能的Go语言JSON到CSV转换工具，从[jehiah/json2csv](https://github.com/jehiah/json2csv)分支而来，并进行了性能优化和不兼容改进。

## 常用命令

### 构建和测试
```bash
# 运行测试
go test -v ./...

# 构建二进制文件
go build -o json2csv ./cmd/...

# 构建分发包（Linux和Darwin）
./dist.sh
```

### 运行示例
```bash
# 基本使用
./json2csv -k user.name,remote_ip -i input.json -o output.csv

# 从标准输入读取
cat input.json | ./json2csv -k user.name,remote_ip > output.csv

# 输出所有字段并包含标题行
./json2csv -a -H -i input.json -o output.csv
```

### 性能测试
```bash
# 运行基准测试
go test -bench=. -benchmem

# 运行特定基准测试
go test -bench=BenchmarkJSON2CSV -benchmem
```

## 代码架构

### 核心结构
- **`json2csv.go`**: 核心转换逻辑，包含性能优化功能
  - `Do()`: 主要的公共API函数
  - `json2csv()`: 内部转换实现
  - `KeyCache`: 键缓存，避免重复字符串分割
  - `BatchWriter`: 批量写入，提高I/O性能
  - `getValue()`: 优化的嵌套字段访问函数

- **`cmd/main.go`**: 命令行入口点
  - 处理命令行参数解析
  - 调用核心转换功能

- **`string_array.go`**: 自定义flag类型，用于处理逗号分隔的字段列表

- **`json2csv_test.go`**: 基准测试和性能测试
  - 包含针对KeyCache、BatchWriter、 getValue等组件的微基准测试
  - 内存分配模式测试

### 性能优化特性

1. **键缓存机制**: 使用`KeyCache`结构缓存已分割的键，避免重复字符串操作
2. **批量写入**: 使用`BatchWriter`结构批量写入CSV记录，减少I/O操作
3. **迭代式字段访问**: `getValue()`函数使用迭代而非递归访问嵌套字段
4. **内存预分配**: 预分配切片容量以减少内存重分配

### 依赖项
- `github.com/astaxie/flatmap`: 用于扁平化嵌套JSON结构
- Go 1.21+ 要求

## 测试数据
- `data/test.json` 和 `data/test2.json`: 用于测试的示例JSON文件

## 版本管理
- `VERSION` 文件包含当前版本号
- `dist.sh` 使用版本信息构建分发包