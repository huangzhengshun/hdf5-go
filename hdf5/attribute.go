package hdf5

import (
	"errors"
	"reflect"

	"github.com/hdf5-go/hdf5/format"
)

var (
	ErrAttributeNotFound = errors.New("hdf5: attribute not found")
	ErrInvalidAttribute  = errors.New("hdf5: invalid attribute")
)

type attribute struct {
	name   string
	dtype  Datatype
	dspace Dataspace
	data   []byte
	parent interface{}
}

func (a *attribute) Name() string {
	return a.name
}

func (a *attribute) Datatype() Datatype {
	return a.dtype
}

func (a *attribute) Dataspace() Dataspace {
	return a.dspace
}

func (a *attribute) Read(data interface{}) error {
	return decodeData(a.data, data, a.dtype)
}

func (a *attribute) Write(data interface{}) error {
	encoded, err := encodeData(data, a.dtype, a.dspace)
	if err != nil {
		return err
	}
	a.data = encoded
	return nil
}

func (am *AttributeManager) Get(name string) (Attribute, error) {
	var header *format.ObjectHeader

	switch obj := am.obj.(type) {
	case *dataset:
		header = obj.header
	case *group:
		header = obj.header
	case *File:
		header = obj.root.(*group).header
	default:
		return nil, ErrInvalidAttribute
	}

	for _, msg := range header.Messages {
		if msg.Type == format.MessageAttribute {
			attrMsg, err := format.DecodeAttributeMessage(msg.Content)
			if err != nil {
				continue
			}
			if attrMsg.Name == name {
				return &attribute{
					name: name,
					dtype: Datatype{
						Class:     DatatypeClass(attrMsg.Datatype.Class),
						Size:      attrMsg.Datatype.Size,
						ByteOrder: ByteOrder(attrMsg.Datatype.ByteOrder),
						Sign:      Sign(attrMsg.Datatype.Sign),
						FloatBits: attrMsg.Datatype.FloatBits,
					},
					dspace: Dataspace{
						Dims:    attrMsg.Dataspace.Dims,
						MaxDims: attrMsg.Dataspace.MaxDims,
					},
					data:   attrMsg.Data,
					parent: am.obj,
				}, nil
			}
		}
	}

	return nil, ErrAttributeNotFound
}

func (am *AttributeManager) Set(name string, data interface{}) error {
	var f *File
	var header *format.ObjectHeader
	var super *format.Superblock

	switch obj := am.obj.(type) {
	case *dataset:
		f = obj.file
		header = obj.header
		super = obj.super
	case *group:
		f = obj.file
		header = obj.header
		super = obj.super
	case *File:
		f = obj
		header = obj.root.(*group).header
		super = obj.super
	default:
		return ErrInvalidAttribute
	}

	rv := reflect.ValueOf(data)
	var dtype Datatype
	var dspace Dataspace

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dtype = Datatype{Class: DatatypeInteger, Size: uint32(rv.Type().Size()), ByteOrder: LittleEndian, Sign: Signed}
		dspace = Dataspace{Dims: []uint64{}, MaxDims: []uint64{}}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		dtype = Datatype{Class: DatatypeInteger, Size: uint32(rv.Type().Size()), ByteOrder: LittleEndian, Sign: Unsigned}
		dspace = Dataspace{Dims: []uint64{}, MaxDims: []uint64{}}
	case reflect.Float32:
		dtype = Datatype{Class: DatatypeFloat, Size: 4, ByteOrder: LittleEndian, FloatBits: 24}
		dspace = Dataspace{Dims: []uint64{}, MaxDims: []uint64{}}
	case reflect.Float64:
		dtype = Datatype{Class: DatatypeFloat, Size: 8, ByteOrder: LittleEndian, FloatBits: 53}
		dspace = Dataspace{Dims: []uint64{}, MaxDims: []uint64{}}
	case reflect.Slice:
		elemKind := rv.Type().Elem().Kind()
		switch elemKind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			dtype = Datatype{Class: DatatypeInteger, Size: uint32(rv.Type().Elem().Size()), ByteOrder: LittleEndian, Sign: Signed}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			dtype = Datatype{Class: DatatypeInteger, Size: uint32(rv.Type().Elem().Size()), ByteOrder: LittleEndian, Sign: Unsigned}
		case reflect.Float32:
			dtype = Datatype{Class: DatatypeFloat, Size: 4, ByteOrder: LittleEndian, FloatBits: 24}
		case reflect.Float64:
			dtype = Datatype{Class: DatatypeFloat, Size: 8, ByteOrder: LittleEndian, FloatBits: 53}
		case reflect.String:
			dtype = Datatype{Class: DatatypeString, Size: 64, ByteOrder: LittleEndian, StringPadding: NullTerminated, StringSize: 64}
		default:
			return ErrInvalidData
		}
		dspace = Dataspace{Dims: []uint64{uint64(rv.Len())}, MaxDims: []uint64{uint64(rv.Len())}}
	case reflect.String:
		dtype = Datatype{Class: DatatypeString, Size: 64, ByteOrder: LittleEndian, StringPadding: NullTerminated, StringSize: 64}
		dspace = Dataspace{Dims: []uint64{1}, MaxDims: []uint64{1}}
	default:
		return ErrInvalidData
	}

	encoded, err := encodeData(data, dtype, dspace)
	if err != nil {
		return err
	}

	dtMsg := format.DatatypeMessage{
		Version:       format.DatatypeVersion2,
		Class:         uint8(dtype.Class),
		ByteOrder:     uint8(dtype.ByteOrder),
		Size:          dtype.Size,
		Sign:          uint8(dtype.Sign),
		FloatBits:     dtype.FloatBits,
		StringPadding: uint8(dtype.StringPadding),
		StringSize:    dtype.StringSize,
	}

	attrMsg := format.AttributeMessage{
		Name:     name,
		Datatype: dtMsg,
		Dataspace: format.DataspaceMessage{
			Type:    format.DataspaceSimple,
			Dims:    dspace.Dims,
			MaxDims: dspace.MaxDims,
		},
		Data: encoded,
	}

	if len(dspace.Dims) == 0 {
		attrMsg.Dataspace.Type = format.DataspaceScalar
	}

	attrEncoded, err := format.EncodeAttributeMessage(attrMsg)
	if err != nil {
		return err
	}

	msgSize := uint16(10 + len(attrEncoded))

	attrObjMsg := format.ObjectHeaderMessage{
		Type:    format.MessageAttribute,
		Size:    msgSize,
		Flags:   0,
		Content: attrEncoded,
	}

	header.Messages = append(header.Messages, attrObjMsg)
	header.NumberOfMessages++

	header.TotalHeaderSize = uint64(4) + uint64(1) + uint64(1) + uint64(super.LengthSize) + uint64(2) + uint64(super.OffsetSize)*uint64(header.NumberOfMessages)
	for _, msg := range header.Messages {
		header.TotalHeaderSize += uint64(msg.Size)
	}

	var addr uint64
	switch obj := am.obj.(type) {
	case *dataset:
		addr = obj.addr
	case *group:
		addr = obj.addr
	case *File:
		addr = obj.root.(*group).addr
	}

	currentPos, _ := f.writer.Seek(0, 1)
	f.writer.Seek(int64(addr), 0)
	format.WriteObjectHeader(f.writer, super, header)
	f.writer.Seek(currentPos, 0)

	return nil
}

func (am *AttributeManager) Delete(name string) error {
	var header *format.ObjectHeader
	var super *format.Superblock
	var f *File
	var addr uint64

	switch obj := am.obj.(type) {
	case *dataset:
		f = obj.file
		header = obj.header
		super = obj.super
		addr = obj.addr
	case *group:
		f = obj.file
		header = obj.header
		super = obj.super
		addr = obj.addr
	case *File:
		f = obj
		header = obj.root.(*group).header
		super = obj.super
		addr = obj.root.(*group).addr
	default:
		return ErrInvalidAttribute
	}

	newMessages := make([]format.ObjectHeaderMessage, 0, len(header.Messages))
	for _, msg := range header.Messages {
		if msg.Type == format.MessageAttribute {
			attrMsg, err := format.DecodeAttributeMessage(msg.Content)
			if err == nil && attrMsg.Name == name {
				continue
			}
		}
		newMessages = append(newMessages, msg)
	}

	header.Messages = newMessages
	header.NumberOfMessages = uint16(len(newMessages))

	header.TotalHeaderSize = uint64(4) + uint64(1) + uint64(1) + uint64(super.LengthSize) + uint64(2) + uint64(super.OffsetSize)*uint64(header.NumberOfMessages)
	for _, msg := range header.Messages {
		header.TotalHeaderSize += uint64(msg.Size)
	}

	currentPos, _ := f.writer.Seek(0, 1)
	f.writer.Seek(int64(addr), 0)
	format.WriteObjectHeader(f.writer, super, header)
	f.writer.Seek(currentPos, 0)

	return nil
}

func (am *AttributeManager) Keys() []string {
	var header *format.ObjectHeader

	switch obj := am.obj.(type) {
	case *dataset:
		header = obj.header
	case *group:
		header = obj.header
	case *File:
		header = obj.root.(*group).header
	default:
		return nil
	}

	keys := make([]string, 0)
	for _, msg := range header.Messages {
		if msg.Type == format.MessageAttribute {
			attrMsg, err := format.DecodeAttributeMessage(msg.Content)
			if err == nil {
				keys = append(keys, attrMsg.Name)
			}
		}
	}

	return keys
}

func (am *AttributeManager) Len() int {
	return len(am.Keys())
}

func (am *AttributeManager) Iter() ([]Attribute, error) {
	var header *format.ObjectHeader

	switch obj := am.obj.(type) {
	case *dataset:
		header = obj.header
	case *group:
		header = obj.header
	case *File:
		header = obj.root.(*group).header
	default:
		return nil, ErrInvalidAttribute
	}

	attrs := make([]Attribute, 0)
	for _, msg := range header.Messages {
		if msg.Type == format.MessageAttribute {
			attrMsg, err := format.DecodeAttributeMessage(msg.Content)
			if err != nil {
				continue
			}
			attrs = append(attrs, &attribute{
				name: attrMsg.Name,
				dtype: Datatype{
					Class:     DatatypeClass(attrMsg.Datatype.Class),
					Size:      attrMsg.Datatype.Size,
					ByteOrder: ByteOrder(attrMsg.Datatype.ByteOrder),
					Sign:      Sign(attrMsg.Datatype.Sign),
					FloatBits: attrMsg.Datatype.FloatBits,
				},
				dspace: Dataspace{
					Dims:    attrMsg.Dataspace.Dims,
					MaxDims: attrMsg.Dataspace.MaxDims,
				},
				data:   attrMsg.Data,
				parent: am.obj,
			})
		}
	}

	return attrs, nil
}

func (am *AttributeManager) Copy(dest *AttributeManager, name string) error {
	attr, err := am.Get(name)
	if err != nil {
		return err
	}

	var data interface{}
	if err := attr.Read(&data); err != nil {
		return err
	}

	return dest.Set(name, data)
}


