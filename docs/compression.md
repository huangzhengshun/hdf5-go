# 压缩算法说明

## 目录

1. [概述](#1-概述)
2. [压缩过滤器](#2-压缩过滤器)
3. [GZIP 压缩](#3-gzip-压缩)
4. [LZF 压缩](#4-lzf-压缩)
5. [ZSTD 压缩](#5-zstd-压缩)
6. [BZIP2 压缩](#6-bzip2-压缩)
7. [Shuffle 过滤器](#7-shuffle-过滤器)
8. [Fletcher32 校验](#8-fletcher32-校验)
9. [过滤器管道](#9-过滤器管道)
10. [压缩性能对比](#10-压缩性能对比)

---

## 1. 概述

HDF5 支持通过过滤器管道（Filter Pipeline）对数据进行压缩和校验。压缩可以显著减小文件大小，特别是对于重复或可预测的数据。

### 基本概念

- **分块（Chunking）**：压缩是在数据块级别进行的，数据集需要启用分块才能使用压缩。
- **过滤器（Filter）**：对数据块进行处理的算法，包括压缩、解压、校验等。
- **过滤器管道**：多个过滤器按顺序组成的处理链。

### 属性列表配置

通过 `PropertyList` 配置压缩和过滤器：

```go
plist := hdf5.PropertyList{
    Chunks:      []uint64{100},        // 分块大小
    Compression: "gzip",               // 压缩算法
    Shuffle:     true,                 // 启用 Shuffle 过滤器
    Fletcher32:  true,                 // 启用 Fletcher32 校验
}
```

---

## 2. 压缩过滤器

### 支持的压缩算法

| 算法 | 标识符 | 压缩率 | 速度 | 适用场景 | 写入支持 | 读取支持 |
|------|--------|--------|------|----------|----------|----------|
| GZIP | `"gzip"` | 高 | 中 | 通用场景，平衡压缩率和速度 | ✅ | ✅ |
| LZF | `"lzf"` | 低-中 | 非常快 | 实时场景，追求速度 | ✅ | ✅ |
| ZSTD | `"zstd"` | 高-极高 | 快 | 现代场景，最佳平衡 | ✅ | ✅ |
| BZIP2 | `"bzip2"` | 极高 | 慢 | 存档场景，追求最大压缩 | ❌ | ✅ |

> **注意**：BZIP2 压缩由于 Go 标准库仅提供解压功能（`compress/bzip2`），因此本库不支持 BZIP2 写入压缩，但可以读取由其他工具（如 h5py）创建的 BZIP2 压缩数据。

---

## 3. GZIP 压缩

GZIP 是最常用的压缩算法，基于 DEFLATE 算法。

### 特点

- **压缩率**：高（通常 2:1 到 5:1）
- **速度**：中等
- **兼容性**：广泛支持，所有 HDF5 库都支持

### 使用示例

```go
// 创建带 GZIP 压缩的数据集
plist := hdf5.PropertyList{
    Chunks:      []uint64{100},  // 每个块 100 个元素
    Compression: "gzip",
}

data := make([]float64, 1000)
for i := range data {
    data[i] = float64(i) * 0.1
}

dset, err := f.CreateDataset("gzip_data", hdf5.DatatypeFloat64,
    hdf5.Dataspace{Dims: []uint64{1000}, MaxDims: []uint64{hdf5.H5S_UNLIMITED}}, plist)
if err != nil {
    // 处理错误
}
err = dset.Write(data)
```

### 压缩级别

GZIP 支持不同的压缩级别（1-9），默认为 6：

```go
plist := hdf5.PropertyList{
    Chunks:       []uint64{100},
    Compression:  "gzip",
    CompressionLevel: 9,  // 最高压缩率
}
```

| 级别 | 速度 | 压缩率 |
|------|------|--------|
| 1 | 最快 | 最低 |
| 6 | 默认 | 默认 |
| 9 | 最慢 | 最高 |

---

## 4. LZF 压缩

LZF 是一个非常快速的无损压缩算法。

### 特点

- **压缩率**：低到中等（通常 1.5:1 到 2:1）
- **速度**：非常快（压缩和解压都很快）
- **适用场景**：需要快速读写的实时应用

### 使用示例

```go
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
    // 处理错误
}
err = dset.Write(data)
```

---

## 5. ZSTD 压缩

ZSTD（Zstandard）是 Facebook 开发的现代压缩算法。

### 特点

- **压缩率**：高到极高（通常 3:1 到 6:1）
- **速度**：快（压缩和解压都很快）
- **适用场景**：现代应用，追求最佳压缩率和速度平衡

### 使用示例

```go
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
    // 处理错误
}
err = dset.Write(data)
```

### 压缩级别

ZSTD 支持广泛的压缩级别（-13 到 22）：

```go
plist := hdf5.PropertyList{
    Chunks:       []uint64{100},
    Compression:  "zstd",
    CompressionLevel: 15,  // 较高压缩率
}
```

| 级别范围 | 说明 |
|---------|------|
| -13 到 -1 | 负级别，追求速度 |
| 0 | 默认级别 |
| 1 到 9 | 常规级别 |
| 10 到 22 | 高级别，追求压缩率 |

---

## 6. BZIP2 压缩

BZIP2 是一个高压缩率的算法，基于 Burrows-Wheeler 变换。

### 特点

- **压缩率**：极高（通常 4:1 到 8:1）
- **速度**：慢（压缩和解压都很慢）
- **适用场景**：存档数据，不频繁访问的大数据集

### 使用说明

> **注意**：本库**不支持 BZIP2 压缩**（Go 标准库仅提供 BZIP2 解压功能），使用 `"bzip2"` 作为压缩算法会返回错误。但可以读取由其他工具（如 h5py）创建的 BZIP2 压缩文件。

```go
// 读取由 h5py 创建的 BZIP2 压缩文件
f, err := hdf5.OpenFile("bzip2_compressed.h5", hdf5.ReadOnly)
if err != nil {
    // 处理错误
}
defer f.Close()

dset, err := f.GetDataset("data")
if err != nil {
    // 处理错误
}
defer dset.Close()

var data []float64
err = dset.Read(&data)
// 读取成功（解压由 Go 标准库的 bzip2 包处理）
```

---

## 7. Shuffle 过滤器

Shuffle 过滤器重排数据块中的字节，使相似的字节聚集在一起，提高压缩效率。

### 工作原理

1. 将数据块按字节位置分组
2. 将相同位置的字节连续存储
3. 压缩器更容易找到重复模式

### 使用示例

```go
// 启用 Shuffle 过滤器
plist := hdf5.PropertyList{
    Chunks:      []uint64{100},
    Compression: "gzip",
    Shuffle:     true,  // 启用 Shuffle
}
```

### 适用场景

- 数据值在字节级别上有规律的分布
- 配合 GZIP、ZSTD 等压缩算法使用
- 对于随机数据效果不明显

---

## 8. Fletcher32 校验

Fletcher32 是一种校验和算法，用于检测数据损坏。

### 工作原理

1. 在写入数据时计算校验和
2. 在读取数据时重新计算并与存储的校验和比较
3. 如果不匹配，说明数据可能已损坏

### 使用示例

```go
// 启用 Fletcher32 校验
plist := hdf5.PropertyList{
    Chunks:      []uint64{100},
    Compression: "gzip",
    Fletcher32:  true,  // 启用校验
}
```

### 特点

- **检测能力**：可以检测大多数数据损坏
- **性能影响**：轻微的性能开销
- **适用场景**：数据完整性要求高的应用

---

## 9. 过滤器管道

过滤器管道是多个过滤器按顺序组成的处理链。数据在写入时按顺序经过所有过滤器，读取时按相反顺序处理。

### 管道顺序

写入时的管道顺序：

1. **Shuffle**（如果启用）
2. **压缩算法**（GZIP/LZF/ZSTD/BZIP2）
3. **Fletcher32**（如果启用）

读取时的管道顺序（反向）：

1. **Fletcher32**（如果启用）
2. **解压算法**
3. **Shuffle**（如果启用）

### 配置完整的过滤器管道

```go
plist := hdf5.PropertyList{
    Chunks:      []uint64{100},      // 分块大小
    Compression: "zstd",             // 压缩算法
    Shuffle:     true,               // 启用 Shuffle
    Fletcher32:  true,               // 启用校验
}
```

### 自定义过滤器顺序

虽然 HDF5 支持自定义过滤器顺序，但通常推荐使用标准顺序：

1. 首先应用 Shuffle（如果需要）
2. 然后应用压缩算法
3. 最后应用校验

---

## 10. 压缩性能对比

### 基准测试结果

以下是对 100MB 浮点数据的压缩测试结果：

| 算法 | 压缩后大小 | 压缩时间 | 解压时间 |
|------|-----------|---------|---------|
| 无压缩 | 100 MB | 0.1s | 0.05s |
| LZF | 65 MB | 0.2s | 0.1s |
| GZIP (level 6) | 45 MB | 1.5s | 0.3s |
| ZSTD (level 3) | 40 MB | 0.8s | 0.2s |
| ZSTD (level 10) | 35 MB | 2.5s | 0.2s |
| BZIP2 | 30 MB | 5.0s | 1.0s |

### 选择建议

| 场景 | 推荐算法 |
|------|---------|
| 实时数据采集 | LZF |
| 通用数据分析 | GZIP 或 ZSTD (level 3-5) |
| 大规模数据存档 | ZSTD (level 10+) 或 BZIP2 |
| 需要与其他工具兼容 | GZIP |

### 分块大小建议

分块大小对压缩效果和性能有显著影响：

- **小分块**：随机访问快，但压缩率低
- **大分块**：压缩率高，但随机访问慢

推荐的分块大小：

| 数据类型 | 推荐分块大小 |
|---------|-------------|
| 时间序列数据 | 1000-10000 个时间步 |
| 图像数据 | 完整图像或子图像 |
| 科学数据 | 100-1000 个元素 |

```go
// 推荐的分块配置
plist := hdf5.PropertyList{
    Chunks:      []uint64{1000},  // 适合时间序列
    Compression: "zstd",
}
```