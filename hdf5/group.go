package hdf5

import (
	"bytes"
	"errors"
	"strings"

	"github.com/huangzhengshun/hdf5-go/format"
)

type LinkInfo struct {
	LinkType      uint8
	TargetAddress uint64
	TargetPath    string
	ExternalFile  string
}

type group struct {
	name   string
	addr   uint64
	file   *File
	header *format.ObjectHeader
	super  *format.Superblock
	links  map[string]LinkInfo
	closed bool
}

func (g *group) Name() string {
	return g.name
}

func (g *group) Parent() Group {
	if g.name == "/" {
		return nil
	}
	idx := bytes.LastIndexByte([]byte(g.name), '/')
	if idx <= 0 {
		return g.file.root
	}
	parentName := g.name[:idx]
	if parentName == "" {
		return g.file.root
	}
	parent, _ := g.file.root.GetGroup(parentName)
	return parent
}

func (g *group) File() *File {
	return g.file
}

func (g *group) CreateGroup(name string) (Group, error) {
	if g.closed {
		return nil, ErrClosedGroup
	}

	fullPath := name
	if g.name == "/" {
		fullPath = "/" + name
	} else {
		fullPath = g.name + "/" + name
	}

	newOH := &format.ObjectHeader{
		Version:          format.ObjectHeaderVersion2,
		Flags:            0,
		TotalHeaderSize:  1024,
		NumberOfMessages: 0,
		Messages:         []format.ObjectHeaderMessage{},
	}

	addr, err := format.WriteObjectHeader(g.file.writer, g.super, newOH)
	if err != nil {
		return nil, err
	}

	if g.links == nil {
		g.links = make(map[string]LinkInfo)
	}
	g.links[name] = LinkInfo{LinkType: format.LinkTypeHard, TargetAddress: addr}

	g.addLinkMessage(name, addr)

	newGroup := &group{
		name:   fullPath,
		addr:   addr,
		file:   g.file,
		header: newOH,
		super:  g.super,
		links:  make(map[string]LinkInfo),
	}

	return newGroup, nil
}

func (g *group) addLinkMessage(name string, addr uint64) error {
	linkMsg := format.LinkMessage{
		Version:       1,
		LinkType:      format.LinkTypeHard,
		CreationOrder: 0,
		Flags:         0,
		Name:          name,
		TargetAddress: addr,
	}

	linkMsgData, err := format.EncodeLinkMessage(linkMsg)
	if err != nil {
		return err
	}

	newLinkMsg := format.ObjectHeaderMessage{
		Type:    format.MessageLinkMessage,
		Size:    uint16(10 + len(linkMsgData)),
		Flags:   0,
		Content: linkMsgData,
	}

	g.header.Messages = append(g.header.Messages, newLinkMsg)
	g.header.NumberOfMessages++

	actualSize := uint64(4) + uint64(1) + uint64(1) + uint64(g.super.LengthSize) + uint64(2) + uint64(g.super.OffsetSize)*uint64(g.header.NumberOfMessages)
	for _, msg := range g.header.Messages {
		actualSize += uint64(msg.Size)
	}

	if g.header.TotalHeaderSize == 0 {
		g.header.TotalHeaderSize = ((actualSize + 1023) / 1024) * 1024
	} else if g.header.TotalHeaderSize < actualSize {
		g.header.TotalHeaderSize = ((actualSize + 1023) / 1024) * 1024
	}

	if _, err := g.file.writer.Seek(int64(g.addr), 0); err != nil {
		return err
	}
	_, err = format.WriteObjectHeader(g.file.writer, g.super, g.header)
	if err != nil {
		return err
	}

	if _, err := g.file.writer.Seek(0, 2); err != nil {
		return err
	}
	return nil
}

func (g *group) CreateLink(targetPath, linkName string) error {
	return g.CreateLinkWithOptions(targetPath, linkName, LinkCreateOptions{})
}

func (g *group) CreateLinkWithOptions(targetPath, linkName string, options LinkCreateOptions) error {
	if g.closed {
		return ErrClosedGroup
	}

	var parentGroup *group = g
	var finalLinkName string = linkName

	if options.CreateIntermediateGroups {
		parts := strings.Split(linkName, "/")
		if len(parts) > 1 {
			var currentPath string
			for i := 0; i < len(parts)-1; i++ {
				if currentPath == "" {
					currentPath = parts[i]
				} else {
					currentPath = currentPath + "/" + parts[i]
				}
				_, err := g.GetGroup(currentPath)
				if err != nil {
					_, err = g.CreateGroup(currentPath)
					if err != nil {
						return err
					}
				}
			}
			parentGrp, _ := g.GetGroup(strings.Join(parts[:len(parts)-1], "/"))
			parentGroup = parentGrp.(*group)
			finalLinkName = parts[len(parts)-1]
		}
	}

	if !options.AllowOverwrite {
		if _, ok := parentGroup.links[finalLinkName]; ok {
			return errors.New("hdf5: link already exists")
		}
	}

	if parentGroup.links == nil {
		parentGroup.links = make(map[string]LinkInfo)
	}
	parentGroup.links[finalLinkName] = LinkInfo{LinkType: format.LinkTypeSoft, TargetPath: targetPath}

	linkMsg := format.LinkMessage{
		Version:       1,
		LinkType:      format.LinkTypeSoft,
		CreationOrder: 0,
		Flags:         0,
		Name:          finalLinkName,
		TargetPath:    targetPath,
	}

	linkMsgData, err := format.EncodeLinkMessage(linkMsg)
	if err != nil {
		return err
	}

	newLinkMsg := format.ObjectHeaderMessage{
		Type:    format.MessageLinkMessage,
		Size:    uint16(10 + len(linkMsgData)),
		Flags:   0,
		Content: linkMsgData,
	}

	parentGroup.header.Messages = append(parentGroup.header.Messages, newLinkMsg)
	parentGroup.header.NumberOfMessages++

	actualSize := uint64(4) + uint64(1) + uint64(1) + uint64(parentGroup.super.LengthSize) + uint64(2) + uint64(parentGroup.super.OffsetSize)*uint64(parentGroup.header.NumberOfMessages)
	for _, msg := range parentGroup.header.Messages {
		actualSize += uint64(msg.Size)
	}

	if parentGroup.header.TotalHeaderSize == 0 {
		parentGroup.header.TotalHeaderSize = ((actualSize + 1023) / 1024) * 1024
	} else if parentGroup.header.TotalHeaderSize < actualSize {
		parentGroup.header.TotalHeaderSize = ((actualSize + 1023) / 1024) * 1024
	}

	if _, err := parentGroup.file.writer.Seek(int64(parentGroup.addr), 0); err != nil {
		return err
	}
	_, err = format.WriteObjectHeader(parentGroup.file.writer, parentGroup.super, parentGroup.header)
	if err != nil {
		return err
	}

	endPos := int64(parentGroup.addr) + int64(parentGroup.header.TotalHeaderSize)
	if _, err := parentGroup.file.writer.Seek(endPos, 0); err != nil {
		return err
	}
	return nil
}

func (g *group) CreateExternalLink(externalFile, targetPath, linkName string) error {
	return g.CreateExternalLinkWithOptions(externalFile, targetPath, linkName, LinkCreateOptions{})
}

func (g *group) CreateExternalLinkWithOptions(externalFile, targetPath, linkName string, options LinkCreateOptions) error {
	if g.closed {
		return ErrClosedGroup
	}

	var parentGroup *group = g
	var finalLinkName string = linkName

	if options.CreateIntermediateGroups {
		parts := strings.Split(linkName, "/")
		if len(parts) > 1 {
			var currentPath string
			for i := 0; i < len(parts)-1; i++ {
				if currentPath == "" {
					currentPath = parts[i]
				} else {
					currentPath = currentPath + "/" + parts[i]
				}
				_, err := g.GetGroup(currentPath)
				if err != nil {
					_, err = g.CreateGroup(currentPath)
					if err != nil {
						return err
					}
				}
			}
			parentGrp, _ := g.GetGroup(strings.Join(parts[:len(parts)-1], "/"))
			parentGroup = parentGrp.(*group)
			finalLinkName = parts[len(parts)-1]
		}
	}

	if !options.AllowOverwrite {
		if _, ok := parentGroup.links[finalLinkName]; ok {
			return errors.New("hdf5: link already exists")
		}
	}

	if parentGroup.links == nil {
		parentGroup.links = make(map[string]LinkInfo)
	}
	parentGroup.links[finalLinkName] = LinkInfo{LinkType: format.LinkTypeExternal, TargetPath: targetPath, ExternalFile: externalFile}

	linkMsg := format.LinkMessage{
		Version:       1,
		LinkType:      format.LinkTypeExternal,
		CreationOrder: 0,
		Flags:         0,
		Name:          finalLinkName,
		ExternalFile:  externalFile,
		TargetPath:    targetPath,
	}

	linkMsgData, err := format.EncodeLinkMessage(linkMsg)
	if err != nil {
		return err
	}

	newLinkMsg := format.ObjectHeaderMessage{
		Type:    format.MessageLinkMessage,
		Size:    uint16(10 + len(linkMsgData)),
		Flags:   0,
		Content: linkMsgData,
	}

	parentGroup.header.Messages = append(parentGroup.header.Messages, newLinkMsg)
	parentGroup.header.NumberOfMessages++

	actualSize := uint64(4) + uint64(1) + uint64(1) + uint64(parentGroup.super.LengthSize) + uint64(2) + uint64(parentGroup.super.OffsetSize)*uint64(parentGroup.header.NumberOfMessages)
	for _, msg := range parentGroup.header.Messages {
		actualSize += uint64(msg.Size)
	}

	if parentGroup.header.TotalHeaderSize == 0 {
		parentGroup.header.TotalHeaderSize = ((actualSize + 1023) / 1024) * 1024
	} else if parentGroup.header.TotalHeaderSize < actualSize {
		parentGroup.header.TotalHeaderSize = ((actualSize + 1023) / 1024) * 1024
	}

	if _, err := parentGroup.file.writer.Seek(int64(parentGroup.addr), 0); err != nil {
		return err
	}
	_, err = format.WriteObjectHeader(parentGroup.file.writer, parentGroup.super, parentGroup.header)
	if err != nil {
		return err
	}

	endPos := int64(parentGroup.addr) + int64(parentGroup.header.TotalHeaderSize)
	if _, err := parentGroup.file.writer.Seek(endPos, 0); err != nil {
		return err
	}
	return nil
}

func (g *group) GetGroup(name string) (Group, error) {
	if g.closed {
		return nil, ErrClosedGroup
	}

	var parentPath, childName string
	if strings.HasPrefix(name, "/") {
		parentPath, childName = parsePath(name)
	} else {
		parentPath = g.name
		childName = name
	}

	if parentPath != g.name {
		parent, err := g.file.root.GetGroup(parentPath)
		if err != nil {
			return nil, err
		}
		return parent.GetGroup(childName)
	}

	if g.links != nil {
		if linkInfo, ok := g.links[childName]; ok {
			var fullPath string
			if g.name == "/" {
				fullPath = "/" + childName
			} else {
				fullPath = g.name + "/" + childName
			}

			switch linkInfo.LinkType {
			case format.LinkTypeSoft:
				var targetPath string
				if strings.HasPrefix(linkInfo.TargetPath, "/") {
					targetPath = linkInfo.TargetPath
				} else {
					if g.name == "/" {
						targetPath = "/" + linkInfo.TargetPath
					} else {
						targetPath = g.name + "/" + linkInfo.TargetPath
					}
				}
				return g.file.root.GetGroup(targetPath)
			case format.LinkTypeExternal:
				extFile, err := OpenFile(linkInfo.ExternalFile, ReadOnly)
				if err != nil {
					return nil, err
				}
				return extFile.root.GetGroup(linkInfo.TargetPath)
			}

			oh, err := format.ReadObjectHeader(g.file.reader, g.super, linkInfo.TargetAddress)
			if err != nil {
				return nil, err
			}

			return &group{
				name:   fullPath,
				addr:   linkInfo.TargetAddress,
				file:   g.file,
				header: oh,
				super:  g.super,
				links:  parseLinksFromObjectHeader(oh),
			}, nil
		}
	}

	return nil, ErrNotFound
}

func (g *group) CreateDataset(name string, dtype Datatype, space Dataspace, plist PropertyList) (Dataset, error) {
	if g.closed {
		return nil, ErrClosedGroup
	}

	fullPath := name
	if g.name == "/" {
		fullPath = "/" + name
	} else {
		fullPath = g.name + "/" + name
	}

	dset, err := createDataset(g.file, fullPath, dtype, space, plist)
	if err != nil {
		return nil, err
	}

	if g.links == nil {
		g.links = make(map[string]LinkInfo)
	}
	g.links[name] = LinkInfo{LinkType: format.LinkTypeHard, TargetAddress: dset.(*dataset).addr}

	if err := g.addLinkMessage(name, dset.(*dataset).addr); err != nil {
		return nil, err
	}

	return dset, nil
}

func (g *group) CreateVirtualDataset(name string, layout VirtualLayout, plist PropertyList) (Dataset, error) {
	if g.closed {
		return nil, ErrClosedGroup
	}

	fullPath := name
	if g.name == "/" {
		fullPath = "/" + name
	} else {
		fullPath = g.name + "/" + name
	}

	dset := &virtualDataset{
		name:   fullPath,
		group:  g,
		layout: layout,
		closed: false,
	}

	if g.links == nil {
		g.links = make(map[string]LinkInfo)
	}
	g.links[name] = LinkInfo{LinkType: format.LinkTypeHard, TargetAddress: 0}

	if err := g.addLinkMessage(name, 0); err != nil {
		return nil, err
	}

	return dset, nil
}

func (g *group) GetDataset(name string) (Dataset, error) {
	if g.closed {
		return nil, ErrClosedGroup
	}

	var parentPath, childName string
	if strings.HasPrefix(name, "/") {
		parentPath, childName = parsePath(name)
	} else {
		parentPath = g.name
		childName = name
	}

	if parentPath != g.name {
		parent, err := g.file.root.GetGroup(parentPath)
		if err != nil {
			return nil, err
		}
		return parent.GetDataset(childName)
	}

	if g.links != nil {
		if linkInfo, ok := g.links[childName]; ok {
			var fullPath string
			if g.name == "/" {
				fullPath = "/" + childName
			} else {
				fullPath = g.name + "/" + childName
			}

			switch linkInfo.LinkType {
			case format.LinkTypeSoft:
				return g.file.root.GetDataset(linkInfo.TargetPath)
			case format.LinkTypeExternal:
				extFile, err := OpenFile(linkInfo.ExternalFile, ReadOnly)
				if err != nil {
					return nil, err
				}
				return extFile.root.GetDataset(linkInfo.TargetPath)
			}

			oh, err := format.ReadObjectHeader(g.file.reader, g.super, linkInfo.TargetAddress)
			if err != nil {
				return nil, err
			}

			dset := &dataset{
				name:   fullPath,
				addr:   linkInfo.TargetAddress,
				file:   g.file,
				super:  g.super,
				header: oh,
				closed: false,
			}

			for _, msg := range oh.Messages {
				if msg.Type == format.MessageDatatype {
					dtMsg, _ := format.DecodeDatatypeMessage(msg.Content)
					dset.dtype = Datatype{
						Class:     DatatypeClass(dtMsg.Class),
						Size:      dtMsg.Size,
						ByteOrder: ByteOrder(dtMsg.ByteOrder),
						Sign:      Sign(dtMsg.Sign),
						FloatBits: dtMsg.FloatBits,
					}
					if dtMsg.Class == format.ClassCompound {
						for _, field := range dtMsg.Fields {
							dset.dtype.Fields = append(dset.dtype.Fields, CompoundField{
								Name:   field.Name,
								Offset: field.Offset,
								Datatype: Datatype{
									Class:     DatatypeClass(field.Datatype.Class),
									Size:      field.Datatype.Size,
									ByteOrder: ByteOrder(field.Datatype.ByteOrder),
									Sign:      Sign(field.Datatype.Sign),
									FloatBits: field.Datatype.FloatBits,
								},
							})
						}
					} else if dtMsg.Class == format.ClassEnum {
						if dtMsg.EnumBase != nil {
							baseType := Datatype{
								Class:     DatatypeClass(dtMsg.EnumBase.Class),
								Size:      dtMsg.EnumBase.Size,
								ByteOrder: ByteOrder(dtMsg.EnumBase.ByteOrder),
								Sign:      Sign(dtMsg.EnumBase.Sign),
							}
							dset.dtype.BaseType = &baseType
						}
						for _, member := range dtMsg.EnumMembers {
							dset.dtype.EnumMembers = append(dset.dtype.EnumMembers, EnumMember{
								Name:  member.Name,
								Value: member.Value,
							})
						}
					} else if dtMsg.Class == format.ClassArray || dtMsg.Class == format.ClassVarLength {
						if dtMsg.ArrayBase != nil {
							baseType := Datatype{
								Class:     DatatypeClass(dtMsg.ArrayBase.Class),
								Size:      dtMsg.ArrayBase.Size,
								ByteOrder: ByteOrder(dtMsg.ArrayBase.ByteOrder),
								Sign:      Sign(dtMsg.ArrayBase.Sign),
								FloatBits: dtMsg.ArrayBase.FloatBits,
							}
							dset.dtype.BaseType = &baseType
						}
						dset.dtype.ArrayDims = dtMsg.ArrayDims
					}
				} else if msg.Type == format.MessageDataspace {
					dsMsg, _ := format.DecodeDataspaceMessage(msg.Content)
					dset.dspace = Dataspace{
						Dims:    dsMsg.Dims,
						MaxDims: dsMsg.MaxDims,
					}
				} else if msg.Type == format.MessageDataset {
					layoutMsg, _ := format.DecodeLayoutMessage(msg.Content)
					dset.layout = layoutMsg
					if layoutMsg.LayoutClass == format.LayoutChunked {
						chunkIndex, _ := format.ReadChunkIndex(g.file.reader, layoutMsg.DataAddress)
						dset.chunkIndex = chunkIndex
					}
				} else if msg.Type == format.MessageFilterPipeline {
					fpm, _ := format.DecodeFilterPipelineMessage(msg.Content)
					for _, f := range fpm.Filters {
						if f.FilterID == format.FilterDeflate {
							dset.dtype.Compression = "gzip"
						} else if f.FilterID == format.FilterLZF {
							dset.dtype.Compression = "lzf"
						} else if f.FilterID == format.FilterZSTD {
							dset.dtype.Compression = "zstd"
						} else if f.FilterID == format.FilterShuffle {
							dset.dtype.Shuffle = true
						} else if f.FilterID == format.FilterFletcher32 {
							dset.dtype.Fletcher32 = true
						}
					}
				}
			}

			return dset, nil
		}
	}

	return nil, ErrNotFound
}

func (g *group) CreateDatatype(name string, dtype Datatype) (Datatype, error) {
	if g.closed {
		return Datatype{}, ErrClosedGroup
	}

	dtypeMsg, err := format.EncodeDatatypeMessage(format.DatatypeMessage{
		Class:         uint8(dtype.Class),
		Size:          dtype.Size,
		ByteOrder:     uint8(dtype.ByteOrder),
		Sign:          uint8(dtype.Sign),
		FloatBits:     dtype.FloatBits,
		StringPadding: uint8(dtype.StringPadding),
		StringSize:    dtype.StringSize,
		ArrayDims:     dtype.ArrayDims,
	})
	if err != nil {
		return Datatype{}, err
	}

	oh := &format.ObjectHeader{
		Version:          format.ObjectHeaderVersion2,
		Flags:            0,
		TotalHeaderSize:  1024,
		NumberOfMessages: 1,
		Messages: []format.ObjectHeaderMessage{
			{
				Type:    format.MessageDatatype,
				Size:    uint16(10 + len(dtypeMsg)),
				Content: dtypeMsg,
			},
		},
	}

	addr, err := format.WriteObjectHeader(g.file.writer, g.super, oh)
	if err != nil {
		return Datatype{}, err
	}

	if g.links == nil {
		g.links = make(map[string]LinkInfo)
	}
	g.links[name] = LinkInfo{LinkType: format.LinkTypeHard, TargetAddress: addr}

	if err := g.addLinkMessage(name, addr); err != nil {
		return Datatype{}, err
	}

	return dtype, nil
}

func (g *group) GetDatatype(name string) (Datatype, error) {
	if g.closed {
		return Datatype{}, ErrClosedGroup
	}

	var parentPath, childName string
	if strings.HasPrefix(name, "/") {
		parentPath, childName = parsePath(name)
	} else {
		parentPath = g.name
		childName = name
	}

	if parentPath != g.name {
		parent, err := g.file.root.GetGroup(parentPath)
		if err != nil {
			return Datatype{}, err
		}
		return parent.GetDatatype(childName)
	}

	if g.links != nil {
		if linkInfo, ok := g.links[childName]; ok {
			oh, err := format.ReadObjectHeader(g.file.reader, g.super, linkInfo.TargetAddress)
			if err != nil {
				return Datatype{}, err
			}

			for _, msg := range oh.Messages {
				if msg.Type == format.MessageDatatype {
					dtMsg, _ := format.DecodeDatatypeMessage(msg.Content)
					dtype := Datatype{
						Class:         DatatypeClass(dtMsg.Class),
						Size:          dtMsg.Size,
						ByteOrder:     ByteOrder(dtMsg.ByteOrder),
						Sign:          Sign(dtMsg.Sign),
						FloatBits:     dtMsg.FloatBits,
						StringPadding: StringPadding(dtMsg.StringPadding),
						StringSize:    dtMsg.StringSize,
						ArrayDims:     dtMsg.ArrayDims,
					}
					return dtype, nil
				}
			}
		}
	}

	return Datatype{}, ErrNotFound
}

func (g *group) Keys() []string {
	if g.links == nil {
		return nil
	}
	keys := make([]string, 0, len(g.links))
	for key := range g.links {
		keys = append(keys, key)
	}
	return keys
}

func (g *group) Visit(fn func(string, interface{}) error) error {
	if g.closed {
		return ErrClosedGroup
	}

	for name, linkInfo := range g.links {
		var obj interface{}
		var err error

		switch linkInfo.LinkType {
		case format.LinkTypeHard:
			oh, readErr := format.ReadObjectHeader(g.file.reader, g.super, linkInfo.TargetAddress)
			if readErr != nil {
				return readErr
			}

			if isGroup(oh) {
				obj, err = g.GetGroup(name)
			} else {
				obj, err = g.GetDataset(name)
			}
			if err != nil {
				return err
			}

		case format.LinkTypeSoft:
			obj, err = g.file.root.GetDataset(linkInfo.TargetPath)
			if err != nil {
				obj, err = g.file.root.GetGroup(linkInfo.TargetPath)
				if err != nil {
					continue
				}
			}

		case format.LinkTypeExternal:
			continue
		}

		var fullPath string
		if g.name == "/" {
			fullPath = "/" + name
		} else {
			fullPath = g.name + "/" + name
		}

		if err := fn(fullPath, obj); err != nil {
			return err
		}

		if grp, ok := obj.(Group); ok {
			if err := grp.Visit(fn); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *group) VisitItems(fn func(name string, obj interface{}) error) error {
	if g.closed {
		return ErrClosedGroup
	}

	for name, linkInfo := range g.links {
		var obj interface{}
		var err error

		switch linkInfo.LinkType {
		case format.LinkTypeHard:
			oh, readErr := format.ReadObjectHeader(g.file.reader, g.super, linkInfo.TargetAddress)
			if readErr != nil {
				return readErr
			}

			if isGroup(oh) {
				obj, err = g.GetGroup(name)
			} else {
				obj, err = g.GetDataset(name)
			}
			if err != nil {
				return err
			}

		case format.LinkTypeSoft:
			obj, err = g.file.root.GetDataset(linkInfo.TargetPath)
			if err != nil {
				obj, err = g.file.root.GetGroup(linkInfo.TargetPath)
				if err != nil {
					continue
				}
			}

		case format.LinkTypeExternal:
			continue
		}

		if err := fn(name, obj); err != nil {
			return err
		}
	}

	return nil
}

func isGroup(oh *format.ObjectHeader) bool {
	for _, msg := range oh.Messages {
		if msg.Type == format.MessageLinkMessage {
			return true
		}
	}
	return false
}

func (g *group) CreateReference() (ObjectReference, error) {
	if g.closed {
		return ObjectReference{}, ErrClosedGroup
	}

	return ObjectReference{
		ObjectAddr: g.addr,
		ObjectType: ObjectTypeGroup,
	}, nil
}

func (g *group) Dereference(ref ObjectReference) (interface{}, error) {
	if g.closed {
		return nil, ErrClosedGroup
	}

	oh, err := format.ReadObjectHeader(g.file.reader, g.super, ref.ObjectAddr)
	if err != nil {
		return nil, err
	}

	if isGroup(oh) {
		return g.file.root.GetGroupByAddr(ref.ObjectAddr)
	}

	return g.file.root.GetDatasetByAddr(ref.ObjectAddr)
}

func (g *group) GetGroupByAddr(addr uint64) (Group, error) {
	oh, err := format.ReadObjectHeader(g.file.reader, g.super, addr)
	if err != nil {
		return nil, err
	}

	var name string
	if addr == g.file.super.RootGroupAddr {
		name = "/"
	}

	return &group{
		name:   name,
		addr:   addr,
		file:   g.file,
		header: oh,
		super:  g.super,
		links:  parseLinksFromObjectHeader(oh),
	}, nil
}

func (g *group) GetDatasetByAddr(addr uint64) (Dataset, error) {
	oh, err := format.ReadObjectHeader(g.file.reader, g.super, addr)
	if err != nil {
		return nil, err
	}

	dset, err := parseDatasetFromObjectHeader(oh, g.file, addr)
	if err != nil {
		return nil, err
	}

	return dset, nil
}

func (g *group) Attributes() *AttributeManager {
	return &AttributeManager{obj: g}
}

func (g *group) Close() error {
	if g.closed {
		return nil
	}
	g.closed = true
	return nil
}

func (g *group) Copy(dest Group, name string) error {
	return g.CopyWithOptions(dest, name, CopyOptions{
		CopyAttrs:    true,
		CopyChildren: true,
		Recursive:    true,
	})
}

func (g *group) CopyWithOptions(dest Group, name string, options CopyOptions) error {
	if g.closed {
		return ErrClosedGroup
	}

	destGroup, ok := dest.(*group)
	if !ok {
		return errors.New("hdf5: invalid destination group")
	}

	if !options.Recursive {
		return g.copySingleObject(destGroup, name, options)
	}

	for linkName, linkInfo := range g.links {
		fullDestPath := name
		if name != "" {
			fullDestPath = name + "/" + linkName
		} else {
			fullDestPath = linkName
		}

		switch linkInfo.LinkType {
		case format.LinkTypeHard:
			oh, err := format.ReadObjectHeader(g.file.reader, g.super, linkInfo.TargetAddress)
			if err != nil {
				return err
			}

			if isGroup(oh) {
				if !options.CopyChildren {
					continue
				}
				subGroup, err := g.GetGroup(linkName)
				if err != nil {
					return err
				}
				newGroup, err := destGroup.CreateGroup(fullDestPath)
				if err != nil {
					return err
				}
				if err := subGroup.(*group).CopyWithOptions(newGroup, "", options); err != nil {
					return err
				}
			} else {
				dset, err := g.GetDataset(linkName)
				if err != nil {
					return err
				}

				newDset, err := g.copyDataset(dset, destGroup, fullDestPath, options)
				if err != nil {
					return err
				}
				newDset.Close()
			}
		case format.LinkTypeSoft:
			if options.ExpandSoftLinks {
				targetGroup, err := g.file.root.GetGroup(linkInfo.TargetPath)
				if err == nil {
					newGroup, err := destGroup.CreateGroup(fullDestPath)
					if err != nil {
						return err
					}
					if err := targetGroup.(*group).CopyWithOptions(newGroup, "", options); err != nil {
						return err
					}
					continue
				}
				targetDset, err := g.file.root.GetDataset(linkInfo.TargetPath)
				if err != nil {
					return err
				}
				newDset, err := g.copyDataset(targetDset, destGroup, fullDestPath, options)
				if err != nil {
					return err
				}
				newDset.Close()
			} else {
				if err := destGroup.CreateLink(linkInfo.TargetPath, fullDestPath); err != nil {
					return err
				}
			}
		case format.LinkTypeExternal:
			if options.ExpandExternalLinks {
				extFile, err := OpenFile(linkInfo.ExternalFile, ReadOnly)
				if err != nil {
					return err
				}
				defer extFile.Close()

				targetGroup, err := extFile.root.GetGroup(linkInfo.TargetPath)
				if err == nil {
					newGroup, err := destGroup.CreateGroup(fullDestPath)
					if err != nil {
						return err
					}
					if err := targetGroup.(*group).CopyWithOptions(newGroup, "", options); err != nil {
						return err
					}
					continue
				}
				targetDset, err := extFile.root.GetDataset(linkInfo.TargetPath)
				if err != nil {
					return err
				}
				newDset, err := g.copyDataset(targetDset, destGroup, fullDestPath, options)
				if err != nil {
					return err
				}
				newDset.Close()
			} else {
				if err := destGroup.CreateExternalLink(linkInfo.ExternalFile, linkInfo.TargetPath, fullDestPath); err != nil {
					return err
				}
			}
		}
	}

	if options.CopyAttrs {
		g.copyAttributes(destGroup)
	}

	return nil
}

func (g *group) copySingleObject(destGroup *group, name string, options CopyOptions) error {
	newGroup, err := destGroup.CreateGroup(name)
	if err != nil {
		return err
	}
	if options.CopyAttrs {
		g.copyAttributes(newGroup)
	}

	if !options.CopyChildren {
		return nil
	}

	for linkName, linkInfo := range g.links {
		fullDestPath := name + "/" + linkName

		if linkInfo.LinkType == format.LinkTypeHard {
			if options.HardLink {
				if err := destGroup.CreateLink(g.name+"/"+linkName, fullDestPath); err != nil {
					return err
				}
			} else {
				oh, err := format.ReadObjectHeader(g.file.reader, g.super, linkInfo.TargetAddress)
				if err != nil {
					return err
				}

				if isGroup(oh) {
					subGroup, err := g.GetGroup(linkName)
					if err != nil {
						return err
					}
					subNewGroup, err := destGroup.CreateGroup(fullDestPath)
					if err != nil {
						return err
					}
					if err := subGroup.(*group).CopyWithOptions(subNewGroup, "", options); err != nil {
						return err
					}
				} else {
					dset, err := g.GetDataset(linkName)
					if err != nil {
						return err
					}
					newDset, err := g.copyDataset(dset, destGroup, fullDestPath, options)
					if err != nil {
						return err
					}
					newDset.Close()
				}
			}
		}
	}

	return nil
}

func (g *group) copyDataset(dset Dataset, destGroup *group, name string, options CopyOptions) (Dataset, error) {
	dtype := dset.Datatype()
	dspace := dset.Dataspace()

	dataSize := dspace.Size() * uint64(dtype.Size)
	data := make([]byte, dataSize)

	if _, err := g.file.reader.Seek(int64(dset.(*dataset).layout.DataAddress), 0); err != nil {
		return nil, err
	}
	if _, err := g.file.reader.Read(data); err != nil {
		return nil, err
	}

	newDset, err := destGroup.CreateDataset(name, dtype, dspace, PropertyList{})
	if err != nil {
		return nil, err
	}

	if _, err := g.file.writer.Seek(int64(newDset.(*dataset).layout.DataAddress), 0); err != nil {
		newDset.Close()
		return nil, err
	}
	if _, err := g.file.writer.Write(data); err != nil {
		newDset.Close()
		return nil, err
	}

	if options.CopyAttrs {
		dset.(*dataset).copyAttributes(newDset.(*dataset))
	}

	return newDset, nil
}

func (g *group) copyAttributes(dest Group) {
	srcAttrs := g.Attributes()
	destAttrs := dest.Attributes()

	for _, name := range srcAttrs.Keys() {
		attr, err := srcAttrs.Get(name)
		if err != nil {
			continue
		}
		var value interface{}
		if err := attr.Read(&value); err == nil {
			destAttrs.Set(name, value)
		}
	}
}

func (g *group) Move(dest Group, name string) error {
	if g.closed {
		return ErrClosedGroup
	}

	if err := g.Copy(dest, name); err != nil {
		return err
	}

	for linkName := range g.links {
		if err := g.Delete(linkName); err != nil {
			return err
		}
	}

	return nil
}

func (g *group) Delete(name string) error {
	if g.closed {
		return ErrClosedGroup
	}

	if _, ok := g.links[name]; !ok {
		return ErrNotFound
	}

	delete(g.links, name)

	newMessages := make([]format.ObjectHeaderMessage, 0, len(g.header.Messages))
	for _, msg := range g.header.Messages {
		if msg.Type == format.MessageLinkMessage {
			linkMsg, err := format.DecodeLinkMessage(msg.Content)
			if err == nil && linkMsg.Name == name {
				continue
			}
		}
		newMessages = append(newMessages, msg)
	}

	g.header.Messages = newMessages
	g.header.NumberOfMessages = uint16(len(newMessages))

	actualSize := uint64(4) + uint64(1) + uint64(1) + uint64(g.super.LengthSize) + uint64(2) + uint64(g.super.OffsetSize)*uint64(g.header.NumberOfMessages)
	for _, msg := range g.header.Messages {
		actualSize += uint64(msg.Size)
	}
	if g.header.TotalHeaderSize < actualSize {
		g.header.TotalHeaderSize = ((actualSize + 1023) / 1024) * 1024
	}

	if _, err := g.file.writer.Seek(int64(g.addr), 0); err != nil {
		return err
	}
	if _, err := format.WriteObjectHeader(g.file.writer, g.super, g.header); err != nil {
		return err
	}

	endPos := int64(g.addr) + int64(g.header.TotalHeaderSize)
	if _, err := g.file.writer.Seek(endPos, 0); err != nil {
		return err
	}

	return nil
}

func parseLinksFromObjectHeader(oh *format.ObjectHeader) map[string]LinkInfo {
	links := make(map[string]LinkInfo)
	for _, msg := range oh.Messages {
		if msg.Type == format.MessageLinkMessage {
			linkMsg, err := format.DecodeLinkMessage(msg.Content)
			if err != nil {
				continue
			}
			links[linkMsg.Name] = LinkInfo{
				LinkType:      linkMsg.LinkType,
				TargetAddress: linkMsg.TargetAddress,
				TargetPath:    linkMsg.TargetPath,
				ExternalFile:  linkMsg.ExternalFile,
			}
		}
	}
	return links
}

func parsePath(path string) (string, string) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if path == "/" {
		return "/", ""
	}

	parts := strings.Split(path[1:], "/")
	if len(parts) == 1 {
		return "/", parts[0]
	}

	parentPath := "/" + strings.Join(parts[:len(parts)-1], "/")
	return parentPath, parts[len(parts)-1]
}
