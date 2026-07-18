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

HDF5 Go 库提供了结构化的错误处理机制，所有 API 操作都返回 `error` 类型。库中预定义了多种错误变量（如 `ErrClosedDataset`、`ErrInvalidData`），同时提供了 `HDF5Error` 结构体用于携带更详细的错误信息（如错误码、位置、堆栈等）。

### 错误处理流程

```go
f, err := hdf5.OpenFile("test.h5", hdf5.Create)
if err != nil {
    // 检查是否为 HDF5Error
    if hdf5.IsHDF5Error(err) {
        // 获取错误码
        code := hdf5.GetErrorCode(err)
        fmt.Printf("Error code: %d\n", code)
        // 使用 %+v 输出详细信息（位置、堆栈等）
        fmt.Printf("Error details: %+v\n", err)
    }
    // 检查是否为预定义错误
    if errors.Is(err, hdf5.ErrClosedFile) {
        fmt.Println("File is closed")
    }
    return
}
```

---

## 2. 错误类型

### 2.1 HDF5Error 结构体

```go
type HDF5Error struct {
    Code     ErrorCode  // 错误码
    Message  string     // 错误消息
    Location string     // 发生位置（文件:行号）
    Stack    []string   // 调用堆栈
    Cause    error      // 原始错误（可选）
}
```

### 2.2 错误方法

```go
// 获取错误消息
func (e *HDF5Error) Error() string

// 获取原始错误
func (e *HDF5Error) Unwrap() error

// 格式化输出（支持 %+v 获取详细信息）
func (e *HDF5Error) Format(f fmt.State, c rune)
```

### 2.3 辅助函数

```go
// 判断是否为 HDF5Error
func IsHDF5Error(err error) bool

// 获取错误码（如果不是 HDF5Error 返回 ErrorCodeError）
func GetErrorCode(err error) ErrorCode

// 包装错误，添加上下文信息
func WrapError(err error, message string) error

// 创建新的 HDF5Error
func NewError(code ErrorCode, message string) *HDF5Error
func NewErrorWithCause(code ErrorCode, message string, cause error) *HDF5Error
```

---

## 3. 错误码

### 3.1 标准错误码

```go
type ErrorCode int

const (
    ErrorCodeOK             ErrorCode = 0   // 成功
    ErrorCodeError          ErrorCode = -1  // 通用错误
    ErrorCodeBadID          ErrorCode = -2  // 无效 ID
    ErrorCodeBadParam       ErrorCode = -3  // 无效参数
    ErrorCodeBadType        ErrorCode = -4  // 无效类型
    ErrorCodeBadAlloc       ErrorCode = -5  // 内存分配失败
    ErrorCodeBusy           ErrorCode = -6  // 资源忙
    ErrorCodeCantOpenFile   ErrorCode = -7  // 无法打开文件
    ErrorCodeCantCreateFile ErrorCode = -8  // 无法创建文件
    ErrorCodeCantRead       ErrorCode = -9  // 无法读取
    ErrorCodeCantWrite      ErrorCode = -10 // 无法写入
    ErrorCodeCantSeek       ErrorCode = -11 // 无法定位
    ErrorCodeEndOfFile      ErrorCode = -12 // 文件结束
    ErrorCodeCorruptFile    ErrorCode = -13 // 文件损坏
    ErrorCodeUnsupported    ErrorCode = -14 // 操作不支持
    ErrorCodeOverflow       ErrorCode = -15 // 溢出
    ErrorCodeUnderflow      ErrorCode = -16 // 下溢
)
```

### 3.2 常用错误变量

库中预定义了以下错误变量，可以通过 `errors.Is()` 进行匹配：

| 错误变量 | 错误消息 | 说明 |
|---------|---------|------|
| `ErrClosedGroup` | `hdf5: group is closed` | 组已关闭 |
| `ErrClosedFile` | `hdf5: file is closed` | 文件已关闭 |
| `ErrClosedDataset` | `hdf5: dataset is closed` | 数据集已关闭 |
| `ErrInvalidData` | `hdf5: invalid data type` | 无效数据类型 |
| `ErrInvalidDatatype` | `hdf5: invalid datatype` | 无效数据类型 |
| `ErrInvalidDataspace` | `hdf5: invalid dataspace` | 无效数据空间 |
| `ErrInvalidLayout` | `hdf5: invalid layout` | 无效存储布局 |
| `ErrNotFound` | `hdf5: object not found` | 对象未找到 |
| `ErrReadOnly` | `hdf5: file is read-only` | 文件只读 |
| `ErrWriteOnly` | `hdf5: write-only` | 仅写模式 |
| `ErrUnsupported` | `hdf5: operation not supported` | 操作不支持 |
| `ErrInternal` | `hdf5: internal error` | 内部错误 |
| `ErrExternalLink` | `hdf5: external links are not supported` | 外部链接不支持 |
| `ErrInvalidSelection` | `hdf5: invalid selection` | 无效选择 |
| `ErrInvalidArgument` | `hdf5: invalid argument` | 无效参数 |
| `ErrIO` | `hdf5: I/O error` | I/O 错误 |
| `ErrMemory` | `hdf5: memory allocation error` | 内存分配错误 |
| `ErrTimeout` | `hdf5: timeout` | 超时 |
| `ErrDeadlock` | `hdf5: deadlock detected` | 检测到死锁 |
| `ErrBusy` | `hdf5: resource is busy` | 资源忙 |
| `ErrCorrupt` | `hdf5: file is corrupt` | 文件损坏 |
| `ErrIncompatible` | `hdf5: incompatible format` | 格式不兼容 |
| `ErrOverflow` | `hdf5: overflow` | 溢出 |
| `ErrUnderflow` | `hdf5: underflow` | 下溢 |

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
    // 检查是否为 HDF5Error
    if hdf5.IsHDF5Error(err) {
        code := hdf5.GetErrorCode(err)
        switch code {
        case hdf5.ErrorCodeCantOpenFile:
            fmt.Println("Failed to open file")
        case hdf5.ErrorCodeCorruptFile:
            fmt.Println("File is corrupt")
        default:
            fmt.Printf("HDF5 error (code=%d): %v\n", code, err)
        }
    } else {
        fmt.Printf("Non-HDF5 error: %v\n", err)
    }
    return
}
```

### 4.4 使用 errors.Is 匹配预定义错误

```go
err := dset.Read(&result)
if err != nil {
    switch {
    case errors.Is(err, hdf5.ErrClosedDataset):
        fmt.Println("Dataset is closed, please reopen it")
    case errors.Is(err, hdf5.ErrInvalidData):
        fmt.Println("Invalid data type, check the dataset type")
    case errors.Is(err, hdf5.ErrIO):
        fmt.Println("I/O error occurred")
    default:
        fmt.Printf("Other error: %v\n", err)
    }
}
```

### 4.5 错误包装

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

### 4.6 日志记录

使用标准日志包记录错误：

```go
import "log"

f, err := hdf5.OpenFile("test.h5", hdf5.Create)
if err != nil {
    log.Printf("ERROR: Failed to open file: %v", err)
    if hdf5.IsHDF5Error(err) {
        // 使用 %+v 格式化输出详细信息（包括位置和堆栈）
        log.Printf("ERROR: Details: %+v", err)
    }
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
Error: hdf5: bzip2 compression is not supported (standard library only provides decompression)
```

**可能原因：**
- 压缩算法名称拼写错误
- 未启用分块
- 使用了不支持的压缩算法（如 BZIP2 写入压缩）

**解决方案：**

```go
// 确保启用分块
plist := hdf5.PropertyList{
    Chunks:      []uint64{100},  // 必须设置分块
    Compression: "gzip",         // 正确的算法名称
}

// 支持的压缩算法：
// - 写入和读取：gzip, lzf, zstd
// - 仅读取：bzip2（不支持写入）
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

### 5.7 "dataset is closed" 错误

**错误信息：**
```
Error: hdf5: dataset is closed
```

**可能原因：**
- 在调用 `dset.Close()` 后再次使用该数据集对象
- 并发场景下数据集被另一个 goroutine 关闭

**解决方案：**

```go
dset, err := f.GetDataset("data")
if err != nil {
    return err
}
defer dset.Close()  // 使用 defer 确保只关闭一次

// 在 Close 之后不要再访问 dset
err = dset.Read(&result)  // 正确
// dset.Close()  // defer 会在函数返回时自动调用
```

### 5.8 "slice bounds out of range" 错误（已修复）

**错误信息：**
```
panic: runtime error: slice bounds out of range [:16] with capacity 10
```

**历史原因（已在最新版本中修复）：**
- `ReadSlice` 方法对分块（chunked）数据集调用 `readChunked()` 时，会尝试将原始字节数据解码到 `[]byte` 类型，导致切片越界

**修复说明：**

新增 `readChunkedRaw()` 方法返回原始字节数据，`ReadSlice` 现已正确处理分块数据集。如果在使用旧版本时遇到此错误，请升级到最新版本。

```go
// 现在可以安全地在分块数据集上使用 ReadSlice
sel := hdf5.Selection{
    Hyperslabs: []hdf5.Hyperslab{
        {Start: []uint64{5}, Count: []uint64{10}},
    },
}
var result []int32
err := dset.ReadSlice(&result, sel)  // 现在可以正常工作
```

### 5.9 字符串属性截断问题（已修复）

**症状：**
- 设置的字符串属性在读取后被截断为 64 字符

**历史原因（已在最新版本中修复）：**
- 字符串属性大小被硬编码为 64 字节，超过 64 字符的字符串会被截断

**修复说明：**

`attribute.go` 中现在根据实际字符串长度动态计算 `Size` 和 `StringSize`。空字符串最小为 1 字节。

```go
// 现在可以设置任意长度的字符串属性
longStr := "This is a very long string that exceeds 64 characters..."
err := g.Attributes().Set("description", longStr)  // 现在可以正常工作

// 读取时不会被截断
attr, _ := g.Attributes().Get("description")
var readStr string
attr.Read(&readStr)
// readStr == longStr
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
    if hdf5.IsHDF5Error(err) {
        // 使用 %+v 输出详细信息（包括位置和堆栈）
        fmt.Printf("Full details: %+v\n", err)
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