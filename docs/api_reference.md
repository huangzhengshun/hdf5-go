# API 参考文档

## 一、文件操作

### 1.1 打开和创建文件

#### OpenFile

```go
func OpenFile(name string, mode FileMode) (*File, error)
```

打开或创建 HDF5 文件。

**参数：**
- `name` - 文件名（字符串）
- `mode` - 文件模式（`ReadOnly`、`ReadWrite`、`Create`、`CreateExclusive`、`Append`）

**返回值：**
- `*File` - 文件对象
- `error` - 错误信息

**示例：**
```go
f, err := hdf5.OpenFile("test.h5", hdf5.Create)
if err != nil {
    // 处理错误
}
defer f.Close()
```

#### CreateMemoryFile

```go
func CreateMemoryFile() (*File, error)
```

创建内存中的 HDF5 文件。

**返回值：**
- `*File` - 文件对象
- `error` - 错误信息

**示例：**
```go
f, err := hdf5.CreateMemoryFile()
if err != nil {
    // 处理错误
}
defer f.Close()
```

#### OpenMmapFile

```go
func OpenMmapFile(name string, mode FileMode) (*File, error)
```

使用内存映射打开 HDF5 文件。

**参数：**
- `name` - 文件名（字符串）
- `mode` - 文件模式

**返回值：**
- `*File` - 文件对象
- `error` - 错误信息

### 1.2 文件管理

#### Close

```go
func (f *File) Close() error
```

关闭文件。

**返回值：**
- `error` - 错误信息

#### Lock / Unlock

```go
func (f *File) Lock() error
func (f *File) Unlock() error
```

锁定/解锁文件，用于并发访问控制。

**返回值：**
- `error` - 错误信息

#### Bytes

```go
func (f *File) Bytes() ([]byte, error)
```

获取文件的字节内容（适用于内存文件）。

**返回值：**
- `[]byte` - 文件字节内容
- `error` - 错误信息

### 1.3 文件空间管理

#### ReclaimFreeSpace

```go
func (f *File) ReclaimFreeSpace() error
```

回收文件中的空闲空间。

**返回值：**
- `error` - 错误信息

#### GetFreeSpaceInfo

```go
func (f *File) GetFreeSpaceInfo() (FreeSpaceInfo, error)
```

获取文件空闲空间信息。

**返回值：**
- `FreeSpaceInfo` - 空闲空间信息结构
- `error` - 错误信息

**FreeSpaceInfo 结构：**
```go
type FreeSpaceInfo struct {
    TotalFreeSpace  uint64  // 总空闲空间
    FreeSpaceBlocks int     // 空闲空间块数
    SmallestBlock   uint64  // 最小块大小
    LargestBlock    uint64  // 最大块大小
}
```

#### CompactFile

```go
func (f *File) CompactFile() error
```

压缩文件，减少文件大小。

**返回值：**
- `error` - 错误信息

#### SetSpaceAllocationStrategy

```go
func (f *File) SetSpaceAllocationStrategy(strategy SpaceAllocationStrategy) error
```

设置空间分配策略。

**参数：**
- `strategy` - 空间分配策略（`SpaceStrategyDefault`、`SpaceStrategyFIFO`、`SpaceStrategyBest`、`SpaceStrategyWorst`）

**返回值：**
- `error` - 错误信息

#### GetUsedSpace / GetTotalSpace

```go
func (f *File) GetUsedSpace() (uint64, error)
func (f *File) GetTotalSpace() (uint64, error)
```

获取已使用空间/总空间大小。

**返回值：**
- `uint64` - 空间大小（字节）
- `error` - 错误信息

## 二、组操作

### 2.1 创建和获取组

#### CreateGroup

```go
func (f *File) CreateGroup(name string) (Group, error)
func (g *group) CreateGroup(name string) (Group, error)
```

创建子组。

**参数：**
- `name` - 组名（字符串）

**返回值：**
- `Group` - 组对象
- `error` - 错误信息

**示例：**
```go
g, err := f.CreateGroup("my_group")
if err != nil {
    // 处理错误
}
defer g.Close()
```

#### GetGroup

```go
func (f *File) GetGroup(name string) (Group, error)
func (g *group) GetGroup(name string) (Group, error)
```

获取子组。

**参数：**
- `name` - 组名（字符串）

**返回值：**
- `Group` - 组对象
- `error` - 错误信息

#### Keys

```go
func (g *group) Keys() []string
```

获取组内所有键名（数据集和子组名称）。

**返回值：**
- `[]string` - 键名列表

**示例：**
```go
keys := g.Keys()
for _, key := range keys {
    fmt.Println(key)
}
```

### 2.2 遍历组

#### Visit

```go
func (g *group) Visit(fn func(string, interface{}) error) error
```

递归访问组内所有对象。

**参数：**
- `fn` - 访问函数，参数为路径和对象

**返回值：**
- `error` - 错误信息

**示例：**
```go
err := g.Visit(func(path string, obj interface{}) error {
    fmt.Println("Path:", path)
    return nil
})
```

#### VisitItems

```go
func (g *group) VisitItems(fn func(name string, obj interface{}) error) error
```

访问组内所有直接项目（非递归）。

**参数：**
- `fn` - 访问函数，参数为名称和对象

**返回值：**
- `error` - 错误信息

### 2.3 组管理

#### Close

```go
func (g *group) Close() error
```

关闭组。

**返回值：**
- `error` - 错误信息

#### Copy / CopyWithOptions

```go
func (g *group) Copy(dest Group, name string) error
func (g *group) CopyWithOptions(dest Group, name string, options CopyOptions) error
```

复制组到目标位置。

**参数：**
- `dest` - 目标组
- `name` - 新名称
- `options` - 复制选项（可选）

**返回值：**
- `error` - 错误信息

#### Move

```go
func (g *group) Move(dest Group, name string) error
```

移动组到目标位置。

**参数：**
- `dest` - 目标组
- `name` - 新名称

**返回值：**
- `error` - 错误信息

#### Delete

```go
func (g *group) Delete(name string) error
```

删除组或数据集。

**参数：**
- `name` - 要删除的对象名称

**返回值：**
- `error` - 错误信息

## 三、数据集操作

### 3.1 创建和获取数据集

#### CreateDataset

```go
func (f *File) CreateDataset(name string, dtype Datatype, space Dataspace, plist PropertyList) (Dataset, error)
func (g *group) CreateDataset(name string, dtype Datatype, space Dataspace, plist PropertyList) (Dataset, error)
```

创建数据集。

**参数：**
- `name` - 数据集名称
- `dtype` - 数据类型（如 `DatatypeInt32`、`DatatypeFloat64`）
- `space` - 数据空间（`Dataspace{Dims: []uint64{...}}`）
- `plist` - 属性列表（可选，用于压缩、分块等）

**返回值：**
- `Dataset` - 数据集对象
- `error` - 错误信息

**示例：**
```go
data := []int32{1, 2, 3, 4, 5}
dset, err := g.CreateDataset("data", hdf5.DatatypeInt32,
    hdf5.Dataspace{Dims: []uint64{5}}, hdf5.PropertyList{})
if err != nil {
    // 处理错误
}
defer dset.Close()
```

#### CreateVirtualDataset

```go
func (g *group) CreateVirtualDataset(name string, layout VirtualLayout, plist PropertyList) (Dataset, error)
```

创建虚拟数据集（VDS）。

**参数：**
- `name` - 数据集名称
- `layout` - 虚拟布局
- `plist` - 属性列表

**返回值：**
- `Dataset` - 数据集对象
- `error` - 错误信息

#### GetDataset

```go
func (f *File) GetDataset(name string) (Dataset, error)
func (g *group) GetDataset(name string) (Dataset, error)
```

获取数据集。

**参数：**
- `name` - 数据集名称

**返回值：**
- `Dataset` - 数据集对象
- `error` - 错误信息

### 3.2 读写操作

#### Read

```go
func (d *dataset) Read(data interface{}) error
```

读取数据集全部数据。

**参数：**
- `data` - 用于接收数据的切片指针

**返回值：**
- `error` - 错误信息

**示例：**
```go
var result []int32
err := dset.Read(&result)
if err != nil {
    // 处理错误
}
```

#### Write

```go
func (d *dataset) Write(data interface{}) error
```

写入数据集全部数据。

**参数：**
- `data` - 要写入的数据

**返回值：**
- `error` - 错误信息

**示例：**
```go
data := []int32{1, 2, 3, 4, 5}
err := dset.Write(data)
if err != nil {
    // 处理错误
}
```

#### ReadSlice

```go
func (d *dataset) ReadSlice(data interface{}, sel Selection) error
```

按选择读取数据（支持点选择、超平面选择、多超平面选择、掩码选择）。

**参数：**
- `data` - 用于接收数据的切片指针
- `sel` - 选择条件

**返回值：**
- `error` - 错误信息

**说明：**

该方法同时支持连续存储（contiguous）和分块存储（chunked）数据集。对于分块数据集，内部使用 `readChunkedRaw()` 读取原始字节数据后再进行切片选择，避免了类型解码冲突。

**示例：**
```go
// 点选择
sel := hdf5.Selection{
    Points: [][]uint64{{1}, {3}, {5}},
}
var result []int32
err := dset.ReadSlice(&result, sel)

// 超平面选择
sel := hdf5.Selection{
    Hyperslabs: []hdf5.Hyperslab{
        {Start: []uint64{0}, Count: []uint64{3}},
    },
}
err = dset.ReadSlice(&result, sel)

// 多超平面选择
sel = hdf5.Selection{
    Hyperslabs: []hdf5.Hyperslab{
        {Start: []uint64{0}, Count: []uint64{3}},
        {Start: []uint64{5}, Count: []uint64{2}},
    },
}
err = dset.ReadSlice(&result, sel)

// 掩码选择
mask := []bool{false, true, false, true, false}
sel = hdf5.Selection{Mask: mask}
err = dset.ReadSlice(&result, sel)
```

#### WriteSlice

```go
func (d *dataset) WriteSlice(data interface{}, sel Selection) error
```

按选择写入数据（支持超平面选择）。

**参数：**
- `data` - 要写入的数据
- `sel` - 选择条件

**返回值：**
- `error` - 错误信息

**示例：**
```go
sel := hdf5.Selection{
    Hyperslabs: []hdf5.Hyperslab{
        {Start: []uint64{5}, Count: []uint64{3}},
    },
}
data := []int32{6, 7, 8}
err := dset.WriteSlice(data, sel)
```

### 3.3 数据集管理

#### Resize

```go
func (d *dataset) Resize(newDims []uint64) error
```

调整数据集大小（仅适用于可扩展数据集）。

**参数：**
- `newDims` - 新的维度大小

**返回值：**
- `error` - 错误信息

**示例：**
```go
// 创建可扩展数据集
space := hdf5.Dataspace{Dims: []uint64{5}, MaxDims: []uint64{hdf5.H5S_UNLIMITED}}
dset, _ := g.CreateDataset("data", hdf5.DatatypeInt32, space, hdf5.PropertyList{})

// 调整大小
err := dset.Resize([]uint64{10})
```

#### Flush

```go
func (d *dataset) Flush() error
```

刷新数据集到磁盘。

**返回值：**
- `error` - 错误信息

#### Close

```go
func (d *dataset) Close() error
```

关闭数据集。

**返回值：**
- `error` - 错误信息

### 3.4 原子操作

#### WriteAtomic

```go
func (d *dataset) WriteAtomic(data interface{}) error
```

原子写入数据集（保证写入的完整性）。

**参数：**
- `data` - 要写入的数据

**返回值：**
- `error` - 错误信息

## 四、属性操作

### 4.1 属性管理

#### Get

```go
func (am *AttributeManager) Get(name string) (Attribute, error)
```

获取属性。

**参数：**
- `name` - 属性名称

**返回值：**
- `Attribute` - 属性对象
- `error` - 错误信息

#### Set

```go
func (am *AttributeManager) Set(name string, data interface{}) error
```

设置属性。

**参数：**
- `name` - 属性名称
- `data` - 属性值（支持基本类型、数组、字符串、字符串切片）

**返回值：**
- `error` - 错误信息

**说明：**

对于字符串属性，库会根据实际字符串长度动态计算存储大小（`Size` 和 `StringSize`），不再使用固定的 64 字节长度，因此支持任意长度的字符串（包括超过 64 字符的长字符串和空字符串）。

**示例：**
```go
err := g.Attributes().Set("description", "This is a test group")
err = g.Attributes().Set("version", int32(1))
err = g.Attributes().Set("values", []float64{1.0, 2.0, 3.0})
err = g.Attributes().Set("long_desc", "This is a very long string that exceeds 64 characters and would have been truncated before the fix")
err = g.Attributes().Set("empty_string", "")
err = g.Attributes().Set("string_list", []string{"apple", "banana", "cherry"})
```

#### Delete

```go
func (am *AttributeManager) Delete(name string) error
```

删除属性。

**参数：**
- `name` - 属性名称

**返回值：**
- `error` - 错误信息

#### Iter

```go
func (am *AttributeManager) Iter() ([]Attribute, error)
```

迭代所有属性。

**返回值：**
- `[]Attribute` - 属性列表
- `error` - 错误信息

#### Copy

```go
func (am *AttributeManager) Copy(dest *AttributeManager, name string) error
```

复制属性到目标对象。

**参数：**
- `dest` - 目标属性管理器
- `name` - 属性名称

**返回值：**
- `error` - 错误信息

### 4.2 属性读写

#### Read

```go
func (a *attribute) Read(data interface{}) error
```

读取属性数据。

**参数：**
- `data` - 用于接收数据的变量指针

**返回值：**
- `error` - 错误信息

**示例：**
```go
attr, _ := g.Attributes().Get("description")
var desc string
err := attr.Read(&desc)
```

#### Write

```go
func (a *attribute) Write(data interface{}) error
```

写入属性数据。

**参数：**
- `data` - 要写入的数据

**返回值：**
- `error` - 错误信息

## 五、链接操作

### 5.1 创建链接

#### CreateLink

```go
func (f *File) CreateLink(targetPath, linkName string) error
func (g *group) CreateLink(targetPath, linkName string) error
```

创建软链接。

**参数：**
- `targetPath` - 目标路径
- `linkName` - 链接名称

**返回值：**
- `error` - 错误信息

**示例：**
```go
err := g.CreateLink("/data/original", "my_link")
```

#### CreateLinkWithOptions

```go
func (g *group) CreateLinkWithOptions(targetPath, linkName string, options LinkCreateOptions) error
```

带选项创建软链接。

**参数：**
- `targetPath` - 目标路径
- `linkName` - 链接名称
- `options` - 链接创建选项

**返回值：**
- `error` - 错误信息

#### CreateExternalLink

```go
func (f *File) CreateExternalLink(externalFile, targetPath, linkName string) error
func (g *group) CreateExternalLink(externalFile, targetPath, linkName string) error
```

创建外部链接（指向其他 HDF5 文件）。

**参数：**
- `externalFile` - 外部文件名
- `targetPath` - 目标路径
- `linkName` - 链接名称

**返回值：**
- `error` - 错误信息

## 六、引用操作

### 6.1 创建引用

#### CreateReference

```go
func (d *dataset) CreateReference(sel Selection) (RegionReference, error)
```

创建区域引用（指向数据集中的特定区域）。

**参数：**
- `sel` - 选择条件

**返回值：**
- `RegionReference` - 区域引用
- `error` - 错误信息

#### CreateObjectReference

```go
func (d *dataset) CreateObjectReference() (ObjectReference, error)
func (g *group) CreateReference() (ObjectReference, error)
```

创建对象引用（指向整个对象）。

**返回值：**
- `ObjectReference` - 对象引用
- `error` - 错误信息

### 6.2 解引用

#### Dereference

```go
func (f *File) Dereference(ref ObjectReference) (interface{}, error)
func (g *group) Dereference(ref ObjectReference) (interface{}, error)
```

解引用，获取引用指向的对象。

**参数：**
- `ref` - 对象引用

**返回值：**
- `interface{}` - 对象（`Group` 或 `Dataset`）
- `error` - 错误信息

## 七、异步操作

### 7.1 异步结果

#### Wait

```go
func (r *AsyncResult) Wait() error
```

等待异步操作完成。

**返回值：**
- `error` - 错误信息

#### GetError

```go
func (r *AsyncResult) GetError() error
```

获取异步操作的错误。

**返回值：**
- `error` - 错误信息

### 7.2 异步管理器

#### Submit

```go
func (m *AsyncManager) Submit(f func() error) *AsyncResult
```

提交异步任务。

**参数：**
- `f` - 要异步执行的函数

**返回值：**
- `*AsyncResult` - 异步结果

**示例：**
```go
manager := hdf5.NewAsyncManager()
result := manager.Submit(func() error {
    return dset.Write(data)
})
// 等待完成
err := result.Wait()
```

## 八、元数据操作

### 8.1 获取元数据

#### GetMetadata

```go
func (f *File) GetMetadata() (FileMetadata, error)
```

获取文件元数据。

**返回值：**
- `FileMetadata` - 文件元数据
- `error` - 错误信息

#### GetObjectInfo

```go
func (f *File) GetObjectInfo(path string) (ObjectInfo, error)
```

获取对象信息。

**参数：**
- `path` - 对象路径

**返回值：**
- `ObjectInfo` - 对象信息
- `error` - 错误信息

### 8.2 访问对象

#### VisitObjects

```go
func (f *File) VisitObjects(path string, visitor func(info ObjectInfo) error) error
```

访问对象及其子对象。

**参数：**
- `path` - 起始路径
- `visitor` - 访问函数

**返回值：**
- `error` - 错误信息

## 九、扩展数据类型操作

### 9.1 添加字段

#### AddField

```go
func (ed *ExtensibleDatatype) AddField(name string, dtype Datatype, offset uint32) error
```

向扩展数据类型添加字段。

**参数：**
- `name` - 字段名称
- `dtype` - 字段数据类型
- `offset` - 字段偏移量

**返回值：**
- `error` - 错误信息

### 9.2 获取版本

#### GetVersion

```go
func (ed *ExtensibleDatatype) GetVersion(version uint8) (Datatype, error)
```

获取扩展数据类型的指定版本。

**参数：**
- `version` - 版本号

**返回值：**
- `Datatype` - 数据类型
- `error` - 错误信息

### 9.3 编码

#### Encode

```go
func (ed *ExtensibleDatatype) Encode() ([]byte, error)
```

编码扩展数据类型为字节数组。

**返回值：**
- `[]byte` - 编码后的字节数组
- `error` - 错误信息

## 十、类型定义

### 10.1 文件模式

```go
type FileMode uint8

const (
    ReadOnly        FileMode = iota  // 只读模式
    ReadWrite                        // 读写模式
    Create                           // 创建模式（覆盖现有文件）
    CreateExclusive                  // 独占创建模式
    Append                           // 追加模式
)
```

### 10.2 数据空间常量

```go
const H5S_UNLIMITED = ^uint64(0)  // 无限维度
```

### 10.3 选择类型

```go
type Selection struct {
    Points     [][]uint64  // 点选择
    Hyperslabs []Hyperslab // 超平面选择（支持多个）
    Mask       []bool      // 掩码选择
}

type Hyperslab struct {
    Start  []uint64 // 起始位置
    Count  []uint64 // 数量
    Stride []uint64 // 步长（可选）
}
```

### 10.4 属性列表

```go
type PropertyList struct {
    Chunks           []uint64      // 分块大小
    Compression      string        // 压缩算法（"gzip", "lzf", "zstd"）
    CompressionLevel int           // 压缩级别
    Shuffle          bool          // 是否启用 Shuffle 过滤器
    Fletcher32       bool          // 是否启用 Fletcher32 校验
    FillValue        interface{}   // 填充值
    DeflateLevel     int           // GZIP 压缩级别（同 CompressionLevel）
    BlockSize        uint64        // 块大小
    CacheSize        uint64        // 缓存大小
    MaxDims          []uint64      // 最大维度
    TrackTimes       bool          // 是否跟踪时间
    // ... 其他字段
}
```

### 10.5 复制选项

```go
type CopyOptions struct {
    CopyAttrs           bool // 是否复制属性
    CopyChildren        bool // 是否复制子对象
    HardLink            bool // 是否创建硬链接
    ExpandSoftLinks     bool // 是否展开软链接
    ExpandExternalLinks bool // 是否展开外部链接
    Recursive           bool // 是否递归复制
}
```

### 10.6 链接创建选项

```go
type LinkCreateOptions struct {
    CreateIntermediateGroups bool // 是否自动创建中间组
    AllowOverwrite           bool // 是否允许覆盖
    DoNotFollowExisting      bool // 是否不跟随已存在的链接
}
```

### 10.7 空间分配策略

```go
type SpaceAllocationStrategy uint8

const (
    SpaceStrategyDefault SpaceAllocationStrategy = 0
    SpaceStrategyFIFO    SpaceAllocationStrategy = 1
    SpaceStrategyBest    SpaceAllocationStrategy = 2
    SpaceStrategyWorst   SpaceAllocationStrategy = 3
)
```

### 10.8 数据类型类别

```go
type DatatypeClass uint8

const (
    DatatypeInteger   DatatypeClass = 0 // 整数
    DatatypeFloat     DatatypeClass = 1 // 浮点
    DatatypeString    DatatypeClass = 2 // 字符串
    DatatypeArray     DatatypeClass = 3 // 数组
    DatatypeCompound  DatatypeClass = 4 // 复合
    DatatypeEnum      DatatypeClass = 5 // 枚举
    DatatypeReference DatatypeClass = 6 // 引用
    DatatypeOpaque    DatatypeClass = 7 // 不透明
    DatatypeBitfield  DatatypeClass = 8 // 位域
    DatatypeVarLength DatatypeClass = 9 // 变长
)
```

> **注意**：上述常量值与 HDF5 规范和 `format` 包中定义的值保持一致，使用时不要修改。

### 10.9 预定义数据类型

```go
// 整数类型
var DatatypeInt8   = Datatype{Class: DatatypeInteger, Size: 1, ...}
var DatatypeInt16  = Datatype{Class: DatatypeInteger, Size: 2, ...}
var DatatypeInt32  = Datatype{Class: DatatypeInteger, Size: 4, ...}
var DatatypeInt64  = Datatype{Class: DatatypeInteger, Size: 8, ...}
var DatatypeUint8  = Datatype{Class: DatatypeInteger, Size: 1, Sign: Unsigned, ...}
var DatatypeUint16 = Datatype{Class: DatatypeInteger, Size: 2, Sign: Unsigned, ...}
var DatatypeUint32 = Datatype{Class: DatatypeInteger, Size: 4, Sign: Unsigned, ...}
var DatatypeUint64 = Datatype{Class: DatatypeInteger, Size: 8, Sign: Unsigned, ...}

// 浮点类型
var DatatypeFloat32 = Datatype{Class: DatatypeFloat, Size: 4, FloatBits: 24, ...}
var DatatypeFloat64 = Datatype{Class: DatatypeFloat, Size: 8, FloatBits: 53, ...}

// 复合类型（复数）
var DatatypeComplex64  = Datatype{Class: DatatypeCompound, Size: 8,  Fields: ...}
var DatatypeComplex128 = Datatype{Class: DatatypeCompound, Size: 16, Fields: ...}
```

### 10.10 引用类型

```go
type ObjectReference struct {
    ObjectAddr uint64
    ObjectType ObjectType
}

type RegionReference struct {
    DatasetAddr uint64
    Selection   Selection
}

type ObjectType uint8

const (
    ObjectTypeGroup    ObjectType = 1
    ObjectTypeDataset  ObjectType = 2
    ObjectTypeDatatype ObjectType = 3
)
```