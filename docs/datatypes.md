# 数据类型说明

## 目录

1. [概述](#1-概述)
2. [基本数据类型](#2-基本数据类型)
3. [复合数据类型](#3-复合数据类型)
4. [枚举数据类型](#4-枚举数据类型)
5. [数组数据类型](#5-数组数据类型)
6. [变长数据类型](#6-变长数据类型)
7. [字符串数据类型](#7-字符串数据类型)
8. [引用数据类型](#8-引用数据类型)
9. [数据类型转换](#9-数据类型转换)
10. [数据类型兼容性](#10-数据类型兼容性)

---

## 1. 概述

HDF5 Go 库支持多种 HDF5 数据类型，包括基本数值类型、复合类型、枚举类型、数组类型、变长类型等。每种数据类型都对应一个 `Datatype` 结构体，用于描述数据的存储格式和属性。

### Datatype 结构体

```go
type Datatype struct {
    Class        DatatypeClass      // 数据类型类别
    Size         uint32             // 数据类型大小（字节）
    ByteOrder    ByteOrder          // 字节序
    Sign         Sign               // 符号类型（仅整数）
    FloatBits    uint8              // 浮点位数
    Fields       []CompoundField    // 复合类型字段
    BaseType     *Datatype          // 基础类型（枚举、数组、变长）
    ArrayDims    []uint64           // 数组维度
    EnumMembers  []EnumMember       // 枚举成员
    StringLength uint32             // 字符串长度
}
```

---

## 2. 基本数据类型

### 2.1 整数类型

| 数据类型常量 | 对应 Go 类型 | 大小（字节） | 符号 |
|-------------|-------------|-------------|------|
| `DatatypeInt8` | `int8` | 1 | 有符号 |
| `DatatypeInt16` | `int16` | 2 | 有符号 |
| `DatatypeInt32` | `int32` | 4 | 有符号 |
| `DatatypeInt64` | `int64` | 8 | 有符号 |
| `DatatypeUint8` | `uint8` | 1 | 无符号 |
| `DatatypeUint16` | `uint16` | 2 | 无符号 |
| `DatatypeUint32` | `uint32` | 4 | 无符号 |
| `DatatypeUint64` | `uint64` | 8 | 无符号 |

**示例：**

```go
// 创建 int32 数据集
dset, err := f.CreateDataset("int_data", hdf5.DatatypeInt32,
    hdf5.Dataspace{Dims: []uint64{10}}, hdf5.PropertyList{})

// 写入数据
data := []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
err = dset.Write(data)

// 读取数据
var result []int32
err = dset.Read(&result)
```

### 2.2 浮点类型

| 数据类型常量 | 对应 Go 类型 | 大小（字节） | 精度 |
|-------------|-------------|-------------|------|
| `DatatypeFloat32` | `float32` | 4 | 单精度 |
| `DatatypeFloat64` | `float64` | 8 | 双精度 |

**示例：**

```go
// 创建 float64 数据集
dset, err := f.CreateDataset("float_data", hdf5.DatatypeFloat64,
    hdf5.Dataspace{Dims: []uint64{10}}, hdf5.PropertyList{})

// 写入数据
data := []float64{1.1, 2.2, 3.3, 4.4, 5.5}
err = dset.Write(data)

// 读取数据
var result []float64
err = dset.Read(&result)
```

### 2.3 字节序

HDF5 支持多种字节序：

| 字节序常量 | 说明 |
|-----------|------|
| `ByteOrderBigEndian` | 大端序（高位字节在前） |
| `ByteOrderLittleEndian` | 小端序（低位字节在前） |
| `ByteOrderNative` | 本机字节序 |
| `ByteOrderVAX` | VAX 字节序 |

**示例：**

```go
// 创建自定义字节序的整数类型
dtype := hdf5.Datatype{
    Class:     hdf5.DatatypeInteger,
    Size:      4,
    ByteOrder: hdf5.ByteOrderBigEndian,
    Sign:      hdf5.Signed,
}
```

---

## 3. 复合数据类型

复合数据类型（Compound Datatype）类似于 Go 中的结构体，可以包含多个字段，每个字段有自己的名称、偏移量和数据类型。

### 3.1 定义复合类型

```go
compoundDtype := hdf5.Datatype{
    Class: hdf5.DatatypeCompound,
    Size:  20, // 总大小
    Fields: []hdf5.CompoundField{
        {Name: "ID", Offset: 0, Datatype: hdf5.DatatypeInt32},
        {Name: "Name", Offset: 4, Datatype: hdf5.DatatypeString},
        {Name: "Value", Offset: 16, Datatype: hdf5.DatatypeFloat64},
    },
}
```

### 3.2 CompoundField 结构体

```go
type CompoundField struct {
    Name     string   // 字段名称
    Offset   uint32   // 字段偏移量（字节）
    Datatype Datatype // 字段数据类型
}
```

### 3.3 使用复合类型

```go
// 定义 Go 结构体
type MyStruct struct {
    ID    int32
    Name  string
    Value float64
}

// 创建复合数据集
dset, err := f.CreateDataset("compound_data", compoundDtype,
    hdf5.Dataspace{Dims: []uint64{3}}, hdf5.PropertyList{})

// 写入数据
data := []MyStruct{
    {1, "Alice", 100.5},
    {2, "Bob", 200.75},
    {3, "Charlie", 300.0},
}
err = dset.Write(data)

// 读取数据
var result []MyStruct
err = dset.Read(&result)
```

### 3.4 嵌套复合类型

复合类型的字段可以是另一个复合类型：

```go
// 定义内嵌结构体
type Point struct {
    X float64
    Y float64
}

type Record struct {
    ID    int32
    Coord Point
}

// 创建嵌套复合类型
pointDtype := hdf5.Datatype{
    Class: hdf5.DatatypeCompound,
    Size:  16,
    Fields: []hdf5.CompoundField{
        {Name: "X", Offset: 0, Datatype: hdf5.DatatypeFloat64},
        {Name: "Y", Offset: 8, Datatype: hdf5.DatatypeFloat64},
    },
}

recordDtype := hdf5.Datatype{
    Class: hdf5.DatatypeCompound,
    Size:  20,
    Fields: []hdf5.CompoundField{
        {Name: "ID", Offset: 0, Datatype: hdf5.DatatypeInt32},
        {Name: "Coord", Offset: 4, Datatype: pointDtype},
    },
}
```

---

## 4. 枚举数据类型

枚举数据类型（Enum Datatype）基于整数类型，为整数值赋予符号名称。

### 4.1 定义枚举类型

```go
enumDtype := hdf5.Datatype{
    Class:     hdf5.DatatypeEnum,
    Size:      4,
    BaseType:  &hdf5.DatatypeInt32,
    EnumMembers: []hdf5.EnumMember{
        {Name: "Low", Value: 0},
        {Name: "Medium", Value: 1},
        {Name: "High", Value: 2},
    },
}
```

### 4.2 EnumMember 结构体

```go
type EnumMember struct {
    Name  string // 枚举成员名称
    Value int64  // 枚举成员值
}
```

### 4.3 使用枚举类型

```go
// 创建枚举数据集
dset, err := f.CreateDataset("enum_data", enumDtype,
    hdf5.Dataspace{Dims: []uint64{5}}, hdf5.PropertyList{})

// 写入数据（使用底层整数值）
data := []int32{0, 1, 2, 1, 0}
err = dset.Write(data)

// 读取数据
var result []int32
err = dset.Read(&result)
```

---

## 5. 数组数据类型

数组数据类型（Array Datatype）表示固定大小的数组，每个元素具有相同的基础类型。

### 5.1 定义数组类型

```go
// 定义 int32[3] 数组类型
arrayDtype := hdf5.Datatype{
    Class:     hdf5.DatatypeArray,
    Size:      12, // 3 * 4 = 12 字节
    BaseType:  &hdf5.DatatypeInt32,
    ArrayDims: []uint64{3},
}

// 定义 float64[2][3] 二维数组类型
twoDArrayDtype := hdf5.Datatype{
    Class:     hdf5.DatatypeArray,
    Size:      48, // 2 * 3 * 8 = 48 字节
    BaseType:  &hdf5.DatatypeFloat64,
    ArrayDims: []uint64{2, 3},
}
```

### 5.2 使用数组类型

```go
// 创建数组数据集
dset, err := f.CreateDataset("array_data", arrayDtype,
    hdf5.Dataspace{Dims: []uint64{2}}, hdf5.PropertyList{})

// 写入数据（扁平化数组）
data := []int32{1, 2, 3, 4, 5, 6}
err = dset.Write(data)

// 读取数据
var result []int32
err = dset.Read(&result)
// result = [1, 2, 3, 4, 5, 6]
```

---

## 6. 变长数据类型

变长数据类型（VarLength Datatype）表示长度可变的数组，每个元素的长度可以不同。

### 6.1 定义变长类型

```go
// 定义变长 int32 数组类型
varlenDtype := hdf5.Datatype{
    Class:    hdf5.DatatypeVarLength,
    BaseType: &hdf5.DatatypeInt32,
}

// 定义变长字符串类型
varlenStringDtype := hdf5.Datatype{
    Class:    hdf5.DatatypeVarLength,
    BaseType: &hdf5.DatatypeString,
}
```

### 6.2 使用变长类型

```go
// 创建变长数据集
dset, err := f.CreateDataset("varlen_data", varlenDtype,
    hdf5.Dataspace{Dims: []uint64{2}}, hdf5.PropertyList{})

// 写入变长数据
data := [][]int32{
    {1, 2, 3},
    {4, 5, 6, 7, 8},
}
err = dset.Write(data)

// 读取数据
var result [][]int32
err = dset.Read(&result)
// result[0] = [1, 2, 3]
// result[1] = [4, 5, 6, 7, 8]
```

---

## 7. 字符串数据类型

字符串数据类型（String Datatype）用于存储文本数据。

### 7.1 固定长度字符串

```go
// 创建固定长度字符串类型（长度为 20）
fixedStringDtype := hdf5.Datatype{
    Class:        hdf5.DatatypeString,
    Size:         20,
    StringLength: 20,
}

// 创建字符串数据集
dset, err := f.CreateDataset("fixed_strings", fixedStringDtype,
    hdf5.Dataspace{Dims: []uint64{3}}, hdf5.PropertyList{})

// 写入数据
data := []string{"Hello", "World", "HDF5"}
err = dset.Write(data)

// 读取数据
var result []string
err = dset.Read(&result)
```

### 7.2 变长字符串

```go
// 创建变长字符串类型
varlenStringDtype := hdf5.Datatype{
    Class:    hdf5.DatatypeVarLength,
    BaseType: &hdf5.DatatypeString,
}

// 创建变长字符串数据集
dset, err := f.CreateDataset("varlen_strings", varlenStringDtype,
    hdf5.Dataspace{Dims: []uint64{3}}, hdf5.PropertyList{})

// 写入数据
data := []string{"Short", "Medium length string", "A very long string that can be any length"}
err = dset.Write(data)

// 读取数据
var result []string
err = dset.Read(&result)
```

### 7.3 字符串编码

HDF5 字符串支持多种编码方式：

| 编码类型 | 说明 |
|---------|------|
| `H5T_CSET_ASCII` | ASCII 编码 |
| `H5T_CSET_UTF8` | UTF-8 编码 |

---

## 8. 引用数据类型

引用数据类型用于在数据集中存储对其他对象或区域的引用。

### 8.1 对象引用

对象引用（Object Reference）指向 HDF5 文件中的另一个对象（组或数据集）。

```go
// 创建对象引用类型
refDtype := hdf5.DatatypeObjectReference

// 创建存储引用的数据集
dset, err := f.CreateDataset("references", refDtype,
    hdf5.Dataspace{Dims: []uint64{1}}, hdf5.PropertyList{})

// 获取目标数据集的引用
targetDset, _ := f.GetDataset("target_data")
ref, _ := targetDset.CreateObjectReference()
targetDset.Close()

// 写入引用
err = dset.Write([]hdf5.ObjectReference{ref})

// 读取引用并解引用
var refs []hdf5.ObjectReference
err = dset.Read(&refs)
obj, err := f.Dereference(refs[0])
```

### 8.2 区域引用

区域引用（Region Reference）指向数据集中的特定区域。

```go
// 创建区域引用类型
regionRefDtype := hdf5.DatatypeRegionReference

// 创建存储区域引用的数据集
dset, err := f.CreateDataset("region_refs", regionRefDtype,
    hdf5.Dataspace{Dims: []uint64{1}}, hdf5.PropertyList{})

// 获取目标数据集的区域引用
targetDset, _ := f.GetDataset("target_data")
sel := hdf5.Selection{
    Hyperslabs: []hdf5.Hyperslab{
        {Start: []uint64{0}, Count: []uint64{5}},
    },
}
ref, _ := targetDset.CreateReference(sel)
targetDset.Close()

// 写入区域引用
err = dset.Write([]hdf5.RegionReference{ref})
```

---

## 9. 数据类型转换

### 9.1 自动类型转换

HDF5 Go 库支持一定程度的自动类型转换：

- 整数类型之间可以互相转换（可能会有精度损失）
- 浮点类型之间可以互相转换
- 整数可以转换为浮点类型
- 字符串可以转换为字节数组

### 9.2 显式类型转换

在某些情况下，需要显式指定数据类型：

```go
// 将 float64 数据写入 int32 数据集（会进行截断）
floatData := []float64{1.9, 2.1, 3.5}
intData := make([]int32, len(floatData))
for i, v := range floatData {
    intData[i] = int32(v)
}
err = dset.Write(intData)
```

---

## 10. 数据类型兼容性

### 10.1 h5py 兼容性

| HDF5 Go 类型 | h5py 对应类型 | 兼容性 |
|-------------|--------------|--------|
| `DatatypeInt8` | `numpy.int8` | ✅ |
| `DatatypeInt16` | `numpy.int16` | ✅ |
| `DatatypeInt32` | `numpy.int32` | ✅ |
| `DatatypeInt64` | `numpy.int64` | ✅ |
| `DatatypeUint8` | `numpy.uint8` | ✅ |
| `DatatypeUint16` | `numpy.uint16` | ✅ |
| `DatatypeUint32` | `numpy.uint32` | ✅ |
| `DatatypeUint64` | `numpy.uint64` | ✅ |
| `DatatypeFloat32` | `numpy.float32` | ✅ |
| `DatatypeFloat64` | `numpy.float64` | ✅ |
| `DatatypeCompound` | `numpy.dtype` | ✅ |
| `DatatypeEnum` | `numpy.dtype` | ✅ |
| `DatatypeArray` | `numpy.ndarray` | ✅ |
| `DatatypeVarLength` | `h5py.vlen_dtype` | ✅ |

### 10.2 注意事项

1. **字节序问题**：不同平台的字节序可能不同，建议使用 `ByteOrderNative` 或在读写时进行转换。
2. **字符串长度**：固定长度字符串会在末尾填充零字节，读取时需要正确处理。
3. **复合类型字段顺序**：复合类型的字段必须按照正确的顺序和偏移量定义。
4. **变长类型内存管理**：变长类型在读取后需要正确释放内存。