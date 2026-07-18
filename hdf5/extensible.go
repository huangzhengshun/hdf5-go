package hdf5

import (
	"errors"

	"github.com/huangzhengshun/hdf5-go/format"
)

type ExtensibleDatatype struct {
	Datatype
	OriginalSize   uint32
	Extensible     bool
	CurrentVersion uint8
	Versions       []Datatype
}

func NewExtensibleDatatype(base Datatype) ExtensibleDatatype {
	return ExtensibleDatatype{
		Datatype:       base,
		OriginalSize:   base.Size,
		Extensible:     true,
		CurrentVersion: 1,
		Versions:       []Datatype{base},
	}
}

func (ed *ExtensibleDatatype) AddField(name string, dtype Datatype, offset uint32) error {
	if !ed.Extensible {
		return errors.New("hdf5: datatype is not extensible")
	}

	if ed.Datatype.Class != DatatypeCompound {
		return errors.New("hdf5: extensible datatype only supports compound types")
	}

	newFields := make([]CompoundField, len(ed.Datatype.Fields)+1)
	copy(newFields, ed.Datatype.Fields)
	newFields[len(newFields)-1] = CompoundField{
		Name:     name,
		Offset:   offset,
		Datatype: dtype,
	}

	newSize := offset + dtype.Size
	if newSize < ed.Datatype.Size {
		newSize = ed.Datatype.Size
	}

	newDtype := Datatype{
		Class:     DatatypeCompound,
		Size:      newSize,
		Fields:    newFields,
		ByteOrder: ed.Datatype.ByteOrder,
	}

	ed.Versions = append(ed.Versions, newDtype)
	ed.Datatype = newDtype
	ed.CurrentVersion++

	return nil
}

func (ed *ExtensibleDatatype) GetVersion(version uint8) (Datatype, error) {
	if version == 0 || version > uint8(len(ed.Versions)) {
		return Datatype{}, errors.New("hdf5: invalid version")
	}
	return ed.Versions[version-1], nil
}

func (ed *ExtensibleDatatype) Encode() ([]byte, error) {
	return format.EncodeDatatypeMessage(format.DatatypeMessage{
		Version:   format.DatatypeVersion2,
		Class:     uint8(ed.Datatype.Class),
		Size:      ed.Datatype.Size,
		ByteOrder: uint8(ed.Datatype.ByteOrder),
	})
}

func decodeExtensibleDatatype(data []byte) (ExtensibleDatatype, error) {
	msg, err := format.DecodeDatatypeMessage(data)
	if err != nil {
		return ExtensibleDatatype{}, err
	}

	dtype := Datatype{
		Class:     DatatypeClass(msg.Class),
		Size:      msg.Size,
		ByteOrder: ByteOrder(msg.ByteOrder),
	}

	return ExtensibleDatatype{
		Datatype:       dtype,
		OriginalSize:   msg.Size,
		Extensible:     false,
		CurrentVersion: msg.Version,
		Versions:       []Datatype{dtype},
	}, nil
}
