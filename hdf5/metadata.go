package hdf5

import (
	"os"
	"time"

	"github.com/zhengshun.huang/hdf5-go/format"
)

type FileMetadata struct {
	Filename             string
	Filesize             int64
	SuperblockVersion    uint8
	OffsetSize           uint8
	LengthSize           uint8
	GroupLeafNodeK       uint16
	GroupInternalNodeK   uint16
	RootGroupAddr        uint64
	FreeSpaceManagerAddr uint64
	FileConsistencyFlags uint32
	CreationTime         time.Time
	ModificationTime     time.Time
}

func (f *File) GetMetadata() (FileMetadata, error) {
	if f.closed {
		return FileMetadata{}, ErrClosedFile
	}

	info, err := f.fileInfo()
	if err != nil {
		return FileMetadata{}, err
	}

	return FileMetadata{
		Filename:             f.name,
		Filesize:             info.size,
		SuperblockVersion:    f.super.Version,
		OffsetSize:           f.super.OffsetSize,
		LengthSize:           f.super.LengthSize,
		GroupLeafNodeK:       f.super.GroupLeafNodeK,
		GroupInternalNodeK:   f.super.GroupInternalNodeK,
		RootGroupAddr:        f.super.RootGroupAddr,
		FreeSpaceManagerAddr: f.super.FreeSpaceManagerAddr,
		FileConsistencyFlags: f.super.FileConsistencyFlags,
		CreationTime:         info.creationTime,
		ModificationTime:     info.modTime,
	}, nil
}

type fileInfo struct {
	size         int64
	creationTime time.Time
	modTime      time.Time
}

func (f *File) fileInfo() (fileInfo, error) {
	if mbuf, ok := f.reader.(*mmapBuffer); ok {
		return fileInfo{
			size: int64(len(mbuf.data)),
		}, nil
	}

	if buf, ok := f.reader.(*memoryBuffer); ok {
		return fileInfo{
			size: int64(len(buf.data)),
		}, nil
	}

	if file, ok := f.reader.(*os.File); ok {
		stat, err := file.Stat()
		if err != nil {
			return fileInfo{}, err
		}
		return fileInfo{
			size:         stat.Size(),
			creationTime: stat.ModTime(),
			modTime:      stat.ModTime(),
		}, nil
	}

	return fileInfo{}, nil
}

type DatasetInfo struct {
	Name       string
	Datatype   Datatype
	Dataspace  Dataspace
	Layout     format.LayoutMessage
	Attributes int
	ChunkCache *ChunkCache
}

func (d *dataset) GetInfo() DatasetInfo {
	return DatasetInfo{
		Name:       d.name,
		Datatype:   d.dtype,
		Dataspace:  d.dspace,
		Layout:     d.layout,
		Attributes: countAttributes(d.header),
		ChunkCache: d.chunkCache,
	}
}

type GroupInfo struct {
	Name       string
	Attributes int
	NumLinks   int
}

func (g *group) GetInfo() GroupInfo {
	return GroupInfo{
		Name:       g.name,
		Attributes: countAttributes(g.header),
		NumLinks:   len(g.links),
	}
}

type ObjectInfo struct {
	Name         string
	Path         string
	Address      uint64
	Type         ObjectType
	HeaderSize   uint64
	NumMessages  uint16
	Attributes   int
	Links        int
	Datatype     Datatype
	Dataspace    Dataspace
	Layout       format.LayoutMessage
	ChunkSize    []uint64
	Compression  string
	FilterCount  int
	IsChunked    bool
	IsCompact    bool
	IsContiguous bool
	IsVirtual    bool
}

func (f *File) GetObjectInfo(path string) (ObjectInfo, error) {
	if f.closed {
		return ObjectInfo{}, ErrClosedFile
	}

	objInfo := ObjectInfo{
		Path: path,
	}

	if path == "/" {
		grp := f.root.(*group)
		return ObjectInfo{
			Name:        "/",
			Path:        "/",
			Address:     grp.addr,
			Type:        ObjectTypeGroup,
			HeaderSize:  grp.header.TotalHeaderSize,
			NumMessages: grp.header.NumberOfMessages,
			Attributes:  countAttributes(grp.header),
			Links:       len(grp.links),
		}, nil
	}

	parentPath, childName := parsePath(path)
	var parent Group = f.root

	if parentPath != "/" && parentPath != "" {
		var err error
		parent, err = f.root.GetGroup(parentPath)
		if err != nil {
			return ObjectInfo{}, err
		}
	}

	grp := parent.(*group)
	if linkInfo, ok := grp.links[childName]; ok {
		objInfo.Name = childName
		objInfo.Address = linkInfo.TargetAddress

		if linkInfo.LinkType == format.LinkTypeSoft || linkInfo.LinkType == format.LinkTypeExternal {
			return ObjectInfo{}, ErrExternalLink
		}

		oh, err := format.ReadObjectHeader(f.reader, f.super, linkInfo.TargetAddress)
		if err != nil {
			return ObjectInfo{}, err
		}

		objInfo.HeaderSize = oh.TotalHeaderSize
		objInfo.NumMessages = oh.NumberOfMessages
		objInfo.Attributes = countAttributes(oh)

		if isGroup(oh) {
			objInfo.Type = ObjectTypeGroup
			subGrp, err := grp.GetGroup(childName)
			if err == nil {
				objInfo.Links = len(subGrp.(*group).links)
			}
		} else {
			objInfo.Type = ObjectTypeDataset
			dset, err := grp.GetDataset(childName)
			if err == nil {
				ds := dset.(*dataset)
				objInfo.Datatype = ds.dtype
				objInfo.Dataspace = ds.dspace
				objInfo.Layout = ds.layout
				objInfo.Compression = ds.dtype.Compression
				objInfo.IsChunked = ds.layout.LayoutClass == format.LayoutChunked
				objInfo.IsCompact = ds.layout.LayoutClass == format.LayoutCompact
				objInfo.IsContiguous = ds.layout.LayoutClass == format.LayoutContiguous

				if objInfo.IsChunked {
					objInfo.ChunkSize = ds.layout.ChunkDims
				}

				filterCount := 0
				for _, msg := range oh.Messages {
					if msg.Type == format.MessageFilterPipeline {
						filterCount++
					}
				}
				objInfo.FilterCount = filterCount
			}
		}

		return objInfo, nil
	}

	return ObjectInfo{}, ErrNotFound
}

func countAttributes(header *format.ObjectHeader) int {
	count := 0
	for _, msg := range header.Messages {
		if msg.Type == format.MessageAttribute {
			count++
		}
	}
	return count
}

func (f *File) VisitObjects(path string, visitor func(info ObjectInfo) error) error {
	if f.closed {
		return ErrClosedFile
	}

	return f.visitObjectsRecursive(path, visitor)
}

func (f *File) visitObjectsRecursive(path string, visitor func(info ObjectInfo) error) error {
	info, err := f.GetObjectInfo(path)
	if err != nil {
		return err
	}

	if err := visitor(info); err != nil {
		return err
	}

	if info.Type == ObjectTypeGroup {
		grp, err := f.root.GetGroup(path)
		if err != nil {
			return err
		}

		for _, name := range grp.Keys() {
			var childPath string
			if path == "/" {
				childPath = "/" + name
			} else {
				childPath = path + "/" + name
			}

			if err := f.visitObjectsRecursive(childPath, visitor); err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *File) GetLibraryVersion() string {
	return "HDF5-Go v1.0.0"
}

func (f *File) GetFileVersion() uint8 {
	if f.closed {
		return 0
	}
	return f.super.Version
}
