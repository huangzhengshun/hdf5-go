package format

import (
	"encoding/binary"
	"errors"
)

const (
	DatatypeVersion1 = 1
	DatatypeVersion2 = 2
	DatatypeVersion3 = 3
)

const (
	ClassInteger   = 0
	ClassFloat     = 1
	ClassString    = 2
	ClassArray     = 3
	ClassCompound  = 4
	ClassEnum      = 5
	ClassReference = 6
	ClassOpaque    = 7
	ClassBitfield  = 8
	ClassVarLength = 9
)

const (
	OrderLittleEndian = 0
	OrderBigEndian    = 1
	OrderNative       = 3
)

type CompoundField struct {
	Name     string
	Offset   uint32
	Datatype DatatypeMessage
}

type EnumMember struct {
	Name  string
	Value int64
}

type DatatypeMessage struct {
	Version       uint8
	Class         uint8
	ByteOrder     uint8
	Size          uint32
	Sign          uint8
	FloatBits     uint8
	StringPadding uint8
	StringSize    uint32
	Fields        []CompoundField
	EnumBase      *DatatypeMessage
	EnumMembers   []EnumMember
	ArrayBase     *DatatypeMessage
	ArrayDims     []uint64
}

func EncodeDatatypeMessage(dt DatatypeMessage) ([]byte, error) {
	buf := make([]byte, 16)
	buf[0] = dt.Version

	switch dt.Version {
	case DatatypeVersion1:
		return encodeDatatypeV1(dt)
	case DatatypeVersion2:
		return encodeDatatypeV2(dt)
	case DatatypeVersion3:
		return encodeDatatypeV3(dt)
	default:
		return nil, errors.New("unsupported datatype version")
	}
}

func encodeDatatypeV1(dt DatatypeMessage) ([]byte, error) {
	buf := make([]byte, 16)
	buf[0] = dt.Version
	buf[1] = dt.Class
	buf[2] = dt.ByteOrder
	buf[3] = byte(dt.Size & 0xFF)
	buf[4] = byte((dt.Size >> 8) & 0xFF)

	switch dt.Class {
	case ClassInteger:
		buf[5] = dt.Sign
		buf[6] = 0
		buf[7] = byte(dt.Size)
	case ClassFloat:
		buf[5] = dt.FloatBits
		buf[6] = byte(dt.Size)
		buf[7] = 0
	case ClassString:
		buf[5] = dt.StringPadding
		binary.LittleEndian.PutUint16(buf[6:8], uint16(dt.StringSize))
	}

	return buf, nil
}

func encodeDatatypeV2(dt DatatypeMessage) ([]byte, error) {
	buf := make([]byte, 16)
	buf[0] = dt.Version
	buf[1] = dt.Class
	buf[2] = dt.ByteOrder

	binary.LittleEndian.PutUint32(buf[4:8], dt.Size)

	switch dt.Class {
	case ClassInteger:
		buf[8] = dt.Sign
		buf[12] = byte(dt.Size)
	case ClassFloat:
		buf[8] = dt.FloatBits
		buf[12] = byte(dt.Size)
	case ClassString:
		buf[8] = dt.StringPadding
		binary.LittleEndian.PutUint32(buf[12:16], dt.StringSize)
	case ClassCompound:
		buf[8] = 0
		binary.LittleEndian.PutUint32(buf[12:16], uint32(len(dt.Fields)))

		for _, field := range dt.Fields {
			nameLen := len(field.Name) + 1
			nameBuf := make([]byte, nameLen)
			copy(nameBuf, field.Name)

			fieldDtEncoded, err := EncodeDatatypeMessage(field.Datatype)
			if err != nil {
				return nil, err
			}

			fieldBuf := make([]byte, 4+nameLen+4+len(fieldDtEncoded))
			binary.LittleEndian.PutUint32(fieldBuf[0:4], uint32(nameLen))
			copy(fieldBuf[4:4+nameLen], nameBuf)
			binary.LittleEndian.PutUint32(fieldBuf[4+nameLen:8+nameLen], field.Offset)
			copy(fieldBuf[8+nameLen:], fieldDtEncoded)

			buf = append(buf, fieldBuf...)
		}
	case ClassEnum:
		buf[8] = 0
		binary.LittleEndian.PutUint32(buf[12:16], uint32(len(dt.EnumMembers)))

		if dt.EnumBase != nil {
			baseEncoded, err := EncodeDatatypeMessage(*dt.EnumBase)
			if err != nil {
				return nil, err
			}
			buf = append(buf, baseEncoded...)
		}

		for _, member := range dt.EnumMembers {
			nameLen := len(member.Name) + 1
			nameBuf := make([]byte, nameLen)
			copy(nameBuf, member.Name)

			memberBuf := make([]byte, 4+nameLen+8)
			binary.LittleEndian.PutUint32(memberBuf[0:4], uint32(nameLen))
			copy(memberBuf[4:4+nameLen], nameBuf)
			binary.LittleEndian.PutUint64(memberBuf[4+nameLen:12+nameLen], uint64(member.Value))

			buf = append(buf, memberBuf...)
		}
	case ClassArray:
		buf[8] = 0
		binary.LittleEndian.PutUint32(buf[12:16], uint32(len(dt.ArrayDims)))

		if dt.ArrayBase != nil {
			baseEncoded, err := EncodeDatatypeMessage(*dt.ArrayBase)
			if err != nil {
				return nil, err
			}
			buf = append(buf, baseEncoded...)
		}

		for _, dim := range dt.ArrayDims {
			dimBuf := make([]byte, 8)
			binary.LittleEndian.PutUint64(dimBuf, dim)
			buf = append(buf, dimBuf...)
		}
	case ClassVarLength:
		buf[8] = 0
		binary.LittleEndian.PutUint32(buf[12:16], 0)

		if dt.ArrayBase != nil {
			baseEncoded, err := EncodeDatatypeMessage(*dt.ArrayBase)
			if err != nil {
				return nil, err
			}
			buf = append(buf, baseEncoded...)
		}
	}

	return buf, nil
}

func encodeDatatypeV3(dt DatatypeMessage) ([]byte, error) {
	buf := make([]byte, 16)
	buf[0] = dt.Version
	buf[1] = dt.Class
	buf[2] = dt.ByteOrder

	binary.LittleEndian.PutUint32(buf[4:8], dt.Size)

	switch dt.Class {
	case ClassInteger:
		buf[8] = dt.Sign
		buf[12] = byte(dt.Size)
	case ClassFloat:
		buf[8] = dt.FloatBits
		buf[12] = byte(dt.Size)
	case ClassString:
		buf[8] = dt.StringPadding
		binary.LittleEndian.PutUint32(buf[12:16], dt.StringSize)
	}

	return buf, nil
}

func DecodeDatatypeMessage(data []byte) (DatatypeMessage, error) {
	if len(data) < 8 {
		return DatatypeMessage{}, errors.New("datatype message too short")
	}

	var dt DatatypeMessage
	dt.Version = data[0]

	switch dt.Version {
	case DatatypeVersion1:
		return decodeDatatypeV1(data)
	case DatatypeVersion2, DatatypeVersion3:
		return decodeDatatypeV2(data)
	default:
		return dt, errors.New("unsupported datatype version")
	}
}

func decodeDatatypeV1(data []byte) (DatatypeMessage, error) {
	var dt DatatypeMessage
	dt.Version = DatatypeVersion1
	dt.Class = data[1]
	dt.ByteOrder = data[2]
	dt.Size = uint32(data[3]) | uint32(data[4])<<8

	switch dt.Class {
	case ClassInteger:
		dt.Sign = data[5]
	case ClassFloat:
		dt.FloatBits = data[5]
	case ClassString:
		dt.StringPadding = data[5]
		dt.StringSize = uint32(binary.LittleEndian.Uint16(data[6:8]))
	}

	return dt, nil
}

func decodeDatatypeV2(data []byte) (DatatypeMessage, error) {
	var dt DatatypeMessage
	dt.Version = data[0]
	dt.Class = data[1]
	dt.ByteOrder = data[2]
	dt.Size = binary.LittleEndian.Uint32(data[4:8])

	switch dt.Class {
	case ClassInteger:
		dt.Sign = data[8]
		dt.Size = uint32(data[12])
	case ClassFloat:
		dt.FloatBits = data[8]
	case ClassString:
		dt.StringPadding = data[8]
		dt.StringSize = binary.LittleEndian.Uint32(data[12:16])
	case ClassCompound:
		numFields := binary.LittleEndian.Uint32(data[12:16])
		dt.Fields = make([]CompoundField, 0, numFields)

		offset := 16
		for i := uint32(0); i < numFields; i++ {
			if offset+4 > len(data) {
				return dt, errors.New("compound datatype truncated")
			}
			nameLen := binary.LittleEndian.Uint32(data[offset : offset+4])
			offset += 4

			if offset+int(nameLen) > len(data) {
				return dt, errors.New("compound datatype truncated")
			}
			name := string(data[offset : offset+int(nameLen)-1])
			offset += int(nameLen)

			if offset+4 > len(data) {
				return dt, errors.New("compound datatype truncated")
			}
			fieldOffset := binary.LittleEndian.Uint32(data[offset : offset+4])
			offset += 4

			if offset+16 > len(data) {
				return dt, errors.New("compound datatype truncated")
			}
			fieldDtVersion := data[offset]
			var fieldDtSize int
			if fieldDtVersion == DatatypeVersion1 {
				fieldDtSize = 16
			} else {
				fieldDtSize = 16
				if offset+fieldDtSize > len(data) {
					return dt, errors.New("compound datatype truncated")
				}
				fieldClass := data[offset+1]
				if fieldClass == ClassCompound {
					numNestedFields := binary.LittleEndian.Uint32(data[offset+12 : offset+16])
					fieldDtSize = 16
					for j := uint32(0); j < numNestedFields; j++ {
						if offset+fieldDtSize+4 > len(data) {
							return dt, errors.New("compound datatype truncated")
						}
						nestedNameLen := binary.LittleEndian.Uint32(data[offset+fieldDtSize : offset+fieldDtSize+4])
						fieldDtSize += 4 + int(nestedNameLen) + 4 + 16
					}
				}
			}

			if offset+fieldDtSize > len(data) {
				return dt, errors.New("compound datatype truncated")
			}

			fieldDt, err := DecodeDatatypeMessage(data[offset : offset+fieldDtSize])
			if err != nil {
				return dt, err
			}
			offset += fieldDtSize

			dt.Fields = append(dt.Fields, CompoundField{
				Name:     name,
				Offset:   fieldOffset,
				Datatype: fieldDt,
			})
		}
	case ClassEnum:
		numMembers := binary.LittleEndian.Uint32(data[12:16])
		dt.EnumMembers = make([]EnumMember, 0, numMembers)

		offset := 16
		if offset+16 <= len(data) {
			baseDt, err := DecodeDatatypeMessage(data[offset : offset+16])
			if err == nil {
				dt.EnumBase = &baseDt
			}
			offset += 16
		}

		for i := uint32(0); i < numMembers; i++ {
			if offset+4 > len(data) {
				return dt, errors.New("enum datatype truncated")
			}
			nameLen := binary.LittleEndian.Uint32(data[offset : offset+4])
			offset += 4

			if offset+int(nameLen) > len(data) {
				return dt, errors.New("enum datatype truncated")
			}
			name := string(data[offset : offset+int(nameLen)-1])
			offset += int(nameLen)

			if offset+8 > len(data) {
				return dt, errors.New("enum datatype truncated")
			}
			value := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
			offset += 8

			dt.EnumMembers = append(dt.EnumMembers, EnumMember{
				Name:  name,
				Value: value,
			})
		}
	case ClassArray:
		numDims := binary.LittleEndian.Uint32(data[12:16])
		dt.ArrayDims = make([]uint64, 0, numDims)

		offset := 16
		if offset+16 <= len(data) {
			baseDt, err := DecodeDatatypeMessage(data[offset : offset+16])
			if err == nil {
				dt.ArrayBase = &baseDt
			}
			offset += 16
		}

		for i := uint32(0); i < numDims; i++ {
			if offset+8 > len(data) {
				return dt, errors.New("array datatype truncated")
			}
			dim := binary.LittleEndian.Uint64(data[offset : offset+8])
			offset += 8
			dt.ArrayDims = append(dt.ArrayDims, dim)
		}
	case ClassVarLength:
		offset := 16
		if offset+16 <= len(data) {
			baseDt, err := DecodeDatatypeMessage(data[offset : offset+16])
			if err == nil {
				dt.ArrayBase = &baseDt
			}
		}
	case ClassReference:
		dt.Size = 8
	}

	return dt, nil
}
