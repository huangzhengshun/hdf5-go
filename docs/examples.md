# 代码示例

## 目录

1. [基本文件操作](#1-基本文件操作)
2. [组操作](#2-组操作)
3. [数据集操作](#3-数据集操作)
4. [属性操作](#4-属性操作)
5. [数据类型](#5-数据类型)
6. [压缩](#6-压缩)
7. [选择操作](#7-选择操作)
8. [链接操作](#8-链接操作)
9. [引用操作](#9-引用操作)
10. [异步操作](#10-异步操作)
11. [原子操作](#11-原子操作)
12. [数据集调整大小](#12-数据集调整大小)
13. [互操作性](#13-互操作性)

---

## 1. 基本文件操作

### 1.1 创建和关闭文件

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    // 创建新文件
    f, err := hdf5.OpenFile("basic_file.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    fmt.Println("File created successfully!")
}
```

### 1.2 打开现有文件

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    // 以只读模式打开
    f, err := hdf5.OpenFile("basic_file.h5", hdf5.ReadOnly)
    if err != nil {
        fmt.Printf("Failed to open file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    fmt.Println("File opened successfully!")
}
```

### 1.3 创建内存文件

```go
package main

import (
    "fmt"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    // 创建内存中的 HDF5 文件
    f, err := hdf5.CreateMemoryFile()
    if err != nil {
        fmt.Printf("Failed to create memory file: %v\n", err)
        return
    }
    defer f.Close()
    
    // 在内存文件中创建数据集
    data := []int32{1, 2, 3}
    dset, err := f.CreateDataset("memory_data", hdf5.DatatypeInt32,
        hdf5.Dataspace{Dims: []uint64{3}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create dataset: %v\n", err)
        return
    }
    dset.Write(data)
    dset.Close()
    
    // 获取文件内容
    bytes, err := f.Bytes()
    if err != nil {
        fmt.Printf("Failed to get bytes: %v\n", err)
        return
    }
    
    fmt.Printf("Memory file size: %d bytes\n", len(bytes))
}
```

---

## 2. 组操作

### 2.1 创建和获取组

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("groups.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建顶层组
    rootGroup, err := f.CreateGroup("root")
    if err != nil {
        fmt.Printf("Failed to create root group: %v\n", err)
        os.Exit(1)
    }
    defer rootGroup.Close()
    
    // 创建嵌套组
    nestedGroup, err := rootGroup.CreateGroup("nested")
    if err != nil {
        fmt.Printf("Failed to create nested group: %v\n", err)
        os.Exit(1)
    }
    defer nestedGroup.Close()
    
    fmt.Println("Groups created successfully!")
}
```

### 2.2 遍历组内容

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("groups.h5", hdf5.ReadOnly)
    if err != nil {
        fmt.Printf("Failed to open file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    g, err := f.GetGroup("root")
    if err != nil {
        fmt.Printf("Failed to get group: %v\n", err)
        os.Exit(1)
    }
    defer g.Close()
    
    // 获取所有键名
    keys := g.Keys()
    fmt.Println("Group keys:", keys)
    
    // 递归访问所有对象
    fmt.Println("\nRecursive visit:")
    err = g.Visit(func(path string, obj interface{}) error {
        fmt.Println("  Path:", path)
        return nil
    })
    if err != nil {
        fmt.Printf("Failed to visit: %v\n", err)
    }
}
```

### 2.3 复制和移动组

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("copy_move.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建源组
    srcGroup, err := f.CreateGroup("source")
    if err != nil {
        fmt.Printf("Failed to create source group: %v\n", err)
        os.Exit(1)
    }
    
    // 在源组中创建数据集
    data := []int32{1, 2, 3}
    _, err = srcGroup.CreateDataset("data", hdf5.DatatypeInt32,
        hdf5.Dataspace{Dims: []uint64{3}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create dataset: %v\n", err)
        os.Exit(1)
    }
    srcGroup.Close()
    
    // 创建目标组
    destGroup, err := f.CreateGroup("destination")
    if err != nil {
        fmt.Printf("Failed to create destination group: %v\n", err)
        os.Exit(1)
    }
    defer destGroup.Close()
    
    // 复制组
    srcGroup, err = f.GetGroup("source")
    if err != nil {
        fmt.Printf("Failed to get source group: %v\n", err)
        os.Exit(1)
    }
    err = srcGroup.Copy(destGroup, "copied")
    if err != nil {
        fmt.Printf("Failed to copy group: %v\n", err)
        os.Exit(1)
    }
    srcGroup.Close()
    
    // 移动组
    srcGroup, err = f.GetGroup("source")
    if err != nil {
        fmt.Printf("Failed to get source group: %v\n", err)
        os.Exit(1)
    }
    err = srcGroup.Move(destGroup, "moved")
    if err != nil {
        fmt.Printf("Failed to move group: %v\n", err)
        os.Exit(1)
    }
    srcGroup.Close()
    
    fmt.Println("Copy and move operations completed!")
}
```

---

## 3. 数据集操作

### 3.1 创建和写入数据集

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("datasets.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建整数数据集
    intData := []int32{1, 2, 3, 4, 5}
    intDset, err := f.CreateDataset("integers", hdf5.DatatypeInt32,
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
    
    // 创建浮点数据集
    floatData := []float64{1.1, 2.2, 3.3, 4.4, 5.5}
    floatDset, err := f.CreateDataset("floats", hdf5.DatatypeFloat64,
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
    
    fmt.Println("Datasets created and written successfully!")
}
```

### 3.2 读取数据集

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("datasets.h5", hdf5.ReadOnly)
    if err != nil {
        fmt.Printf("Failed to open file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 读取整数数据集
    intDset, err := f.GetDataset("integers")
    if err != nil {
        fmt.Printf("Failed to get integer dataset: %v\n", err)
        os.Exit(1)
    }
    var intResult []int32
    err = intDset.Read(&intResult)
    if err != nil {
        fmt.Printf("Failed to read integer dataset: %v\n", err)
        os.Exit(1)
    }
    intDset.Close()
    fmt.Println("Integer data:", intResult)
    
    // 读取浮点数据集
    floatDset, err := f.GetDataset("floats")
    if err != nil {
        fmt.Printf("Failed to get float dataset: %v\n", err)
        os.Exit(1)
    }
    var floatResult []float64
    err = floatDset.Read(&floatResult)
    if err != nil {
        fmt.Printf("Failed to read float dataset: %v\n", err)
        os.Exit(1)
    }
    floatDset.Close()
    fmt.Println("Float data:", floatResult)
}
```

### 3.3 多维数据集

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("multidim.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建 2D 数据集 (3x4)
    data := []float32{
        1.0, 2.0, 3.0, 4.0,
        5.0, 6.0, 7.0, 8.0,
        9.0, 10.0, 11.0, 12.0,
    }
    dset, err := f.CreateDataset("2d_data", hdf5.DatatypeFloat32,
        hdf5.Dataspace{Dims: []uint64{3, 4}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create 2D dataset: %v\n", err)
        os.Exit(1)
    }
    err = dset.Write(data)
    if err != nil {
        fmt.Printf("Failed to write 2D dataset: %v\n", err)
        os.Exit(1)
    }
    dset.Close()
    
    fmt.Println("2D dataset created successfully!")
}
```

---

## 4. 属性操作

### 4.1 设置和获取属性

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("attributes.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    g, err := f.CreateGroup("my_group")
    if err != nil {
        fmt.Printf("Failed to create group: %v\n", err)
        os.Exit(1)
    }
    defer g.Close()
    
    // 设置属性
    err = g.Attributes().Set("description", "This is a test group")
    if err != nil {
        fmt.Printf("Failed to set description: %v\n", err)
        os.Exit(1)
    }
    
    err = g.Attributes().Set("version", int32(1))
    if err != nil {
        fmt.Printf("Failed to set version: %v\n", err)
        os.Exit(1)
    }
    
    err = g.Attributes().Set("values", []float64{1.0, 2.0, 3.0})
    if err != nil {
        fmt.Printf("Failed to set values: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("Attributes set successfully!")
}
```

### 4.2 读取属性

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("attributes.h5", hdf5.ReadOnly)
    if err != nil {
        fmt.Printf("Failed to open file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    g, err := f.GetGroup("my_group")
    if err != nil {
        fmt.Printf("Failed to get group: %v\n", err)
        os.Exit(1)
    }
    defer g.Close()
    
    // 获取字符串属性
    attr, err := g.Attributes().Get("description")
    if err != nil {
        fmt.Printf("Failed to get description: %v\n", err)
        os.Exit(1)
    }
    var desc string
    err = attr.Read(&desc)
    if err != nil {
        fmt.Printf("Failed to read description: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Description:", desc)
    
    // 获取整数属性
    attr, err = g.Attributes().Get("version")
    if err != nil {
        fmt.Printf("Failed to get version: %v\n", err)
        os.Exit(1)
    }
    var version int32
    err = attr.Read(&version)
    if err != nil {
        fmt.Printf("Failed to read version: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Version:", version)
    
    // 获取数组属性
    attr, err = g.Attributes().Get("values")
    if err != nil {
        fmt.Printf("Failed to get values: %v\n", err)
        os.Exit(1)
    }
    var values []float64
    err = attr.Read(&values)
    if err != nil {
        fmt.Printf("Failed to read values: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Values:", values)
}
```

### 4.3 迭代属性

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("attributes.h5", hdf5.ReadOnly)
    if err != nil {
        fmt.Printf("Failed to open file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    g, err := f.GetGroup("my_group")
    if err != nil {
        fmt.Printf("Failed to get group: %v\n", err)
        os.Exit(1)
    }
    defer g.Close()
    
    // 迭代所有属性
    attrs, err := g.Attributes().Iter()
    if err != nil {
        fmt.Printf("Failed to iterate attributes: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Printf("Found %d attributes:\n", len(attrs))
    for _, attr := range attrs {
        fmt.Println("  Attribute:", attr.Name())
    }
}
```

---

## 5. 数据类型

### 5.1 复合数据类型

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
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
    
    fmt.Println("Compound dataset created successfully!")
}
```

### 5.2 枚举数据类型

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("enum.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 定义枚举数据类型
    enumDtype := hdf5.Datatype{
        Class: hdf5.DatatypeEnum,
        Size:  4,
        BaseType: &hdf5.DatatypeInt32,
        EnumMembers: []hdf5.EnumMember{
            {Name: "Low", Value: 0},
            {Name: "Medium", Value: 1},
            {Name: "High", Value: 2},
        },
    }
    
    // 创建枚举数据集
    dset, err := f.CreateDataset("enum_data", enumDtype,
        hdf5.Dataspace{Dims: []uint64{3}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create enum dataset: %v\n", err)
        os.Exit(1)
    }
    
    // 写入枚举数据
    data := []int32{0, 1, 2}
    err = dset.Write(data)
    if err != nil {
        fmt.Printf("Failed to write enum dataset: %v\n", err)
        os.Exit(1)
    }
    dset.Close()
    
    fmt.Println("Enum dataset created successfully!")
}
```

### 5.3 数组数据类型

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("array.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 定义数组数据类型（每个元素是 3 个 int32）
    arrayDtype := hdf5.Datatype{
        Class:     hdf5.DatatypeArray,
        Size:      12,
        BaseType:  &hdf5.DatatypeInt32,
        ArrayDims: []uint64{3},
    }
    
    // 创建数组数据集
    dset, err := f.CreateDataset("array_data", arrayDtype,
        hdf5.Dataspace{Dims: []uint64{2}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create array dataset: %v\n", err)
        os.Exit(1)
    }
    
    // 写入数组数据
    data := []int32{1, 2, 3, 4, 5, 6}
    err = dset.Write(data)
    if err != nil {
        fmt.Printf("Failed to write array dataset: %v\n", err)
        os.Exit(1)
    }
    dset.Close()
    
    fmt.Println("Array dataset created successfully!")
}
```

### 5.4 变长数据类型

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("varlength.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 定义变长数据类型
    varlenDtype := hdf5.Datatype{
        Class:    hdf5.DatatypeVarLength,
        BaseType: &hdf5.DatatypeInt32,
    }
    
    // 创建变长数据集
    dset, err := f.CreateDataset("varlen_data", varlenDtype,
        hdf5.Dataspace{Dims: []uint64{2}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create varlength dataset: %v\n", err)
        os.Exit(1)
    }
    
    // 写入变长数据
    data := [][]int32{
        {1, 2, 3},
        {4, 5, 6, 7, 8},
    }
    err = dset.Write(data)
    if err != nil {
        fmt.Printf("Failed to write varlength dataset: %v\n", err)
        os.Exit(1)
    }
    dset.Close()
    
    fmt.Println("VarLength dataset created successfully!")
}
```

---

## 6. 压缩

### 6.1 GZIP 压缩

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("compressed_gzip.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建带 GZIP 压缩的数据集
    plist := hdf5.PropertyList{
        Chunks:      []uint64{100},
        Compression: "gzip",
    }
    
    data := make([]float64, 1000)
    for i := range data {
        data[i] = float64(i) * 0.1
    }
    
    dset, err := f.CreateDataset("gzip_data", hdf5.DatatypeFloat64,
        hdf5.Dataspace{Dims: []uint64{1000}, MaxDims: []uint64{hdf5.H5S_UNLIMITED}}, plist)
    if err != nil {
        fmt.Printf("Failed to create GZIP dataset: %v\n", err)
        os.Exit(1)
    }
    
    err = dset.Write(data)
    if err != nil {
        fmt.Printf("Failed to write GZIP data: %v\n", err)
        os.Exit(1)
    }
    dset.Close()
    
    fmt.Println("GZIP compressed dataset created successfully!")
}
```

### 6.2 LZF 压缩

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("compressed_lzf.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建带 LZF 压缩的数据集
    plist := hdf5.PropertyList{
        Chunks:      []uint64{100},
        Compression: "lzf",
    }
    
    data := make([]float64, 1000)
    for i := range data {
        data[i] = float64(i) * 0.1
    }
    
    dset, err := f.CreateDataset("lzf_data", hdf5.DatatypeFloat64,
        hdf5.Dataspace{Dims: []uint64{1000}, MaxDims: []uint64{hdf5.H5S_UNLIMITED}}, plist)
    if err != nil {
        fmt.Printf("Failed to create LZF dataset: %v\n", err)
        os.Exit(1)
    }
    
    err = dset.Write(data)
    if err != nil {
        fmt.Printf("Failed to write LZF data: %v\n", err)
        os.Exit(1)
    }
    dset.Close()
    
    fmt.Println("LZF compressed dataset created successfully!")
}
```

### 6.3 ZSTD 压缩

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("compressed_zstd.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建带 ZSTD 压缩的数据集
    plist := hdf5.PropertyList{
        Chunks:      []uint64{100},
        Compression: "zstd",
    }
    
    data := make([]float64, 1000)
    for i := range data {
        data[i] = float64(i) * 0.1
    }
    
    dset, err := f.CreateDataset("zstd_data", hdf5.DatatypeFloat64,
        hdf5.Dataspace{Dims: []uint64{1000}, MaxDims: []uint64{hdf5.H5S_UNLIMITED}}, plist)
    if err != nil {
        fmt.Printf("Failed to create ZSTD dataset: %v\n", err)
        os.Exit(1)
    }
    
    err = dset.Write(data)
    if err != nil {
        fmt.Printf("Failed to write ZSTD data: %v\n", err)
        os.Exit(1)
    }
    dset.Close()
    
    fmt.Println("ZSTD compressed dataset created successfully!")
}
```

---

## 7. 选择操作

### 7.1 点选择

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("selection.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建数据集
    data := make([]int32, 10)
    for i := range data {
        data[i] = int32(i)
    }
    dset, err := f.CreateDataset("data", hdf5.DatatypeInt32,
        hdf5.Dataspace{Dims: []uint64{10}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create dataset: %v\n", err)
        os.Exit(1)
    }
    err = dset.Write(data)
    if err != nil {
        fmt.Printf("Failed to write data: %v\n", err)
        os.Exit(1)
    }
    
    // 点选择：读取索引为 1, 3, 5, 7, 9 的元素
    sel := hdf5.Selection{
        Points: [][]uint64{{1}, {3}, {5}, {7}, {9}},
    }
    var result []int32
    err = dset.ReadSlice(&result, sel)
    if err != nil {
        fmt.Printf("Failed to read points: %v\n", err)
        os.Exit(1)
    }
    dset.Close()
    
    fmt.Println("Point selection result:", result)
}
```

### 7.2 超平面选择

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("selection.h5", hdf5.ReadOnly)
    if err != nil {
        fmt.Printf("Failed to open file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    dset, err := f.GetDataset("data")
    if err != nil {
        fmt.Printf("Failed to get dataset: %v\n", err)
        os.Exit(1)
    }
    defer dset.Close()
    
    // 超平面选择：从索引 0 开始，读取 3 个元素
    sel := hdf5.Selection{
        Hyperslabs: []hdf5.Hyperslab{
            {Start: []uint64{0}, Count: []uint64{3}},
        },
    }
    var result []int32
    err = dset.ReadSlice(&result, sel)
    if err != nil {
        fmt.Printf("Failed to read hyperslab: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("Hyperslab selection result:", result)
}
```

### 7.3 多超平面选择

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("selection.h5", hdf5.ReadOnly)
    if err != nil {
        fmt.Printf("Failed to open file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    dset, err := f.GetDataset("data")
    if err != nil {
        fmt.Printf("Failed to get dataset: %v\n", err)
        os.Exit(1)
    }
    defer dset.Close()
    
    // 多超平面选择
    sel := hdf5.Selection{
        Hyperslabs: []hdf5.Hyperslab{
            {Start: []uint64{0}, Count: []uint64{3}},
            {Start: []uint64{5}, Count: []uint64{2}},
        },
    }
    var result []int32
    err = dset.ReadSlice(&result, sel)
    if err != nil {
        fmt.Printf("Failed to read multi-hyperslab: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("Multi-hyperslab selection result:", result)
}
```

### 7.4 掩码选择

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("selection.h5", hdf5.ReadOnly)
    if err != nil {
        fmt.Printf("Failed to open file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    dset, err := f.GetDataset("data")
    if err != nil {
        fmt.Printf("Failed to get dataset: %v\n", err)
        os.Exit(1)
    }
    defer dset.Close()
    
    // 掩码选择：只读取索引为 1, 3, 5, 7, 9 的元素
    mask := []bool{false, true, false, true, false, true, false, true, false, true}
    sel := hdf5.Selection{Mask: mask}
    var result []int32
    err = dset.ReadSlice(&result, sel)
    if err != nil {
        fmt.Printf("Failed to read with mask: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("Mask selection result:", result)
}
```

---

## 8. 链接操作

### 8.1 创建软链接

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("links.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建原始数据集
    data := []int32{1, 2, 3}
    dset, err := f.CreateDataset("original_data", hdf5.DatatypeInt32,
        hdf5.Dataspace{Dims: []uint64{3}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create original dataset: %v\n", err)
        os.Exit(1)
    }
    err = dset.Write(data)
    if err != nil {
        fmt.Printf("Failed to write original data: %v\n", err)
        os.Exit(1)
    }
    dset.Close()
    
    // 创建软链接
    err = f.CreateLink("/original_data", "my_link")
    if err != nil {
        fmt.Printf("Failed to create soft link: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("Soft link created successfully!")
}
```

### 8.2 创建外部链接

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    // 创建外部文件
    externalFile, err := hdf5.OpenFile("external.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create external file: %v\n", err)
        os.Exit(1)
    }
    
    // 在外部文件中创建数据集
    data := []float64{1.1, 2.2, 3.3}
    dset, err := externalFile.CreateDataset("external_data", hdf5.DatatypeFloat64,
        hdf5.Dataspace{Dims: []uint64{3}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create external dataset: %v\n", err)
        os.Exit(1)
    }
    err = dset.Write(data)
    if err != nil {
        fmt.Printf("Failed to write external data: %v\n", err)
        os.Exit(1)
    }
    dset.Close()
    externalFile.Close()
    
    // 创建主文件并添加外部链接
    mainFile, err := hdf5.OpenFile("main.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create main file: %v\n", err)
        os.Exit(1)
    }
    defer mainFile.Close()
    
    err = mainFile.CreateExternalLink("external.h5", "/external_data", "external_link")
    if err != nil {
        fmt.Printf("Failed to create external link: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("External link created successfully!")
}
```

---

## 9. 引用操作

### 9.1 创建对象引用

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("references.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建数据集
    data := []int32{1, 2, 3}
    dset, err := f.CreateDataset("target_data", hdf5.DatatypeInt32,
        hdf5.Dataspace{Dims: []uint64{3}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create target dataset: %v\n", err)
        os.Exit(1)
    }
    err = dset.Write(data)
    if err != nil {
        fmt.Printf("Failed to write target data: %v\n", err)
        os.Exit(1)
    }
    
    // 创建对象引用
    ref, err := dset.CreateObjectReference()
    if err != nil {
        fmt.Printf("Failed to create object reference: %v\n", err)
        os.Exit(1)
    }
    dset.Close()
    
    // 创建存储引用的数据集
    refDset, err := f.CreateDataset("reference", hdf5.DatatypeObjectReference,
        hdf5.Dataspace{Dims: []uint64{1}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create reference dataset: %v\n", err)
        os.Exit(1)
    }
    err = refDset.Write([]hdf5.ObjectReference{ref})
    if err != nil {
        fmt.Printf("Failed to write reference: %v\n", err)
        os.Exit(1)
    }
    refDset.Close()
    
    fmt.Println("Object reference created successfully!")
}
```

### 9.2 创建区域引用

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("region_refs.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建目标数据集
    data := make([]int32, 10)
    for i := range data {
        data[i] = int32(i)
    }
    dset, err := f.CreateDataset("target", hdf5.DatatypeInt32,
        hdf5.Dataspace{Dims: []uint64{10}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create target dataset: %v\n", err)
        os.Exit(1)
    }
    err = dset.Write(data)
    if err != nil {
        fmt.Printf("Failed to write target data: %v\n", err)
        os.Exit(1)
    }
    
    // 创建区域引用（指向索引 2-5 的元素）
    sel := hdf5.Selection{
        Hyperslabs: []hdf5.Hyperslab{
            {Start: []uint64{2}, Count: []uint64{4}},
        },
    }
    ref, err := dset.CreateReference(sel)
    if err != nil {
        fmt.Printf("Failed to create region reference: %v\n", err)
        os.Exit(1)
    }
    dset.Close()
    
    // 创建存储区域引用的数据集
    refDset, err := f.CreateDataset("region_ref", hdf5.DatatypeRegionReference,
        hdf5.Dataspace{Dims: []uint64{1}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create region reference dataset: %v\n", err)
        os.Exit(1)
    }
    err = refDset.Write([]hdf5.RegionReference{ref})
    if err != nil {
        fmt.Printf("Failed to write region reference: %v\n", err)
        os.Exit(1)
    }
    refDset.Close()
    
    fmt.Println("Region reference created successfully!")
}
```

---

## 10. 异步操作

```go
package main

import (
    "fmt"
    "sync"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.CreateMemoryFile()
    if err != nil {
        fmt.Printf("Failed to create memory file: %v\n", err)
        return
    }
    defer f.Close()
    
    // 创建数据集
    dset, err := f.CreateDataset("async_data", hdf5.DatatypeInt32,
        hdf5.Dataspace{Dims: []uint64{1000}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create dataset: %v\n", err)
        return
    }
    defer dset.Close()
    
    // 创建异步管理器
    manager := hdf5.NewAsyncManager()
    
    // 准备数据
    data := make([]int32, 1000)
    for i := range data {
        data[i] = int32(i)
    }
    
    // 提交异步写入任务
    var wg sync.WaitGroup
    wg.Add(1)
    
    go func() {
        defer wg.Done()
        
        result := manager.Submit(func() error {
            return dset.Write(data)
        })
        
        // 等待完成
        err := result.Wait()
        if err != nil {
            fmt.Printf("Async write failed: %v\n", err)
            return
        }
        
        fmt.Println("Async write completed successfully!")
    }()
    
    wg.Wait()
}
```

---

## 11. 原子操作

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("atomic.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建数据集
    dset, err := f.CreateDataset("atomic_data", hdf5.DatatypeInt32,
        hdf5.Dataspace{Dims: []uint64{100}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create dataset: %v\n", err)
        os.Exit(1)
    }
    defer dset.Close()
    
    // 准备数据
    data := make([]int32, 100)
    for i := range data {
        data[i] = int32(i)
    }
    
    // 原子写入（保证写入的完整性）
    err = dset.WriteAtomic(data)
    if err != nil {
        fmt.Printf("Atomic write failed: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("Atomic write completed successfully!")
}
```

---

## 12. 数据集调整大小

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("resize.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建可扩展数据集
    data := []int32{1, 2, 3, 4, 5}
    space := hdf5.Dataspace{Dims: []uint64{5}, MaxDims: []uint64{hdf5.H5S_UNLIMITED}}
    dset, err := f.CreateDataset("resize_data", hdf5.DatatypeInt32, space, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create dataset: %v\n", err)
        os.Exit(1)
    }
    defer dset.Close()
    
    // 写入初始数据
    err = dset.Write(data)
    if err != nil {
        fmt.Printf("Failed to write initial data: %v\n", err)
        os.Exit(1)
    }
    
    // 调整大小
    err = dset.Resize([]uint64{8})
    if err != nil {
        fmt.Printf("Failed to resize dataset: %v\n", err)
        os.Exit(1)
    }
    
    // 写入追加数据
    appendData := []int32{6, 7, 8}
    sel := hdf5.Selection{
        Hyperslabs: []hdf5.Hyperslab{
            {Start: []uint64{5}, Count: []uint64{3}},
        },
    }
    err = dset.WriteSlice(appendData, sel)
    if err != nil {
        fmt.Printf("Failed to write appended data: %v\n", err)
        os.Exit(1)
    }
    
    // 读取完整数据
    var result []int32
    err = dset.Read(&result)
    if err != nil {
        fmt.Printf("Failed to read data: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("Resized data:", result)
}
```

---

## 13. 互操作性

### 13.1 Go 生成文件，h5py 读取

**Go 代码：**

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("interop.h5", hdf5.Create)
    if err != nil {
        fmt.Printf("Failed to create file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    // 创建组
    g, err := f.CreateGroup("go_data")
    if err != nil {
        fmt.Printf("Failed to create group: %v\n", err)
        os.Exit(1)
    }
    defer g.Close()
    
    // 设置属性
    err = g.Attributes().Set("creator", "go-hdf5")
    if err != nil {
        fmt.Printf("Failed to set attribute: %v\n", err)
        os.Exit(1)
    }
    
    // 创建数据集
    data := []float64{1.1, 2.2, 3.3, 4.4, 5.5}
    dset, err := g.CreateDataset("values", hdf5.DatatypeFloat64,
        hdf5.Dataspace{Dims: []uint64{5}}, hdf5.PropertyList{})
    if err != nil {
        fmt.Printf("Failed to create dataset: %v\n", err)
        os.Exit(1)
    }
    err = dset.Write(data)
    if err != nil {
        fmt.Printf("Failed to write data: %v\n", err)
        os.Exit(1)
    }
    dset.Close()
    
    fmt.Println("File created successfully. Can be read by h5py.")
}
```

**h5py 代码：**

```python
import h5py

# 读取 Go 生成的文件
f = h5py.File('interop.h5', 'r')
print("Creator:", f['go_data'].attrs['creator'])
print("Values:", f['go_data/values'][:])
f.close()
```

### 13.2 h5py 生成文件，Go 读取

**h5py 代码：**

```python
import h5py
import numpy as np

# 创建文件
f = h5py.File('from_h5py.h5', 'w')

# 创建组
g = f.create_group('python_data')

# 设置属性
g.attrs['creator'] = 'h5py'

# 创建数据集
data = np.array([1, 2, 3, 4, 5], dtype=np.int32)
g.create_dataset('values', data=data)

f.close()
```

**Go 代码：**

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/hdf5-go/hdf5"
)

func main() {
    f, err := hdf5.OpenFile("from_h5py.h5", hdf5.ReadOnly)
    if err != nil {
        fmt.Printf("Failed to open file: %v\n", err)
        os.Exit(1)
    }
    defer f.Close()
    
    g, err := f.GetGroup("python_data")
    if err != nil {
        fmt.Printf("Failed to get group: %v\n", err)
        os.Exit(1)
    }
    defer g.Close()
    
    // 读取属性
    attr, err := g.Attributes().Get("creator")
    if err != nil {
        fmt.Printf("Failed to get attribute: %v\n", err)
        os.Exit(1)
    }
    var creator string
    err = attr.Read(&creator)
    if err != nil {
        fmt.Printf("Failed to read attribute: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Creator:", creator)
    
    // 读取数据集
    dset, err := g.GetDataset("values")
    if err != nil {
        fmt.Printf("Failed to get dataset: %v\n", err)
        os.Exit(1)
    }
    defer dset.Close()
    
    var result []int32
    err = dset.Read(&result)
    if err != nil {
        fmt.Printf("Failed to read data: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Values:", result)
}
```