package format

import (
	"encoding/binary"
	"errors"
)

type AttributeMessage struct {
	Name      string
	Datatype  DatatypeMessage
	Dataspace DataspaceMessage
	Data      []byte
}

func EncodeAttributeMessage(attr AttributeMessage) ([]byte, error) {
	nameLen := len(attr.Name)
	nameBytes := make([]byte, (nameLen+7)/8*8)
	copy(nameBytes, attr.Name)

	dtEncoded, err := EncodeDatatypeMessage(attr.Datatype)
	if err != nil {
		return nil, err
	}

	dsEncoded, err := EncodeDataspaceMessage(attr.Dataspace.Dims, attr.Dataspace.MaxDims)
	if err != nil {
		return nil, err
	}

	dataSize := len(attr.Data)
	dataBytes := make([]byte, (dataSize+7)/8*8)
	copy(dataBytes, attr.Data)

	totalSize := 4 + len(nameBytes) + len(dtEncoded) + len(dsEncoded) + len(dataBytes)
	buf := make([]byte, totalSize)

	offset := 0
	binary.LittleEndian.PutUint32(buf[offset:offset+4], uint32(nameLen))
	offset += 4

	copy(buf[offset:offset+len(nameBytes)], nameBytes)
	offset += len(nameBytes)

	copy(buf[offset:offset+len(dtEncoded)], dtEncoded)
	offset += len(dtEncoded)

	copy(buf[offset:offset+len(dsEncoded)], dsEncoded)
	offset += len(dsEncoded)

	copy(buf[offset:offset+len(dataBytes)], dataBytes)

	return buf, nil
}

func DecodeAttributeMessage(data []byte) (AttributeMessage, error) {
	if len(data) < 4 {
		return AttributeMessage{}, errors.New("attribute message too short")
	}

	var attr AttributeMessage
	offset := 0

	nameLen := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	nameBytesLen := (int(nameLen) + 7) / 8 * 8
	if offset+nameBytesLen > len(data) {
		return attr, errors.New("attribute message truncated")
	}

	attr.Name = string(data[offset : offset+int(nameLen)])
	offset += nameBytesLen

	dt, err := DecodeDatatypeMessage(data[offset:])
	if err != nil {
		return attr, err
	}
	attr.Datatype = dt

	dtEncodedLen := 16
	if offset+dtEncodedLen > len(data) {
		return attr, errors.New("attribute message truncated")
	}
	offset += dtEncodedLen

	ds, err := DecodeDataspaceMessage(data[offset:])
	if err != nil {
		return attr, err
	}
	attr.Dataspace = ds

	dsEncodedLen := 8 + 16*len(ds.Dims)
	if ds.Type == DataspaceScalar {
		dsEncodedLen = 4
	}
	if offset+dsEncodedLen > len(data) {
		return attr, errors.New("attribute message truncated")
	}
	offset += dsEncodedLen

	dataSize := len(data) - offset
	if dataSize < 0 {
		dataSize = 0
	}
	attr.Data = make([]byte, dataSize)
	copy(attr.Data, data[offset:offset+dataSize])

	return attr, nil
}
