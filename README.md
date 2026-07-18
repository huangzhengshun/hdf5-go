# HDF5 Go 库操作文档

## 项目概述

hdf5-go 是一个纯 Go 实现的 HDF5（Hierarchical Data Format version 5）库，不依赖任何 CGO 或外部 C 库。该库提供了完整的 HDF5 文件读写功能，支持多种数据类型、压缩算法和高级特性。

### 主要特性

- **纯 Go 实现**: 无 CGO 依赖，跨平台兼容
- **完整的 HDF5 支持**: 文件、组、数据集、属性等核心操作
- **多种数据类型**: 整数、浮点数、字符串、复合类型、数组、枚举、变长等
- **压缩算法**: GZIP、LZF、ZSTD（BZIP2 仅支持读取）
- **选择操作**: 点选择、超平面选择、多超平面选择、掩码选择
- **异步操作**: 支持异步 I/O 操作
- **原子操作**: 支持原子写入
- **引用操作**: 区域引用和对象引用
- **虚拟数据集**: 支持虚拟数据集（VDS）创建为只读
- **h5py 互操作性**: 生成的文件可被 h5py 读取，反之亦然

### 支持的操作系统

- Linux
- Windows
- macOS

### 已验证功能

以下功能已通过完整的测试套件验证（详见 `hdf5/feature_test.go` 和 `hdf5/full_test.go`）：

- **数据集操作**: 多维数据集（2D/3D）、所有整数类型（int8-int64、uint8-uint64）、浮点类型（float32/float64）
- **字符串数据集**: 支持普通字符串和变长字符串
- **属性操作**: 文件、组、数据集属性；支持基本类型、字符串切片、大字符串、空字符串
- **切片读写**: 支持分块（chunked）数据集的 `WriteSlice`/`ReadSlice` 操作
- **组操作**: 嵌套组、混合对象（组与数据集共存）
- **复合数据类型**: 包含混合字段的复合类型
- **特殊值**: 浮点无穷大、负零、布尔类型、负数
- **填充值**: 支持自定义填充值
- **压缩**: GZIP、LZF、ZSTD 压缩数据集的读写
- **边界情况**: 关闭对象访问、属性覆盖/删除/重建
- **变长数据类型**: 支持 `[][]int32` 等变长数据的读写往返

## 近期更新

### Bug 修复

1. **字符串属性动态大小**（`attribute.go`）：之前字符串属性大小被硬编码为 64 字节，超过 64 字符的字符串会被截断。现已根据实际字符串长度动态计算 `Size` 和 `StringSize`。

2. **分块数据集 ReadSlice 修复**（`dataset.go`）：`ReadSlice` 之前对分块数据集调用 `readChunked()`，会尝试将原始字节数据解码到 `[]byte` 类型，导致 "slice bounds out of range" panic。新增 `readChunkedRaw()` 方法返回原始字节数据，`ReadSlice` 现已正确处理分块数据集。

3. **变长数据类型往返读写**（`types.go`、`dataset.go`、`format/datatype.go`）：
   - 修复 `DatatypeClass` 常量值与 `format` 包不匹配的问题（`DatatypeArray=3`、`DatatypeEnum=5` 等）
   - `Read()` 方法对 `DatatypeVarLength` 类型现在使用 `layout.DataSize` 而不是 `dspace.Size() * dtype.Size`
   - `decodeSliceData` 对变长类型跳过基于 `dtype.Size` 的切片长度计算
   - `decodeDatatypeV2` 在 `ClassVarLength` 解码中添加简单基础类型（如 int/float）的处理

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

## 测试覆盖

项目包含以下测试文件，覆盖各种使用场景：

- `hdf5/hdf5_test.go` - 核心功能测试
- `hdf5/comprehensive_test.go` - 综合功能测试
- `hdf5/compression_test.go` - 压缩算法测试
- `hdf5/feature_test.go` - 特性测试（多维数据集、属性、切片、复合类型等）
- `hdf5/full_test.go` - 完整边界条件测试（空数据集、并发访问、特殊字符等）

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