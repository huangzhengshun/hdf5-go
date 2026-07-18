package format

import (
	"encoding/binary"
	"errors"
)

const (
	DataspaceScalar = 0
	DataspaceSimple = 1
	DataspaceNull   = 2
)

type DataspaceMessage struct {
	Type    uint8
	Dims    []uint64
	MaxDims []uint64
}

func EncodeDataspaceMessage(dims, maxDims []uint64) ([]byte, error) {
	ndims := len(dims)
	if ndims == 0 {
		return []byte{DataspaceScalar, 0, 0, 0}, nil
	}

	if maxDims == nil {
		maxDims = make([]uint64, ndims)
		copy(maxDims, dims)
	}

	buf := make([]byte, 8+16*ndims)
	buf[0] = DataspaceSimple
	binary.LittleEndian.PutUint32(buf[4:8], uint32(ndims))

	for i := 0; i < ndims; i++ {
		binary.LittleEndian.PutUint64(buf[8+8*i:8+8*(i+1)], dims[i])
		binary.LittleEndian.PutUint64(buf[8+8*ndims+8*i:8+8*ndims+8*(i+1)], maxDims[i])
	}

	return buf, nil
}

func DecodeDataspaceMessage(data []byte) (DataspaceMessage, error) {
	if len(data) < 4 {
		return DataspaceMessage{}, errors.New("dataspace message too short")
	}

	var ds DataspaceMessage
	ds.Type = data[0]

	switch ds.Type {
	case DataspaceScalar:
		ds.Dims = []uint64{}
		ds.MaxDims = []uint64{}
	case DataspaceSimple:
		ndims := binary.LittleEndian.Uint32(data[4:8])
		ds.Dims = make([]uint64, ndims)
		ds.MaxDims = make([]uint64, ndims)

		for i := uint32(0); i < ndims; i++ {
			ds.Dims[i] = binary.LittleEndian.Uint64(data[8+8*i : 8+8*(i+1)])
			ds.MaxDims[i] = binary.LittleEndian.Uint64(data[8+8*ndims+8*i : 8+8*ndims+8*(i+1)])
		}
	case DataspaceNull:
		ds.Dims = []uint64{}
		ds.MaxDims = []uint64{}
	default:
		return ds, errors.New("unsupported dataspace type")
	}

	return ds, nil
}
