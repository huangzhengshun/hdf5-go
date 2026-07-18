package format

import (
	"encoding/binary"
	"errors"
	"io"
)

type ChunkIndexEntry struct {
	ChunkOffset uint64
	ChunkSize   uint64
	FilterMask  uint32
}

type ChunkIndex struct {
	NumChunks uint64
	ChunkDims []uint64
	Entries   []ChunkIndexEntry
}

func computeNumChunks(datasetDims, chunkDims []uint64) uint64 {
	result := uint64(1)
	for i, dim := range datasetDims {
		result *= (dim + chunkDims[i] - 1) / chunkDims[i]
	}
	return result
}

func computeLinearIndex(chunkCoords, chunkDims []uint64) uint64 {
	index := uint64(0)
	stride := uint64(1)
	for i := len(chunkCoords) - 1; i >= 0; i-- {
		index += chunkCoords[i] * stride
		stride *= chunkDims[i]
	}
	return index
}

func decodeLinearIndex(linearIndex uint64, chunkDims []uint64) []uint64 {
	coords := make([]uint64, len(chunkDims))
	stride := uint64(1)
	for i := len(chunkDims) - 1; i >= 0; i-- {
		if i < len(chunkDims)-1 {
			stride *= chunkDims[i+1]
		}
		coords[i] = linearIndex / stride
		linearIndex = linearIndex % stride
	}
	return coords
}

func EncodeChunkIndex(index *ChunkIndex) ([]byte, error) {
	rank := uint32(len(index.ChunkDims))

	headerSize := uint64(8 + 4 + 4*rank)
	entrySize := uint64(16)
	totalSize := headerSize + uint64(len(index.Entries))*entrySize

	buf := make([]byte, totalSize)

	binary.LittleEndian.PutUint64(buf[0:8], index.NumChunks)
	binary.LittleEndian.PutUint32(buf[8:12], rank)

	for i, dim := range index.ChunkDims {
		binary.LittleEndian.PutUint32(buf[12+4*i:16+4*i], uint32(dim))
	}

	for i, entry := range index.Entries {
		offset := headerSize + uint64(i)*entrySize
		binary.LittleEndian.PutUint64(buf[offset:offset+8], entry.ChunkOffset)
		binary.LittleEndian.PutUint64(buf[offset+8:offset+16], entry.ChunkSize)
	}

	return buf, nil
}

func DecodeChunkIndex(data []byte) (*ChunkIndex, error) {
	if len(data) < 12 {
		return nil, errors.New("chunk index too short")
	}

	numChunks := binary.LittleEndian.Uint64(data[0:8])
	rank := binary.LittleEndian.Uint32(data[8:12])

	headerSize := uint64(12 + 4*rank)
	if uint64(len(data)) < headerSize {
		return nil, errors.New("chunk index header truncated")
	}

	chunkDims := make([]uint64, rank)
	for i := uint32(0); i < rank; i++ {
		chunkDims[i] = uint64(binary.LittleEndian.Uint32(data[12+4*i : 16+4*i]))
	}

	entrySize := uint64(16)
	expectedSize := headerSize + numChunks*entrySize
	if uint64(len(data)) < expectedSize {
		return nil, errors.New("chunk index truncated")
	}

	entries := make([]ChunkIndexEntry, numChunks)
	for i := uint64(0); i < numChunks; i++ {
		offset := headerSize + i*entrySize
		entries[i] = ChunkIndexEntry{
			ChunkOffset: binary.LittleEndian.Uint64(data[offset : offset+8]),
			ChunkSize:   binary.LittleEndian.Uint64(data[offset+8 : offset+16]),
		}
	}

	return &ChunkIndex{
		NumChunks: numChunks,
		ChunkDims: chunkDims,
		Entries:   entries,
	}, nil
}

func ReadChunkIndex(r io.ReadSeeker, addr uint64) (*ChunkIndex, error) {
	if _, err := r.Seek(int64(addr), 0); err != nil {
		return nil, err
	}

	header := make([]byte, 12)
	if _, err := r.Read(header); err != nil {
		return nil, err
	}

	numChunks := binary.LittleEndian.Uint64(header[0:8])
	rank := binary.LittleEndian.Uint32(header[8:12])

	headerSize := uint64(12 + 4*rank)
	entrySize := uint64(16)
	totalSize := headerSize + numChunks*entrySize

	if _, err := r.Seek(int64(addr), 0); err != nil {
		return nil, err
	}

	data := make([]byte, totalSize)
	if _, err := r.Read(data); err != nil {
		return nil, err
	}

	return DecodeChunkIndex(data)
}

func WriteChunkIndex(w io.WriteSeeker, index *ChunkIndex) (uint64, uint64, error) {
	pos, err := w.Seek(0, 1)
	if err != nil {
		return 0, 0, err
	}

	data, err := EncodeChunkIndex(index)
	if err != nil {
		return 0, 0, err
	}

	if _, err := w.Write(data); err != nil {
		return 0, 0, err
	}

	return uint64(pos), uint64(len(data)), nil
}


