# HDF5 Go 库操作文档

## 项目概述

hdf5-go 是一个纯 Go 实现的 HDF5（Hierarchical Data Format version 5）库，不依赖任何 CGO 或外部 C 库。该库提供了完整的 HDF5 文件读写功能，支持多种数据类型、压缩算法和高级特性。

### 主要特性



- **纯 Go 实现**: 无 CGO 依赖，跨平台兼容
- **完整的 HDF5 支持**: 文件、组、数据集、属性等核心操作
- **多种数据类型**: 整数、浮点数、字符串、复合类型、数组、枚举、变长等
- **压缩算法**: GZIP、LZF、ZSTD、BZIP2
- **选择操作**: 点选择、超平面选择、掩码选择
- **异步操作**: 支持异步 I/O 操作
- **原子操作**: 支持原子写入
- **引用操作**: 区域引用和对象引用
- **h5py 互操作性**: 生成的文件可被 h5py 读取，反之亦然

### 支持的操作系统

- Linux
- Windows
- macOS

## 快速入门

### 安装

```bash
go get github.com/huangzhengshun/hdf5-go
```

### 基本操作示例

#### 1. 创建文件和写入数据

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/huangzhengshun/hdf5-go"
)

func main() {
    // 创建文件
    f, err := hdf5.OpenFile("example.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建组
    g, err := f.CreateGroup("data")
    if err != nil {
        fmt.Printf("Failed to create group: %v\n", err)
        os.Exit(1)
    }
    defer g.Close()
    
    // 写入整数数据
    intData := []int32{1, 2, 3, 4, 5}
    intDset, err := g.CreateDataset("integers", hdf5.DatatypeInt32, 
        hdf5.Dataspace{Dims: []uint64{5}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create integer dataset: %v\n", err)
        os.Exit(1)
    }
    err = intDset.Write(intData)
    if err != nil {
        fmt.Printf("Failed to write integer dataset: %v\n", err)
        os.Exit(1)
    }
    intDset.Close()
    
    // 写入浮点数据
    floatData := []float64{1.1, 2.2, 3.3, 4.4, 5.5}
    floatDset, err := g.CreateDataset("floats", hdf5.DatatypeFloat64,
        hdf5.Dataspace{Dims: []uint64{5}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create float dataset: %v\n", err)
        os.Exit(1)
    }
    err = floatDset.Write(floatData)
    if err != nil {
        fmt.Printf("Failed to write float dataset: %v\n", err)
        os.Exit(1)
    }
    floatDset.Close()
    
    fmt.Println("Data written successfully!")
}
```

#### 2. 读取数据

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/huangzhengshun/hdf5-go"
)

func main() {
    // 打开文件
    f, err := hdf5.OpenFile("example.h5", hdf5.ReadOnly)
    if err != nil {
        fmt.Printf("Failed to open file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 获取组
    g, err := f.GetGroup("data")
    if err != nil {
        fmt.Printf("Failed to get group: %v\n", err)
        os.Exit(1)
    }
    defer g.Close()
    
    // 读取整数数据
    intDset, err := g.GetDataset("integers")
    if err != nil {
        fmt.Printf("Failed to get integer dataset: %v\n", err)
        os.Exit(1)
    }
    var readIntData []int32
    err = intDset.Read(&readIntData)
    if err != nil {
        fmt.Printf("Failed to read integer dataset: %v\n", err)
        os.Exit(1)
    }
    intDset.Close()
    fmt.Println("Integer data:", readIntData)
    
    // 读取浮点数据
    floatDset, err := g.GetDataset("floats")
    if err != nil {
        fmt.Printf("Failed to get float dataset: %v\n", err)
        os.Exit(1)
    }
    var readFloatData []float64
    err = floatDset.Read(&readFloatData)
    if err != nil {
        fmt.Printf("Failed to read float dataset: %v\n", err)
        os.Exit(1)
    }
    floatDset.Close()
    fmt.Println("Float data:", readFloatData)
}
```

#### 3. 创建复合数据类型

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/huangzhengshun/hdf5-go"
)

func main() {
    f, err := hdf5.OpenFile("compound.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 定义复合数据类型
    compoundDtype := hdf5.Datatype{
        Class: hdf5.DatatypeCompound,
        Size:  12,
        Fields: []hdf5.CompoundField{
            {Name: "ID", Offset: 0, Datatype: hdf5.DatatypeInt32},
            {Name: "Value", Offset: 4, Datatype: hdf5.DatatypeFloat64},
        },
    }
    
    // 创建复合数据集
    compoundDset, err := f.CreateDataset("compound_data", compoundDtype,
        hdf5.Dataspace{Dims: []uint64{3}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create compound dataset: %v\n", err)
        os.Exit(1)
    }
    
    // 写入复合数据
    compoundData := []struct {
        ID    int32
        Value float64
    }{
        {1, 10.5},
        {2, 20.5},
        {3, 30.5},
    }
    err = compoundDset.Write(compoundData)
    if err != nil {
        fmt.Printf("Failed to write compound dataset: %v\n", err)
        os.Exit(1)
    }
    compoundDset.Close()
    
    fmt.Println("Compound data written successfully!")
}
```

#### 4. 使用压缩

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/huangzhengshun/hdf5-go"
)

func main() {
    f, err := hdf5.OpenFile("compressed.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建压缩数据集（GZIP）
    plist := hdf5.PropertyList{
        Chunks:      []uint64{100},
        Compression: "gzip",
    }
    
    data := make([]float64, 1000)
    for i := range data {
        data[i] = float64(i) * 0.1
    }
    
    dset, err := f.CreateDataset("compressed_data", hdf5.DatatypeFloat64,
        hdf5.Dataspace{Dims: []uint64{1000}, MaxDims: []uint64{hdf5.H5S_UNLIMITED}}, plist)
    if err != nil {
        fmt.Printf("Failed to create compressed dataset: %v\n", err)
        os.Exit(1)
    }
    
    err = dset.Write(data)
    if err != nil {
        fmt.Printf("Failed to write compressed data: %v\n", err)
        os.Exit(1)
    }
    dset.Close()
    
    fmt.Println("Compressed data written successfully!")
}
```

## 文档目录

- [README.md](README.md) - 项目概述和快速入门
- [docs/api_reference.md](docs/api_reference.md) - 完整 API 参考
- [docs/examples.md](docs/examples.md) - 详细代码示例
- [docs/datatypes.md](docs/datatypes.md) - 数据类型说明
- [docs/compression.md](docs/compression.md) - 压缩算法说明
- [docs/errors.md](docs/errors.md) - 错误处理指南

## 测试

运行所有测试：

```bash
go test ./hdf5/
```

运行特定测试：

```bash
go test -v ./hdf5/ -run TestDatasetCreateAndWriteRead
```

## 互操作性

生成的 HDF5 文件可以使用 h5py 读取：

```python
import h5py

# 读取 Go 生成的文件
f = h5py.File('example.h5', 'r')
print(f['data/integers'][:])
print(f['data/floats'][:])
f.close()
```

h5py 生成的文件也可以使用本库读取。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！