package format

import (
	"encoding/binary"
	"errors"
)

const (
	LayoutCompact    = 0
	LayoutContiguous = 1
	LayoutChunked    = 2
)

type LayoutMessage struct {
	LayoutClass uint8
	DataSize    uint64
	DataAddress uint64
	ChunkDims   []uint64
	ChunkSize   uint32
}

func EncodeLayoutMessage(layoutClass uint8, dataSize, dataAddress uint64, chunkDims []uint64) ([]byte, error) {
	switch layoutClass {
	case LayoutCompact:
		buf := make([]byte, 13)
		buf[0] = layoutClass
		binary.LittleEndian.PutUint32(buf[1:5], uint32(dataSize))
		binary.LittleEndian.PutUint64(buf[5:13], dataAddress)
		return buf, nil
	case LayoutContiguous:
		buf := make([]byte, 17)
		buf[0] = layoutClass
		binary.LittleEndian.PutUint64(buf[1:9], dataAddress)
		binary.LittleEndian.PutUint64(buf[9:17], dataSize)
		return buf, nil
	case LayoutChunked:
		ndims := len(chunkDims)
		chunkSize := uint32(1)
		for _, dim := range chunkDims {
			chunkSize *= uint32(dim)
		}

		buf := make([]byte, 25+8*ndims)
		buf[0] = layoutClass
		binary.LittleEndian.PutUint32(buf[1:5], chunkSize)
		binary.LittleEndian.PutUint32(buf[5:9], uint32(ndims))
		binary.LittleEndian.PutUint64(buf[9:17], dataAddress)
		binary.LittleEndian.PutUint64(buf[17:25], dataSize)

		for i := 0; i < ndims; i++ {
			binary.LittleEndian.PutUint64(buf[25+8*i:33+8*i], chunkDims[i])
		}

		return buf, nil
	default:
		return nil, errors.New("unsupported layout class")
	}
}

func DecodeLayoutMessage(data []byte) (LayoutMessage, error) {
	if len(data) < 4 {
		return LayoutMessage{}, errors.New("layout message too short")
	}

	var lm LayoutMessage
	lm.LayoutClass = data[0]

	switch lm.LayoutClass {
	case LayoutCompact:
		lm.DataSize = uint64(binary.LittleEndian.Uint32(data[1:5]))
		lm.DataAddress = binary.LittleEndian.Uint64(data[5:13])
	case LayoutContiguous:
		lm.DataAddress = binary.LittleEndian.Uint64(data[1:9])
		lm.DataSize = binary.LittleEndian.Uint64(data[9:17])
	case LayoutChunked:
		lm.ChunkSize = binary.LittleEndian.Uint32(data[1:5])
		ndims := binary.LittleEndian.Uint32(data[5:9])
		lm.DataAddress = binary.LittleEndian.Uint64(data[9:17])
		lm.DataSize = binary.LittleEndian.Uint64(data[17:25])

		lm.ChunkDims = make([]uint64, ndims)
		for i := uint32(0); i < ndims; i++ {
			lm.ChunkDims[i] = binary.LittleEndian.Uint64(data[25+8*i : 33+8*i])
		}
	default:
		return lm, errors.New("unsupported layout class")
	}

	return lm, nil
}
