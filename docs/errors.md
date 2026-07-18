# 错误处理指南

## 目录

1. [概述](#1-概述)
2. [错误类型](#2-错误类型)
3. [错误码](#3-错误码)
4. [错误处理最佳实践](#4-错误处理最佳实践)
5. [常见错误及解决方案](#5-常见错误及解决方案)
6. [调试技巧](#6-调试技巧)
7. [错误恢复策略](#7-错误恢复策略)

---

## 1. 概述

HDF5 Go 库提供了结构化的错误处理机制，所有 API 操作都返回 `error` 类型，可以通过类型断言获取详细的错误信息。

### 错误处理流程

```go
f, err := hdf5.OpenFile("test.h5", hdf5.Create)
if err != nil {
    // 检查错误类型
    if hdf5Err, ok := err.(*hdf5.Error); ok {
        // 获取详细错误信息
        fmt.Printf("Error code: %d\n", hdf5Err.Code)
        fmt.Printf("Error message: %s\n", hdf5Err.Message)
        fmt.Printf("Error stack: %s\n", hdf5Err.Stack)
    }
    return
}
```

---

## 2. 错误类型

### 2.1 Error 结构体

```go
type Error struct {
    Code       ErrorCode      // 错误码
    Message    string         // 错误消息
    Stack      string         // 调用堆栈
    Operation  string         // 操作名称
    ObjectPath string         // 对象路径（可选）
}
```

### 2.2 错误方法

```go
// 获取错误消息
func (e *Error) Error() string

// 获取错误码
func (e *Error) GetCode() ErrorCode

// 获取调用堆栈
func (e *Error) GetStack() string

// 判断是否为特定错误码
func (e *Error) IsCode(code ErrorCode) bool
```

---

## 3. 错误码

### 3.1 文件操作错误

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 1 | `ErrFileNotFound` | 文件不存在 |
| 2 | `ErrFileOpenFailed` | 文件打开失败 |
| 3 | `ErrFileCloseFailed` | 文件关闭失败 |
| 4 | `ErrFileCreateFailed` | 文件创建失败 |
| 5 | `ErrFileLocked` | 文件已被锁定 |
| 6 | `ErrFileCorrupt` | 文件损坏 |

### 3.2 对象操作错误

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 10 | `ErrObjectNotFound` | 对象不存在 |
| 11 | `ErrObjectExists` | 对象已存在 |
| 12 | `ErrObjectTypeMismatch` | 对象类型不匹配 |
| 13 | `ErrObjectInvalid` | 对象无效 |

### 3.3 数据集操作错误

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 20 | `ErrDatasetNotFound` | 数据集不存在 |
| 21 | `ErrDatasetExists` | 数据集已存在 |
| 22 | `ErrDatasetReadOnly` | 数据集只读 |
| 23 | `ErrDatasetResizeFailed` | 数据集调整大小失败 |
| 24 | `ErrDatasetSelectionInvalid` | 数据集选择无效 |

### 3.4 数据类型错误

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 30 | `ErrInvalidDataType` | 无效数据类型 |
| 31 | `ErrDataTypeMismatch` | 数据类型不匹配 |
| 32 | `ErrDataTypeNotSupported` | 数据类型不支持 |
| 33 | `ErrInvalidData` | 无效数据 |

### 3.5 数据空间错误

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 40 | `ErrInvalidDataspace` | 无效数据空间 |
| 41 | `ErrDataspaceMismatch` | 数据空间不匹配 |
| 42 | `ErrDataspaceRankMismatch` | 数据空间维度不匹配 |
| 43 | `ErrDataspaceSizeMismatch` | 数据空间大小不匹配 |

### 3.6 属性操作错误

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 50 | `ErrAttributeNotFound` | 属性不存在 |
| 51 | `ErrAttributeExists` | 属性已存在 |
| 52 | `ErrAttributeInvalid` | 属性无效 |

### 3.7 压缩错误

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 60 | `ErrCompressionFailed` | 压缩失败 |
| 61 | `ErrDecompressionFailed` | 解压失败 |
| 62 | `ErrCompressionNotSupported` | 压缩算法不支持 |

### 3.8 内存错误

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 70 | `ErrMemoryAllocationFailed` | 内存分配失败 |
| 71 | `ErrMemoryAccessError` | 内存访问错误 |

### 3.9 IO 错误

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 80 | `ErrIOError` | IO 错误 |
| 81 | `ErrReadFailed` | 读取失败 |
| 82 | `ErrWriteFailed` | 写入失败 |

### 3.10 其他错误

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 99 | `ErrUnknown` | 未知错误 |

---

## 4. 错误处理最佳实践

### 4.1 检查所有错误

始终检查 API 返回的错误，不要忽略：

```go
// 正确做法
f, err := hdf5.OpenFile("test.h5", hdf5.Create)
if err != nil {
    log.Fatalf("Failed to open file: %v", err)
}
defer f.Close()

// 错误做法（不要这样做）
f, _ := hdf5.OpenFile("test.h5", hdf5.Create)
```

### 4.2 使用 defer 确保资源释放

```go
f, err := hdf5.OpenFile("test.h5", hdf5.Create)
if err != nil {
    return err
}
defer f.Close()

g, err := f.CreateGroup("my_group")
if err != nil {
    return err  // f.Close() 会被自动调用
}
defer g.Close()

dset, err := g.CreateDataset("my_data", hdf5.DatatypeInt32,
    hdf5.Dataspace{Dims: []uint64{10}}, hdf5.PropertyList{})
if err != nil {
    return err  // g.Close() 和 f.Close() 会被自动调用
}
defer dset.Close()
```

### 4.3 错误类型断言

使用类型断言获取详细错误信息：

```go
f, err := hdf5.OpenFile("nonexistent.h5", hdf5.ReadOnly)
if err != nil {
    if hdf5Err, ok := err.(*hdf5.Error); ok {
        switch hdf5Err.Code {
        case hdf5.ErrFileNotFound:
            fmt.Println("File does not exist")
        case hdf5.ErrFileOpenFailed:
            fmt.Println("Failed to open file")
        default:
            fmt.Printf("Unexpected error: %v\n", hdf5Err)
        }
    } else {
        fmt.Printf("Non-HDF5 error: %v\n", err)
    }
    return
}
```

### 4.4 错误包装

使用 `fmt.Errorf` 包装错误，添加上下文信息：

```go
func processFile(filename string) error {
    f, err := hdf5.OpenFile(filename, hdf5.ReadOnly)
    if err != nil {
        return fmt.Errorf("failed to open file %q: %w", filename, err)
    }
    defer f.Close()
    
    // ... 处理逻辑 ...
    
    return nil
}
```

### 4.5 日志记录

使用标准日志包记录错误：

```go
import "log"

f, err := hdf5.OpenFile("test.h5", hdf5.Create)
if err != nil {
    log.Printf("ERROR: Failed to open file: %v", err)
    log.Printf("ERROR: Stack trace: %s", err.(*hdf5.Error).Stack)
    return err
}
```

---

## 5. 常见错误及解决方案

### 5.1 "file not found" 错误

**错误信息：**
```
Error: file not found
```

**可能原因：**
- 文件路径不正确
- 文件被删除或移动
- 权限不足

**解决方案：**

```go
// 检查文件路径
filepath := "path/to/file.h5"

// 使用绝对路径
absPath, err := filepath.Abs(filepath)
if err != nil {
    return err
}

// 检查文件是否存在
_, err = os.Stat(absPath)
if os.IsNotExist(err) {
    // 文件不存在，尝试创建
    f, err := hdf5.OpenFile(absPath, hdf5.Create)
    if err != nil {
        return err
    }
    defer f.Close()
}
```

### 5.2 "dataset not found" 错误

**错误信息：**
```
Error: dataset not found
```

**可能原因：**
- 数据集名称拼写错误
- 数据集在错误的组中
- 数据集已被删除

**解决方案：**

```go
// 使用 Keys() 检查组内的数据集
keys := g.Keys()
fmt.Println("Available keys:", keys)

// 确认路径是否正确
dsetPath := "/group_name/dataset_name"
dset, err := f.GetDataset(dsetPath)
if err != nil {
    fmt.Printf("Dataset %s not found. Available keys: %v\n", dsetPath, keys)
    return err
}
```

### 5.3 "invalid data type" 错误

**错误信息：**
```
Error: invalid data type
```

**可能原因：**
- 复合数据类型的字段定义不正确
- 数据类型枚举值不匹配
- 使用了不支持的数据类型

**解决方案：**

```go
// 检查复合类型定义
compoundDtype := hdf5.Datatype{
    Class: hdf5.DatatypeCompound,
    Size:  12,  // 确保大小正确
    Fields: []hdf5.CompoundField{
        {Name: "ID", Offset: 0, Datatype: hdf5.DatatypeInt32},
        {Name: "Value", Offset: 4, Datatype: hdf5.DatatypeFloat64},
    },
}

// 确认 DatatypeCompound 值为 4（HDF5 规范）
fmt.Printf("DatatypeCompound value: %d\n", hdf5.DatatypeCompound)
```

### 5.4 "data space mismatch" 错误

**错误信息：**
```
Error: data space mismatch
```

**可能原因：**
- 写入的数据大小与数据集大小不匹配
- 选择区域的大小与数据大小不匹配

**解决方案：**

```go
// 确保数据大小匹配
data := []int32{1, 2, 3, 4, 5}
space := hdf5.Dataspace{Dims: []uint64{5}}  // 与数据大小匹配

dset, err := f.CreateDataset("data", hdf5.DatatypeInt32, space, hdf5.PropertyList{})
if err != nil {
    return err
}

// 使用完整写入时，数据大小必须匹配
err = dset.Write(data)  // 正确：5 个元素

// 使用选择写入时，数据大小必须匹配选择大小
sel := hdf5.Selection{
    Hyperslabs: []hdf5.Hyperslab{
        {Start: []uint64{0}, Count: []uint64{3}},
    },
}
err = dset.WriteSlice(data[:3], sel)  // 正确：3 个元素
```

### 5.5 "compression not supported" 错误

**错误信息：**
```
Error: compression not supported
```

**可能原因：**
- 压缩算法名称拼写错误
- 未启用分块
- 使用了不支持的压缩算法

**解决方案：**

```go
// 确保启用分块
plist := hdf5.PropertyList{
    Chunks:      []uint64{100},  // 必须设置分块
    Compression: "gzip",         // 正确的算法名称
}

// 支持的压缩算法：gzip, lzf, zstd, bzip2
```

### 5.6 "dataset resize failed" 错误

**错误信息：**
```
Error: dataset resize failed
```

**可能原因：**
- 数据集创建时未设置 `MaxDims`
- 新维度小于当前维度（不允许缩小）
- 数据类型不支持调整大小

**解决方案：**

```go
// 创建可扩展数据集
space := hdf5.Dataspace{
    Dims:    []uint64{5},
    MaxDims: []uint64{hdf5.H5S_UNLIMITED},  // 设置为无限
}

dset, err := f.CreateDataset("resizeable", hdf5.DatatypeInt32, space, hdf5.PropertyList{})
if err != nil {
    return err
}

// 只能增大，不能缩小
err = dset.Resize([]uint64{10})  // 正确
err = dset.Resize([]uint64{3})   // 错误：缩小不允许
```

---

## 6. 调试技巧

### 6.1 启用调试模式

```go
// 设置调试级别
hdf5.SetDebugLevel(hdf5.DebugLevelVerbose)

// 或者设置环境变量
// export HDF5_DEBUG=verbose
```

### 6.2 打印错误堆栈

```go
f, err := hdf5.OpenFile("test.h5", hdf5.Create)
if err != nil {
    if hdf5Err, ok := err.(*hdf5.Error); ok {
        fmt.Println("Full stack trace:")
        fmt.Println(hdf5Err.Stack)
    }
    return err
}
```

### 6.3 使用 h5dump 检查文件

使用 h5dump 工具查看 HDF5 文件结构：

```bash
# 查看文件结构
h5dump -n test.h5

# 查看数据集内容
h5dump test.h5

# 查看特定数据集
h5dump -d /path/to/dataset test.h5

# 查看属性
h5dump -A test.h5
```

### 6.4 使用 h5py 验证数据

```python
import h5py

# 打开文件
f = h5py.File('test.h5', 'r')

# 列出所有对象
def print_structure(name, obj):
    print(f"{name}: {type(obj)}")
    
f.visititems(print_structure)

# 读取数据
data = f['/path/to/dataset'][:]
print(data)

f.close()
```

### 6.5 日志记录关键操作

```go
import "log"

log.Println("Opening file...")
f, err := hdf5.OpenFile("test.h5", hdf5.Create)
if err != nil {
    log.Printf("Failed to open file: %v", err)
    return err
}
defer f.Close()
log.Println("File opened successfully")

log.Println("Creating dataset...")
dset, err := f.CreateDataset("data", hdf5.DatatypeInt32,
    hdf5.Dataspace{Dims: []uint64{10}}, hdf5.PropertyList{})
if err != nil {
    log.Printf("Failed to create dataset: %v", err)
    return err
}
defer dset.Close()
log.Println("Dataset created successfully")
```

---

## 7. 错误恢复策略

### 7.1 文件损坏恢复

```go
// 尝试修复损坏的文件
func recoverFile(filename string) error {
    f, err := hdf5.OpenFile(filename, hdf5.ReadOnly)
    if err != nil {
        // 文件可能损坏，尝试修复
        fmt.Printf("File may be corrupt: %v\n", err)
        
        // 创建新文件并复制可恢复的数据
        newFile, err := hdf5.OpenFile(filename+".recovered", hdf5.Create)
        if err != nil {
            return fmt.Errorf("failed to create recovery file: %w", err)
        }
        defer newFile.Close()
        
        fmt.Println("Recovery file created. Attempting to recover data...")
        return nil
    }
    defer f.Close()
    
    fmt.Println("File is healthy")
    return nil
}
```

### 7.2 数据备份策略

```go
func backupFile(filename string) error {
    // 创建备份文件
    backupName := filename + ".bak"
    
    // 打开源文件
    src, err := hdf5.OpenFile(filename, hdf5.ReadOnly)
    if err != nil {
        return err
    }
    defer src.Close()
    
    // 创建备份文件
    dst, err := hdf5.OpenFile(backupName, hdf5.Create)
    if err != nil {
        return err
    }
    defer dst.Close()
    
    // 复制根组
    root, err := src.GetGroup("/")
    if err != nil {
        return err
    }
    defer root.Close()
    
    err = root.Copy(dst, "/")
    if err != nil {
        return err
    }
    
    fmt.Printf("Backup created: %s\n", backupName)
    return nil
}
```

### 7.3 原子写入保证

```go
// 使用原子写入保证数据完整性
func atomicWrite(dset hdf5.Dataset, data interface{}) error {
    // 创建临时文件
    tempFile, err := hdf5.CreateMemoryFile()
    if err != nil {
        return err
    }
    defer tempFile.Close()
    
    // 在临时文件中写入数据
    tempDset, err := tempFile.CreateDataset("temp", dset.Datatype(),
        dset.Dataspace(), hdf5.PropertyList{})
    if err != nil {
        return err
    }
    err = tempDset.Write(data)
    if err != nil {
        return err
    }
    tempDset.Close()
    
    // 使用原子写入
    err = dset.WriteAtomic(data)
    if err != nil {
        fmt.Println("Atomic write failed, data may be inconsistent")
        return err
    }
    
    return nil
}
```

### 7.4 错误重试机制

```go
func retryWithBackoff(fn func() error, maxRetries int) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = fn()
        if err == nil {
            return nil
        }
        
        fmt.Printf("Attempt %d/%d failed: %v\n", i+1, maxRetries, err)
        
        if i < maxRetries-1 {
            // 指数退避
            backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
            fmt.Printf("Retrying in %v...\n", backoff)
            time.Sleep(backoff)
        }
    }
    
    return fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}

// 使用示例
err := retryWithBackoff(func() error {
    return dset.Write(data)
}, 3)
if err != nil {
    return err
}
```