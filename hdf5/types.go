package hdf5

import (
	"errors"
	"io"
	"os"
	"sync"

	"github.com/huangzhengshun/hdf5-go/format"
)

type FileMode uint8

const (
	ReadOnly FileMode = iota
	ReadWrite
	Create
	CreateExclusive
	Append
)

func (m FileMode) String() string {
	switch m {
	case ReadOnly:
		return "r"
	case ReadWrite:
		return "r+"
	case Create:
		return "w"
	case CreateExclusive:
		return "w-"
	case Append:
		return "a"
	default:
		return "unknown"
	}
}

type DatatypeClass uint8

const (
	DatatypeInteger   DatatypeClass = 0
	DatatypeFloat                   = 1
	DatatypeString                  = 2
	DatatypeCompound                = 4
	DatatypeArray                   = 5
	DatatypeEnum                    = 6
	DatatypeReference               = 7
	DatatypeOpaque                  = 8
	DatatypeBitfield                = 9
	DatatypeVarLength               = 10
)

var (
	DatatypeInt8      = Datatype{Class: DatatypeInteger, Size: 1, ByteOrder: LittleEndian, Sign: Signed}
	DatatypeInt16     = Datatype{Class: DatatypeInteger, Size: 2, ByteOrder: LittleEndian, Sign: Signed}
	DatatypeInt32     = Datatype{Class: DatatypeInteger, Size: 4, ByteOrder: LittleEndian, Sign: Signed}
	DatatypeInt64     = Datatype{Class: DatatypeInteger, Size: 8, ByteOrder: LittleEndian, Sign: Signed}
	DatatypeUint8     = Datatype{Class: DatatypeInteger, Size: 1, ByteOrder: LittleEndian, Sign: Unsigned}
	DatatypeUint16    = Datatype{Class: DatatypeInteger, Size: 2, ByteOrder: LittleEndian, Sign: Unsigned}
	DatatypeUint32    = Datatype{Class: DatatypeInteger, Size: 4, ByteOrder: LittleEndian, Sign: Unsigned}
	DatatypeUint64    = Datatype{Class: DatatypeInteger, Size: 8, ByteOrder: LittleEndian, Sign: Unsigned}
	DatatypeFloat32   = Datatype{Class: DatatypeFloat, Size: 4, ByteOrder: LittleEndian, FloatBits: 24}
	DatatypeFloat64   = Datatype{Class: DatatypeFloat, Size: 8, ByteOrder: LittleEndian, FloatBits: 53}
	DatatypeComplex64 = Datatype{Class: DatatypeCompound, Size: 8,
		Fields: []CompoundField{
			{Name: "r", Offset: 0, Datatype: DatatypeFloat32},
			{Name: "i", Offset: 4, Datatype: DatatypeFloat32},
		},
	}
	DatatypeComplex128 = Datatype{Class: DatatypeCompound, Size: 16,
		Fields: []CompoundField{
			{Name: "r", Offset: 0, Datatype: DatatypeFloat64},
			{Name: "i", Offset: 8, Datatype: DatatypeFloat64},
		},
	}
)

type CompoundField struct {
	Name     string
	Offset   uint32
	Datatype Datatype
}

type EnumMember struct {
	Name  string
	Value int64
}

type Datatype struct {
	Class         DatatypeClass
	Size          uint32
	ByteOrder     ByteOrder
	Sign          Sign
	FloatBits     uint8
	StringPadding StringPadding
	StringSize    uint32
	Compression   string
	Shuffle       bool
	Fletcher32    bool
	Fields        []CompoundField
	EnumMembers   []EnumMember
	ArrayDims     []uint64
	BaseType      *Datatype
}

type ByteOrder uint8

const (
	LittleEndian ByteOrder = iota
	BigEndian
	NativeEndian
)

type Sign uint8

const (
	Unsigned Sign = iota
	Signed
)

type StringPadding uint8

const (
	NullTerminated StringPadding = iota
	NullPad
	SpacePad
)

const H5S_UNLIMITED = ^uint64(0)

type LayoutClass uint8

const (
	LayoutCompact LayoutClass = iota
	LayoutContiguous
	LayoutChunked
)

type Dataspace struct {
	Dims    []uint64
	MaxDims []uint64
}

func (d Dataspace) Rank() int {
	return len(d.Dims)
}

func (d Dataspace) Size() uint64 {
	size := uint64(1)
	for _, dim := range d.Dims {
		size *= dim
	}
	return size
}

func (d Dataspace) Intersection(other Dataspace) (Dataspace, error) {
	if len(d.Dims) != len(other.Dims) {
		return Dataspace{}, errors.New("hdf5: dataspace rank mismatch")
	}

	result := Dataspace{
		Dims:    make([]uint64, len(d.Dims)),
		MaxDims: make([]uint64, len(d.Dims)),
	}

	for i := range d.Dims {
		result.Dims[i] = min(d.Dims[i], other.Dims[i])
		result.MaxDims[i] = min(d.MaxDims[i], other.MaxDims[i])
	}

	return result, nil
}

func (d Dataspace) Union(other Dataspace) (Dataspace, error) {
	if len(d.Dims) != len(other.Dims) {
		return Dataspace{}, errors.New("hdf5: dataspace rank mismatch")
	}

	result := Dataspace{
		Dims:    make([]uint64, len(d.Dims)),
		MaxDims: make([]uint64, len(d.Dims)),
	}

	for i := range d.Dims {
		result.Dims[i] = max(d.Dims[i], other.Dims[i])
		result.MaxDims[i] = max(d.MaxDims[i], other.MaxDims[i])
	}

	return result, nil
}

func (d Dataspace) Difference(other Dataspace) (Dataspace, error) {
	if len(d.Dims) != len(other.Dims) {
		return Dataspace{}, errors.New("hdf5: dataspace rank mismatch")
	}

	result := Dataspace{
		Dims:    make([]uint64, len(d.Dims)),
		MaxDims: make([]uint64, len(d.Dims)),
	}

	for i := range d.Dims {
		diff := int64(d.Dims[i]) - int64(other.Dims[i])
		if diff > 0 {
			result.Dims[i] = uint64(diff)
		}
		diff = int64(d.MaxDims[i]) - int64(other.MaxDims[i])
		if diff > 0 {
			result.MaxDims[i] = uint64(diff)
		}
	}

	return result, nil
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

type Filter struct {
	ID     uint8
	Params []uint32
}

type PropertyListClass uint8

const (
	PropertyListDefault       PropertyListClass = 0
	PropertyListFileCreate    PropertyListClass = 1
	PropertyListFileAccess    PropertyListClass = 2
	PropertyListDatasetCreate PropertyListClass = 3
	PropertyListDatasetAccess PropertyListClass = 4
	PropertyListGroupCreate   PropertyListClass = 5
)

type PropertyList struct {
	Class              PropertyListClass
	Chunks             []uint64
	Compression        string
	Shuffle            bool
	Fletcher32         bool
	CompressionLevel   int
	FillValue          interface{}
	DeflateLevel       int
	BlockSize          uint64
	CacheSize          uint64
	CacheSlabs         uint64
	Sizes              []uint64
	MaxDims            []uint64
	ExternalFile       string
	ExternalOffset     uint64
	ExternalSize       uint64
	ScaleType          int
	ScaleOffset        float64
	Bitshuffle         bool
	TrackTimes         bool
	SharedAccess       bool
	Alignment          uint64
	Threshold          uint64
	BlockSizeHint      uint64
	ChunkCacheSize     uint64
	ChunkCacheNSlots   int
	ChunkCacheW0       float64
	MetaBlockSize      uint64
	SmallBlockSize     uint64
	FreeSpacePersist   bool
	FreeSpaceThreshold float64
}

func NewPropertyList(class PropertyListClass) PropertyList {
	return PropertyList{Class: class}
}

func (p *PropertyList) SetChunk(chunks []uint64) {
	p.Chunks = chunks
}

func (p *PropertyList) SetCompression(compression string) {
	p.Compression = compression
}

func (p *PropertyList) SetCompressionLevel(level int) {
	p.CompressionLevel = level
}

func (p *PropertyList) SetShuffle() {
	p.Shuffle = true
}

func (p *PropertyList) SetFletcher32() {
	p.Fletcher32 = true
}

func (p *PropertyList) SetFillValue(value interface{}) {
	p.FillValue = value
}

func (p *PropertyList) SetDeflate(level int) {
	p.Compression = "gzip"
	p.DeflateLevel = level
	p.CompressionLevel = level
}

func (p *PropertyList) SetLZF() {
	p.Compression = "lzf"
}

func (p *PropertyList) SetZSTD(level int) {
	p.Compression = "zstd"
	p.CompressionLevel = level
}

func (p *PropertyList) SetScaleOffset(scaleType int, scaleOffset float64) {
	p.ScaleType = scaleType
	p.ScaleOffset = scaleOffset
}

func (p *PropertyList) SetBitshuffle() {
	p.Bitshuffle = true
}

func (p *PropertyList) SetTrackTimes(track bool) {
	p.TrackTimes = track
}

func (p *PropertyList) SetSharedAccess(shared bool) {
	p.SharedAccess = shared
}

func (p *PropertyList) SetAlignment(alignment uint64) {
	p.Alignment = alignment
}

func (p *PropertyList) SetThreshold(threshold uint64) {
	p.Threshold = threshold
}

func (p *PropertyList) SetChunkCache(size uint64, nslots int, w0 float64) {
	p.ChunkCacheSize = size
	p.ChunkCacheNSlots = nslots
	p.ChunkCacheW0 = w0
}

func (p *PropertyList) SetMetaBlockSize(size uint64) {
	p.MetaBlockSize = size
}

func (p *PropertyList) SetSmallBlockSize(size uint64) {
	p.SmallBlockSize = size
}

func (p *PropertyList) SetFreeSpace(persist bool, threshold float64) {
	p.FreeSpacePersist = persist
	p.FreeSpaceThreshold = threshold
}

func (p *PropertyList) Get(name string) interface{} {
	switch name {
	case "chunks":
		return p.Chunks
	case "compression":
		return p.Compression
	case "shuffle":
		return p.Shuffle
	case "fletcher32":
		return p.Fletcher32
	case "fillvalue":
		return p.FillValue
	case "track_times":
		return p.TrackTimes
	case "shared_access":
		return p.SharedAccess
	default:
		return nil
	}
}

func (p *PropertyList) Set(name string, value interface{}) {
	switch name {
	case "chunks":
		if v, ok := value.([]uint64); ok {
			p.Chunks = v
		}
	case "compression":
		if v, ok := value.(string); ok {
			p.Compression = v
		}
	case "shuffle":
		if v, ok := value.(bool); ok {
			p.Shuffle = v
		}
	case "fletcher32":
		if v, ok := value.(bool); ok {
			p.Fletcher32 = v
		}
	case "fillvalue":
		p.FillValue = value
	case "track_times":
		if v, ok := value.(bool); ok {
			p.TrackTimes = v
		}
	case "shared_access":
		if v, ok := value.(bool); ok {
			p.SharedAccess = v
		}
	}
}

type Group interface {
	Name() string
	Parent() Group
	File() *File
	CreateGroup(name string) (Group, error)
	GetGroup(name string) (Group, error)
	CreateDataset(name string, dtype Datatype, space Dataspace, plist PropertyList) (Dataset, error)
	GetDataset(name string) (Dataset, error)
	CreateVirtualDataset(name string, layout VirtualLayout, plist PropertyList) (Dataset, error)
	CreateDatatype(name string, dtype Datatype) (Datatype, error)
	GetDatatype(name string) (Datatype, error)
	CreateLink(targetPath, linkName string) error
	CreateLinkWithOptions(targetPath, linkName string, options LinkCreateOptions) error
	CreateExternalLink(externalFile, targetPath, linkName string) error
	CreateExternalLinkWithOptions(externalFile, targetPath, linkName string, options LinkCreateOptions) error
	Keys() []string
	Visit(func(string, interface{}) error) error
	VisitItems(func(name string, obj interface{}) error) error
	Attributes() *AttributeManager
	Copy(dest Group, name string) error
	CopyWithOptions(dest Group, name string, options CopyOptions) error
	Move(dest Group, name string) error
	Delete(name string) error
	CreateReference() (ObjectReference, error)
	Dereference(ref ObjectReference) (interface{}, error)
	GetGroupByAddr(addr uint64) (Group, error)
	GetDatasetByAddr(addr uint64) (Dataset, error)
	Close() error
}

type Hyperslab struct {
	Start  []uint64
	Count  []uint64
	Stride []uint64
}

type Selection struct {
	Hyperslabs []Hyperslab
	Points     [][]uint64
	Mask       []bool
}

func NewSelection() Selection {
	return Selection{}
}

func NewHyperslab(start, count, stride []uint64) Hyperslab {
	return Hyperslab{
		Start:  start,
		Count:  count,
		Stride: stride,
	}
}

func (s *Selection) AddHyperslab(start, count, stride []uint64) {
	s.Hyperslabs = append(s.Hyperslabs, Hyperslab{
		Start:  start,
		Count:  count,
		Stride: stride,
	})
}

func (s *Selection) AddPoints(points [][]uint64) {
	s.Points = append(s.Points, points...)
}

type RegionReference struct {
	DatasetAddr uint64
	Selection   Selection
}

type ObjectReference struct {
	ObjectAddr uint64
	ObjectType ObjectType
}

type ReferenceType uint8

const (
	ReferenceDataset ReferenceType = iota
	ReferenceRegion
)

type ObjectType uint8

const (
	ObjectTypeGroup    ObjectType = 1
	ObjectTypeDataset  ObjectType = 2
	ObjectTypeDatatype ObjectType = 3
)

type VirtualSource struct {
	FilePath    string
	DatasetPath string
	Selection   Selection
	TargetStart []uint64
	TargetCount []uint64
	Scale       []uint64
}

type VirtualLayout struct {
	Dataspace Dataspace
	Datatype  Datatype
	Sources   []VirtualSource
}

type CopyOptions struct {
	CopyAttrs           bool
	CopyChildren        bool
	HardLink            bool
	ExpandSoftLinks     bool
	ExpandExternalLinks bool
	Recursive           bool
}

type LinkCreateOptions struct {
	CreateIntermediateGroups bool
	AllowOverwrite           bool
	DoNotFollowExisting      bool
}

type Dataset interface {
	Name() string
	Parent() Group
	File() *File
	Datatype() Datatype
	Dataspace() Dataspace
	Read(data interface{}) error
	Write(data interface{}) error
	ReadSlice(data interface{}, sel Selection) error
	WriteSlice(data interface{}, sel Selection) error
	Resize(newDims []uint64) error
	Flush() error
	CreateReference(sel Selection) (RegionReference, error)
	CreateObjectReference() (ObjectReference, error)
	Attributes() *AttributeManager
	Close() error
}

type Attribute interface {
	Name() string
	Datatype() Datatype
	Dataspace() Dataspace
	Read(data interface{}) error
	Write(data interface{}) error
}

type AttributeManager struct {
	obj interface{}
}

type File struct {
	name   string
	mode   FileMode
	reader io.ReadSeeker
	writer io.WriteSeeker
	root   Group
	super  *format.Superblock
	closed bool
	mu     sync.RWMutex
	locked bool
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Mode() FileMode {
	return f.mode
}

func (f *File) Root() Group {
	return f.root
}

func (f *File) CreateGroup(name string) (Group, error) {
	return f.root.CreateGroup(name)
}

func (f *File) GetGroup(name string) (Group, error) {
	return f.root.GetGroup(name)
}

func (f *File) CreateDataset(name string, dtype Datatype, space Dataspace, plist PropertyList) (Dataset, error) {
	return f.root.CreateDataset(name, dtype, space, plist)
}

func (f *File) GetDataset(name string) (Dataset, error) {
	return f.root.GetDataset(name)
}

func (f *File) CreateLink(targetPath, linkName string) error {
	return f.root.CreateLink(targetPath, linkName)
}

func (f *File) CreateExternalLink(externalFile, targetPath, linkName string) error {
	return f.root.CreateExternalLink(externalFile, targetPath, linkName)
}

func (f *File) Attributes() *AttributeManager {
	return &AttributeManager{obj: f}
}

func (f *File) Dereference(ref ObjectReference) (interface{}, error) {
	return f.root.Dereference(ref)
}

func (f *File) Keys() []string {
	return f.root.Keys()
}

func (f *File) Lock() error {
	if f.closed {
		return errors.New("hdf5: file is closed")
	}

	file, ok := f.reader.(*os.File)
	if !ok {
		return errors.New("hdf5: file locking not supported for memory files")
	}

	return lockFile(file, true)
}

func (f *File) Unlock() error {
	if f.closed {
		return errors.New("hdf5: file is closed")
	}

	file, ok := f.reader.(*os.File)
	if !ok {
		return errors.New("hdf5: file locking not supported for memory files")
	}

	return unlockFile(file)
}

func (f *File) TryLock() bool {
	if f.closed {
		return false
	}

	file, ok := f.reader.(*os.File)
	if !ok {
		return false
	}

	err := lockFile(file, false)
	return err == nil
}

func (f *File) Visit(fn func(string, interface{}) error) error {
	return f.root.Visit(fn)
}

func (f *File) VisitItems(fn func(name string, obj interface{}) error) error {
	return f.root.VisitItems(fn)
}

func (f *File) Copy(dest Group, name string) error {
	return f.root.Copy(dest, name)
}

func (f *File) Move(dest Group, name string) error {
	return f.root.Move(dest, name)
}

func (f *File) Delete(name string) error {
	return f.root.Delete(name)
}
