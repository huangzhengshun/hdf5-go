package hdf5

import (
	"bytes"
	"compress/bzip2"
	"compress/flate"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
	"reflect"
	"strings"

	"github.com/huangzhengshun/hdf5-go/format"
	"github.com/klauspost/compress/zstd"
	lzf "github.com/zhuyie/golzf"
)

var (
	ErrClosedDataset = errors.New("hdf5: dataset is closed")
	ErrInvalidData   = errors.New("hdf5: invalid data type")
)

func compressData(data []byte, compression string) ([]byte, error) {
	switch compression {
	case "gzip":
		var buf bytes.Buffer
		w, err := flate.NewWriter(&buf, flate.DefaultCompression)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(data); err != nil {
			w.Close()
			return nil, err
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case "lzf":
		maxSize := len(data) + len(data)/4 + 1
		buf := make([]byte, maxSize)
		n, err := lzf.Compress(data, buf)
		if err != nil {
			return nil, err
		}
		return buf[:n], nil
	case "zstd":
		var buf bytes.Buffer
		w, err := zstd.NewWriter(&buf)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(data); err != nil {
			w.Close()
			return nil, err
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case "bzip2":
		return nil, errors.New("hdf5: bzip2 compression is not supported (standard library only provides decompression)")
	default:
		return data, nil
	}
}

func decompressData(data []byte, compression string, expectedSize int) ([]byte, error) {
	switch compression {
	case "gzip":
		r := flate.NewReader(bytes.NewReader(data))
		defer r.Close()
		buf := make([]byte, expectedSize)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		return buf, nil
	case "lzf":
		buf := make([]byte, expectedSize)
		n, err := lzf.Decompress(data, buf)
		if err != nil {
			return nil, err
		}
		return buf[:n], nil
	case "zstd":
		r, err := zstd.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer r.Close()
		buf := make([]byte, expectedSize)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		return buf, nil
	case "bzip2":
		r := bzip2.NewReader(bytes.NewReader(data))
		buf := make([]byte, expectedSize)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		return buf, nil
	default:
		return data, nil
	}
}

func applyShuffle(data []byte, elemSize int) []byte {
	if elemSize <= 1 {
		return data
	}
	result := make([]byte, len(data))
	numElems := len(data) / elemSize
	for i := 0; i < elemSize; i++ {
		for j := 0; j < numElems; j++ {
			result[i*numElems+j] = data[j*elemSize+i]
		}
	}
	return result
}

func undoShuffle(data []byte, elemSize int) []byte {
	if elemSize <= 1 {
		return data
	}
	result := make([]byte, len(data))
	numElems := len(data) / elemSize
	for i := 0; i < elemSize; i++ {
		for j := 0; j < numElems; j++ {
			result[j*elemSize+i] = data[i*numElems+j]
		}
	}
	return result
}

func computeFletcher32(data []byte) uint32 {
	s1, s2 := uint32(0), uint32(0)
	for i := 0; i < len(data); i += 2 {
		var word uint32
		if i+1 < len(data) {
			word = uint32(binary.LittleEndian.Uint16(data[i : i+2]))
		} else {
			word = uint32(data[i])
		}
		s1 = (s1 + word) % 65535
		s2 = (s2 + s1) % 65535
	}
	return (s2 << 16) | s1
}

func validateFletcher32(data []byte, checksum uint32) bool {
	return computeFletcher32(data) == checksum
}

func applyScaleOffset(data []byte, elemSize int, scaleType int, scaleOffset float64) ([]byte, error) {
	if elemSize != 4 && elemSize != 8 {
		return nil, errors.New("hdf5: ScaleOffset only supports 32-bit and 64-bit floats")
	}

	result := make([]byte, len(data)/elemSize*4)
	numElems := len(data) / elemSize

	for i := 0; i < numElems; i++ {
		var val float64
		if elemSize == 4 {
			val = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[i*4:])))
		} else {
			val = math.Float64frombits(binary.LittleEndian.Uint64(data[i*8:]))
		}

		scaled := int32((val - scaleOffset) * float64(scaleType))
		binary.LittleEndian.PutUint32(result[i*4:], uint32(scaled))
	}

	return result, nil
}

func undoScaleOffset(data []byte, elemSize int, scaleType int, scaleOffset float64) ([]byte, error) {
	if elemSize != 4 && elemSize != 8 {
		return nil, errors.New("hdf5: ScaleOffset only supports 32-bit and 64-bit floats")
	}

	result := make([]byte, len(data)/4*elemSize)
	numElems := len(data) / 4

	for i := 0; i < numElems; i++ {
		scaled := int32(binary.LittleEndian.Uint32(data[i*4:]))
		val := float64(scaled)/float64(scaleType) + scaleOffset

		if elemSize == 4 {
			binary.LittleEndian.PutUint32(result[i*4:], math.Float32bits(float32(val)))
		} else {
			binary.LittleEndian.PutUint64(result[i*8:], math.Float64bits(val))
		}
	}

	return result, nil
}

func applyBitshuffle(data []byte, elemSize int) []byte {
	if elemSize <= 1 || len(data)%elemSize != 0 {
		return data
	}

	result := make([]byte, len(data))
	numElems := len(data) / elemSize
	numBytes := elemSize

	for bytePos := 0; bytePos < numBytes; bytePos++ {
		for bitGroup := 0; bitGroup < 8; bitGroup++ {
			for elemIdx := 0; elemIdx < numElems; elemIdx++ {
				srcByte := data[elemIdx*numBytes+bytePos]
				bit := (srcByte >> bitGroup) & 1
				destByteIdx := (bitGroup*numBytes + bytePos) * numElems / 8
				if destByteIdx >= len(result) {
					continue
				}
				destBitIdx := (bitGroup*numBytes + bytePos) % 8
				if bit != 0 {
					result[destByteIdx] |= 1 << destBitIdx
				}
			}
		}
	}

	return result
}

func undoBitshuffle(data []byte, elemSize int) []byte {
	if elemSize <= 1 || len(data)%elemSize != 0 {
		return data
	}

	result := make([]byte, len(data))
	numElems := len(data) / elemSize
	numBytes := elemSize

	for bytePos := 0; bytePos < numBytes; bytePos++ {
		for bitGroup := 0; bitGroup < 8; bitGroup++ {
			for elemIdx := 0; elemIdx < numElems; elemIdx++ {
				destByteIdx := (bitGroup*numBytes + bytePos) * numElems / 8
				if destByteIdx >= len(data) {
					continue
				}
				destBitIdx := (bitGroup*numBytes + bytePos) % 8
				bit := (data[destByteIdx] >> destBitIdx) & 1
				if bit != 0 {
					result[elemIdx*numBytes+bytePos] |= 1 << bitGroup
				}
			}
		}
	}

	return result
}

type dataset struct {
	name       string
	addr       uint64
	parent     Group
	file       *File
	header     *format.ObjectHeader
	super      *format.Superblock
	dtype      Datatype
	dspace     Dataspace
	layout     format.LayoutMessage
	chunkIndex *format.ChunkIndex
	chunkCache *ChunkCache
	closed     bool
}

func createDataset(f *File, name string, dtype Datatype, space Dataspace, plist PropertyList) (Dataset, error) {
	dtMsg := format.DatatypeMessage{
		Version: format.DatatypeVersion2,
	}

	switch dtype.Class {
	case DatatypeInteger:
		dtMsg.Class = format.ClassInteger
		dtMsg.ByteOrder = format.OrderLittleEndian
		dtMsg.Size = dtype.Size
		dtMsg.Sign = uint8(dtype.Sign)
	case DatatypeFloat:
		dtMsg.Class = format.ClassFloat
		dtMsg.ByteOrder = format.OrderLittleEndian
		dtMsg.Size = dtype.Size
		dtMsg.FloatBits = dtype.FloatBits
	case DatatypeString:
		dtMsg.Class = format.ClassString
		dtMsg.ByteOrder = format.OrderLittleEndian
		dtMsg.Size = dtype.Size
		dtMsg.StringPadding = uint8(dtype.StringPadding)
		dtMsg.StringSize = dtype.StringSize
	case DatatypeCompound:
		dtMsg.Class = format.ClassCompound
		dtMsg.ByteOrder = format.OrderLittleEndian
		if dtype.Size == 0 {
			for _, field := range dtype.Fields {
				fieldEnd := field.Offset + uint32(field.Datatype.Size)
				if fieldEnd > dtype.Size {
					dtype.Size = fieldEnd
				}
			}
		}
		dtMsg.Size = dtype.Size
		for _, field := range dtype.Fields {
			fieldDtMsg := format.DatatypeMessage{
				Version: format.DatatypeVersion2,
			}
			switch field.Datatype.Class {
			case DatatypeInteger:
				fieldDtMsg.Class = format.ClassInteger
				fieldDtMsg.ByteOrder = format.OrderLittleEndian
				fieldDtMsg.Size = field.Datatype.Size
				fieldDtMsg.Sign = uint8(field.Datatype.Sign)
			case DatatypeFloat:
				fieldDtMsg.Class = format.ClassFloat
				fieldDtMsg.ByteOrder = format.OrderLittleEndian
				fieldDtMsg.Size = field.Datatype.Size
				fieldDtMsg.FloatBits = field.Datatype.FloatBits
			case DatatypeString:
				fieldDtMsg.Class = format.ClassString
				fieldDtMsg.ByteOrder = format.OrderLittleEndian
				fieldDtMsg.Size = field.Datatype.Size
				fieldDtMsg.StringPadding = uint8(field.Datatype.StringPadding)
				fieldDtMsg.StringSize = field.Datatype.StringSize
			}
			dtMsg.Fields = append(dtMsg.Fields, format.CompoundField{
				Name:     field.Name,
				Offset:   field.Offset,
				Datatype: fieldDtMsg,
			})
		}
	case DatatypeEnum:
		dtMsg.Class = format.ClassEnum
		dtMsg.ByteOrder = format.OrderLittleEndian
		dtMsg.Size = dtype.Size
		if dtype.BaseType != nil {
			baseMsg := format.DatatypeMessage{
				Version:   format.DatatypeVersion2,
				Class:     format.ClassInteger,
				ByteOrder: format.OrderLittleEndian,
				Size:      dtype.BaseType.Size,
				Sign:      uint8(dtype.BaseType.Sign),
			}
			dtMsg.EnumBase = &baseMsg
		}
		for _, member := range dtype.EnumMembers {
			dtMsg.EnumMembers = append(dtMsg.EnumMembers, format.EnumMember{
				Name:  member.Name,
				Value: member.Value,
			})
		}
	case DatatypeArray:
		dtMsg.Class = format.ClassArray
		dtMsg.ByteOrder = format.OrderLittleEndian
		dtMsg.Size = dtype.Size
		if dtype.BaseType != nil {
			baseMsg := format.DatatypeMessage{
				Version:   format.DatatypeVersion2,
				Class:     uint8(dtype.BaseType.Class),
				ByteOrder: format.OrderLittleEndian,
				Size:      dtype.BaseType.Size,
				Sign:      uint8(dtype.BaseType.Sign),
			}
			dtMsg.ArrayBase = &baseMsg
		}
		dtMsg.ArrayDims = dtype.ArrayDims
	case DatatypeVarLength:
		dtMsg.Class = format.ClassVarLength
		dtMsg.ByteOrder = format.OrderLittleEndian
		dtMsg.Size = dtype.Size
		if dtype.BaseType != nil {
			baseMsg := format.DatatypeMessage{
				Version:   format.DatatypeVersion2,
				Class:     uint8(dtype.BaseType.Class),
				ByteOrder: format.OrderLittleEndian,
				Size:      dtype.BaseType.Size,
				Sign:      uint8(dtype.BaseType.Sign),
			}
			dtMsg.ArrayBase = &baseMsg
		}
	case DatatypeReference:
		dtMsg.Class = format.ClassReference
		dtMsg.ByteOrder = format.OrderLittleEndian
		dtMsg.Size = 8
	default:
		return nil, errors.New("unsupported datatype")
	}

	dtEncoded, err := format.EncodeDatatypeMessage(dtMsg)
	if err != nil {
		return nil, err
	}

	dsEncoded, err := format.EncodeDataspaceMessage(space.Dims, space.MaxDims)
	if err != nil {
		return nil, err
	}

	layoutClass := uint8(format.LayoutContiguous)
	if len(plist.Chunks) > 0 {
		layoutClass = uint8(format.LayoutChunked)
	}

	if dtype.Class == DatatypeCompound && dtype.Size == 0 {
		for _, field := range dtype.Fields {
			fieldEnd := field.Offset + uint32(field.Datatype.Size)
			if fieldEnd > dtype.Size {
				dtype.Size = fieldEnd
			}
		}
	}

	dataSize := space.Size() * uint64(dtype.Size)
	maxCompactSize := uint64(64 * 1024)
	if layoutClass == format.LayoutContiguous && dataSize > 0 && dataSize <= maxCompactSize {
		layoutClass = uint8(format.LayoutCompact)
	}

	layoutEncoded, err := format.EncodeLayoutMessage(layoutClass, dataSize, 0, plist.Chunks)
	if err != nil {
		return nil, err
	}

	msgSize1 := uint16(10 + len(dtEncoded))
	msgSize2 := uint16(10 + len(dsEncoded))
	msgSize3 := uint16(10 + len(layoutEncoded))

	messages := []format.ObjectHeaderMessage{
		{
			Type:    format.MessageDatatype,
			Size:    msgSize1,
			Flags:   0,
			Content: dtEncoded,
		},
		{
			Type:    format.MessageDataspace,
			Size:    msgSize2,
			Flags:   0,
			Content: dsEncoded,
		},
		{
			Type:    format.MessageDataset,
			Size:    msgSize3,
			Flags:   0,
			Content: layoutEncoded,
		},
	}

	numMessages := 3
	var filterEncoded []byte
	if plist.Compression != "" || plist.Shuffle || plist.Fletcher32 {
		var filters []format.FilterInfo

		if plist.Shuffle {
			filters = append(filters, format.FilterInfo{
				FilterID:       format.FilterShuffle,
				Flags:          0,
				NumberOfParams: 1,
				Params:         []uint32{dtype.Size},
			})
		}

		if plist.Compression == "gzip" {
			level := uint32(5)
			if plist.CompressionLevel > 0 {
				level = uint32(plist.CompressionLevel)
			}
			filters = append(filters, format.FilterInfo{
				FilterID:       format.FilterDeflate,
				Flags:          0,
				NumberOfParams: 1,
				Params:         []uint32{level},
			})
		} else if plist.Compression == "lzf" {
			filters = append(filters, format.FilterInfo{
				FilterID:       format.FilterLZF,
				Flags:          0,
				NumberOfParams: 0,
			})
		} else if plist.Compression == "zstd" {
			level := uint32(3)
			if plist.CompressionLevel > 0 {
				level = uint32(plist.CompressionLevel)
			}
			filters = append(filters, format.FilterInfo{
				FilterID:       format.FilterZSTD,
				Flags:          0,
				NumberOfParams: 1,
				Params:         []uint32{level},
			})
		}

		if plist.Fletcher32 {
			filters = append(filters, format.FilterInfo{
				FilterID:       format.FilterFletcher32,
				Flags:          0,
				NumberOfParams: 0,
			})
		}

		filterEncoded, err = format.EncodeFilterPipelineMessage(filters)
		if err != nil {
			return nil, err
		}

		msgSize4 := uint16(10 + len(filterEncoded))
		messages = append(messages, format.ObjectHeaderMessage{
			Type:    format.MessageFilterPipeline,
			Size:    msgSize4,
			Flags:   0,
			Content: filterEncoded,
		})
		numMessages++
	}

	totalSize := uint64(4) + uint64(1) + uint64(1) + uint64(f.super.LengthSize) + uint64(2) + uint64(f.super.OffsetSize)*uint64(numMessages)
	for _, msg := range messages {
		totalSize += uint64(msg.Size)
	}

	alignedTotalSize := ((totalSize + 1023) / 1024) * 1024

	oh := &format.ObjectHeader{
		Version:          format.ObjectHeaderVersion2,
		Flags:            0,
		TotalHeaderSize:  alignedTotalSize,
		NumberOfMessages: uint16(numMessages),
		Messages:         messages,
	}

	f.writer.Seek(0, 2)

	currentPos, _ := f.writer.Seek(0, 1)

	alignedPos := ((currentPos + 511) / 512) * 512
	if alignedPos < currentPos {
		alignedPos += 512
	}

	if _, err := f.writer.Seek(alignedPos, 0); err != nil {
		return nil, err
	}

	addr, err := format.WriteObjectHeader(f.writer, f.super, oh)
	if err != nil {
		return nil, err
	}

	dataAddr64, _ := f.writer.Seek(0, 1)
	dataAddr := uint64(dataAddr64)

	layoutMsg, _ := format.DecodeLayoutMessage(layoutEncoded)

	var chunkIndex *format.ChunkIndex
	if layoutMsg.LayoutClass == format.LayoutChunked {
		chunkIndex = &format.ChunkIndex{
			NumChunks: 0,
			ChunkDims: plist.Chunks,
			Entries:   make([]format.ChunkIndexEntry, 0),
		}
		chunkIndexAddr, chunkIndexSize, _ := format.WriteChunkIndex(f.writer, chunkIndex)
		layoutMsg.DataAddress = chunkIndexAddr
		dataAddr = chunkIndexAddr + chunkIndexSize
	} else {
		layoutMsg.DataAddress = dataAddr
	}

	newLayoutEncoded, _ := format.EncodeLayoutMessage(layoutMsg.LayoutClass, layoutMsg.DataSize, layoutMsg.DataAddress, layoutMsg.ChunkDims)
	oh.Messages[2].Content = newLayoutEncoded
	oh.Messages[2].Size = uint16(10 + len(newLayoutEncoded))

	f.writer.Seek(int64(addr), 0)
	format.WriteObjectHeader(f.writer, f.super, oh)

	f.writer.Seek(int64(dataAddr), 0)

	dtype.Compression = plist.Compression
	dtype.Shuffle = plist.Shuffle
	dtype.Fletcher32 = plist.Fletcher32

	dset := &dataset{
		name:       name,
		addr:       addr,
		file:       f,
		header:     oh,
		super:      f.super,
		dtype:      dtype,
		dspace:     space,
		layout:     layoutMsg,
		chunkIndex: chunkIndex,
	}

	return dset, nil
}

func (d *dataset) Name() string {
	return d.name
}

func (d *dataset) Parent() Group {
	return d.parent
}

func (d *dataset) File() *File {
	return d.file
}

func (d *dataset) Datatype() Datatype {
	return d.dtype
}

func (d *dataset) Dataspace() Dataspace {
	return d.dspace
}

func (d *dataset) Read(data interface{}) error {
	if d.closed {
		return ErrClosedDataset
	}

	if d.layout.LayoutClass == format.LayoutChunked {
		return d.readChunked(data)
	}

	dataSize := d.dspace.Size() * uint64(d.dtype.Size)
	if d.dtype.Class == DatatypeVarLength && d.layout.DataSize > 0 {
		dataSize = d.layout.DataSize
	}

	if _, err := d.file.reader.Seek(int64(d.layout.DataAddress), 0); err != nil {
		return err
	}

	var buf []byte
	if d.dtype.Compression != "" {
		compressedBuf := make([]byte, d.layout.DataSize)
		if _, err := d.file.reader.Read(compressedBuf); err != nil {
			return err
		}

		var storedChecksum uint32
		if d.dtype.Fletcher32 && len(compressedBuf) >= 4 {
			storedChecksum = binary.LittleEndian.Uint32(compressedBuf[len(compressedBuf)-4:])
			compressedBuf = compressedBuf[:len(compressedBuf)-4]
		}

		var err error
		buf, err = decompressData(compressedBuf, d.dtype.Compression, int(dataSize))
		if err != nil {
			return err
		}

		if d.dtype.Fletcher32 {
			computedChecksum := computeFletcher32(buf)
			if computedChecksum != storedChecksum {
				return errors.New("hdf5: fletcher32 checksum mismatch")
			}
		}

		if d.dtype.Shuffle {
			buf = undoShuffle(buf, int(d.dtype.Size))
		}
	} else {
		buf = make([]byte, dataSize)
		if _, err := d.file.reader.Read(buf); err != nil {
			return err
		}
	}

	return decodeData(buf, data, d.dtype)
}

func (d *dataset) readChunked(data interface{}) error {
	dataSize := d.dspace.Size() * uint64(d.dtype.Size)
	buf := make([]byte, dataSize)

	if d.chunkIndex == nil {
		var err error
		d.chunkIndex, err = format.ReadChunkIndex(d.file.reader, d.layout.DataAddress)
		if err != nil {
			return err
		}
	}

	if d.chunkCache == nil {
		d.chunkCache = globalFileCache.GetOrCreate(d.addr, 0)
	}

	if d.chunkIndex.NumChunks == 0 {
		return decodeData(buf, data, d.dtype)
	}

	var chunkSize uint64 = 1
	for _, dim := range d.layout.ChunkDims {
		chunkSize *= dim
	}
	chunkByteSize := chunkSize * uint64(d.dtype.Size)

	for i, entry := range d.chunkIndex.Entries {
		if entry.ChunkOffset == 0 {
			continue
		}

		var chunkBuf []byte
		var err error

		if cachedData, ok := d.chunkCache.Get(uint64(i)); ok {
			chunkBuf = cachedData
		} else {
			chunkData := make([]byte, entry.ChunkSize)
			if _, err := d.file.reader.Seek(int64(entry.ChunkOffset), 0); err != nil {
				return err
			}
			if _, err := d.file.reader.Read(chunkData); err != nil {
				return err
			}

			var storedChecksum uint32
			if d.dtype.Fletcher32 && len(chunkData) >= 4 {
				storedChecksum = binary.LittleEndian.Uint32(chunkData[len(chunkData)-4:])
				chunkData = chunkData[:len(chunkData)-4]
			}

			remainingDataSize := dataSize - uint64(i)*chunkByteSize
			expectedSize := int(chunkByteSize)
			if remainingDataSize < chunkByteSize {
				expectedSize = int(remainingDataSize)
			}
			chunkBuf, err = decompressData(chunkData, d.dtype.Compression, expectedSize)
			if err != nil {
				return err
			}

			if d.dtype.Fletcher32 {
				computedChecksum := computeFletcher32(chunkBuf)
				if computedChecksum != storedChecksum {
					return errors.New("hdf5: fletcher32 checksum mismatch")
				}
			}

			if d.dtype.Shuffle {
				chunkBuf = undoShuffle(chunkBuf, int(d.dtype.Size))
			}

			d.chunkCache.Set(uint64(i), chunkBuf)
		}

		chunkOffset := uint64(i) * chunkSize * uint64(d.dtype.Size)
		if chunkOffset < dataSize {
			copySize := uint64(len(chunkBuf))
			if chunkOffset+copySize > dataSize {
				copySize = dataSize - chunkOffset
			}
			copy(buf[chunkOffset:chunkOffset+copySize], chunkBuf[:copySize])
		}
	}

	return decodeData(buf, data, d.dtype)
}

func (d *dataset) Write(data interface{}) error {
	if d.closed {
		return ErrClosedDataset
	}

	if d.layout.LayoutClass == format.LayoutChunked {
		return d.writeChunked(data)
	}

	buf, err := encodeData(data, d.dtype, d.dspace)
	if err != nil {
		return err
	}

	if d.dtype.Shuffle {
		buf = applyShuffle(buf, int(d.dtype.Size))
	}

	writeBuf, err := compressData(buf, d.dtype.Compression)
	if err != nil {
		return err
	}

	if d.dtype.Fletcher32 {
		checksum := computeFletcher32(buf)
		checksumBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(checksumBuf, checksum)
		writeBuf = append(writeBuf, checksumBuf...)
	}

	newDataSize := uint64(len(writeBuf))

	if d.layout.DataSize == 0 {
		if _, err := d.file.writer.Seek(int64(d.layout.DataAddress), 0); err != nil {
			return err
		}
		if _, err := d.file.writer.Write(writeBuf); err != nil {
			return err
		}
		d.layout.DataSize = newDataSize
		if _, err := d.file.writer.Seek(0, 2); err != nil {
			return err
		}
		return nil
	}

	layoutEncoded, err := format.EncodeLayoutMessage(d.layout.LayoutClass, newDataSize, d.layout.DataAddress, d.layout.ChunkDims)
	if err != nil {
		return err
	}

	var oldLayoutSize uint16
	for _, msg := range d.header.Messages {
		if msg.Type == format.MessageDataset {
			oldLayoutSize = msg.Size
			break
		}
	}

	newLayoutSize := uint16(10 + len(layoutEncoded))
	sizeDiff := int64(newLayoutSize) - int64(oldLayoutSize)

	if sizeDiff > 0 && d.layout.DataSize > 0 {
		currentDataAddr := d.layout.DataAddress
		existingData := make([]byte, d.layout.DataSize)
		if _, err := d.file.reader.Seek(int64(currentDataAddr), 0); err != nil {
			return err
		}
		if _, err := d.file.reader.Read(existingData); err != nil {
			return err
		}

		newDataAddr := currentDataAddr + uint64(sizeDiff)
		d.layout.DataAddress = newDataAddr

		if _, err := d.file.writer.Seek(int64(newDataAddr), 0); err != nil {
			return err
		}
		if _, err := d.file.writer.Write(existingData); err != nil {
			return err
		}

		layoutEncoded, err = format.EncodeLayoutMessage(d.layout.LayoutClass, newDataSize, newDataAddr, d.layout.ChunkDims)
		if err != nil {
			return err
		}
	}

	for i, msg := range d.header.Messages {
		if msg.Type == format.MessageDataset {
			d.header.Messages[i].Content = layoutEncoded
			d.header.Messages[i].Size = newLayoutSize
			d.header.TotalHeaderSize += uint64(sizeDiff)
			break
		}
	}

	if _, err := d.file.writer.Seek(int64(d.addr), 0); err != nil {
		return err
	}
	if _, err := format.WriteObjectHeader(d.file.writer, d.super, d.header); err != nil {
		return err
	}

	if _, err := d.file.writer.Seek(int64(d.layout.DataAddress), 0); err != nil {
		return err
	}
	if _, err := d.file.writer.Write(writeBuf); err != nil {
		return err
	}

	d.layout.DataSize = newDataSize

	return nil
}

func (d *dataset) writeChunked(data interface{}) error {
	buf, err := encodeData(data, d.dtype, d.dspace)
	if err != nil {
		return err
	}

	var chunkSize uint64 = 1
	for _, dim := range d.layout.ChunkDims {
		chunkSize *= dim
	}
	chunkByteSize := chunkSize * uint64(d.dtype.Size)
	numChunks := (uint64(len(buf)) + chunkByteSize - 1) / chunkByteSize

	if d.chunkIndex == nil {
		d.chunkIndex = &format.ChunkIndex{
			NumChunks: numChunks,
			ChunkDims: d.layout.ChunkDims,
			Entries:   make([]format.ChunkIndexEntry, numChunks),
		}
	} else if d.chunkIndex.NumChunks < numChunks {
		newEntries := make([]format.ChunkIndexEntry, numChunks)
		copy(newEntries, d.chunkIndex.Entries)
		d.chunkIndex.Entries = newEntries
		d.chunkIndex.NumChunks = numChunks
		d.chunkIndex.ChunkDims = d.layout.ChunkDims
	}

	for i := uint64(0); i < numChunks; i++ {
		start := i * chunkByteSize
		end := start + chunkByteSize
		if end > uint64(len(buf)) {
			end = uint64(len(buf))
		}
		chunkBuf := buf[start:end]

		if d.dtype.Shuffle {
			chunkBuf = applyShuffle(chunkBuf, int(d.dtype.Size))
		}

		writeBuf, err := compressData(chunkBuf, d.dtype.Compression)
		if err != nil {
			return err
		}

		if d.dtype.Fletcher32 {
			checksum := computeFletcher32(chunkBuf)
			checksumBuf := make([]byte, 4)
			binary.LittleEndian.PutUint32(checksumBuf, checksum)
			writeBuf = append(writeBuf, checksumBuf...)
		}

		if d.chunkIndex.Entries[i].ChunkOffset == 0 {
			pos, _ := d.file.writer.Seek(0, 2)
			d.chunkIndex.Entries[i].ChunkOffset = uint64(pos)
		}

		if _, err := d.file.writer.Seek(int64(d.chunkIndex.Entries[i].ChunkOffset), 0); err != nil {
			return err
		}
		if _, err := d.file.writer.Write(writeBuf); err != nil {
			return err
		}

		d.chunkIndex.Entries[i].ChunkSize = uint64(len(writeBuf))

		if d.chunkCache != nil {
			d.chunkCache.Remove(uint64(i))
		}
	}

	indexData, err := format.EncodeChunkIndex(d.chunkIndex)
	if err != nil {
		return err
	}

	if d.chunkIndex.NumChunks > 0 {
		pos, _ := d.file.writer.Seek(0, 2)
		d.layout.DataAddress = uint64(pos)
	}

	d.layout.DataSize = uint64(len(indexData))

	if _, err := d.file.writer.Seek(int64(d.layout.DataAddress), 0); err != nil {
		return err
	}
	if _, _, err := format.WriteChunkIndex(d.file.writer, d.chunkIndex); err != nil {
		return err
	}

	layoutEncoded, err := format.EncodeLayoutMessage(d.layout.LayoutClass, d.layout.DataSize, d.layout.DataAddress, d.layout.ChunkDims)
	if err != nil {
		return err
	}

	newLayoutSize := uint16(10 + len(layoutEncoded))

	for i, msg := range d.header.Messages {
		if msg.Type == format.MessageDataset {
			sizeDiff := int64(newLayoutSize) - int64(d.header.Messages[i].Size)
			d.header.Messages[i].Content = layoutEncoded
			d.header.Messages[i].Size = newLayoutSize
			d.header.TotalHeaderSize += uint64(sizeDiff)
			break
		}
	}

	if _, err := d.file.writer.Seek(int64(d.addr), 0); err != nil {
		return err
	}
	if _, err := format.WriteObjectHeader(d.file.writer, d.super, d.header); err != nil {
		return err
	}

	if file, ok := d.file.writer.(*os.File); ok {
		if err := file.Sync(); err != nil {
			return err
		}
	}

	return nil
}

func (d *dataset) ReadSlice(data interface{}, sel Selection) error {
	if d.closed {
		return ErrClosedDataset
	}

	if len(sel.Points) > 0 {
		return d.readPoints(data, sel.Points)
	}

	if len(sel.Mask) > 0 {
		return d.readMask(data, sel.Mask)
	}

	if len(sel.Hyperslabs) == 0 {
		return errors.New("hdf5: selection is empty")
	}

	rank := len(d.dspace.Dims)

	totalSize := uint64(0)
	for _, h := range sel.Hyperslabs {
		if len(h.Start) != rank || len(h.Count) != rank {
			return errors.New("hdf5: hyperslab dimensions mismatch")
		}
		slabSize := uint64(1)
		for i := range h.Count {
			slabSize *= h.Count[i]
		}
		totalSize += slabSize
	}

	dataSize := totalSize * uint64(d.dtype.Size)
	buf := make([]byte, dataSize)

	if d.layout.LayoutClass == format.LayoutChunked {
		fullBuf := make([]byte, d.dspace.Size()*uint64(d.dtype.Size))
		if err := d.readChunked(&fullBuf); err != nil {
			return err
		}
		readOffset := uint64(0)
		for _, h := range sel.Hyperslabs {
			bytesRead, err := d.readSingleHyperslab(buf[readOffset:], fullBuf, h)
			if err != nil {
				return err
			}
			readOffset += bytesRead
		}
		return decodeData(buf, data, d.dtype)
	}

	if _, err := d.file.reader.Seek(int64(d.layout.DataAddress), 0); err != nil {
		return err
	}

	fullBuf := make([]byte, d.dspace.Size()*uint64(d.dtype.Size))
	if d.dtype.Compression != "" {
		compressedBuf := make([]byte, d.layout.DataSize)
		if _, err := d.file.reader.Read(compressedBuf); err != nil {
			return err
		}
		var err error
		fullBuf, err = decompressData(compressedBuf, d.dtype.Compression, int(d.dspace.Size()*uint64(d.dtype.Size)))
		if err != nil {
			return err
		}
	} else {
		if _, err := d.file.reader.Read(fullBuf); err != nil {
			return err
		}
	}

	readOffset := uint64(0)
	for _, h := range sel.Hyperslabs {
		bytesRead, err := d.readSingleHyperslab(buf[readOffset:], fullBuf, h)
		if err != nil {
			return err
		}
		readOffset += bytesRead
	}
	return decodeData(buf, data, d.dtype)
}

func (d *dataset) readSingleHyperslab(buf, fullBuf []byte, h Hyperslab) (uint64, error) {
	rank := len(d.dspace.Dims)

	readOffset := uint64(0)
	indices := make([]uint64, rank)
	copy(indices, h.Start)

	if h.Stride == nil {
		h.Stride = make([]uint64, rank)
		for i := range h.Stride {
			h.Stride[i] = 1
		}
	}

	for {
		fullIndex := uint64(0)
		stride := uint64(1)
		for i := rank - 1; i >= 0; i-- {
			fullIndex += indices[i] * stride
			stride *= d.dspace.Dims[i]
		}

		srcOffset := fullIndex * uint64(d.dtype.Size)
		copy(buf[readOffset:readOffset+uint64(d.dtype.Size)], fullBuf[srcOffset:srcOffset+uint64(d.dtype.Size)])
		readOffset += uint64(d.dtype.Size)

		for i := rank - 1; i >= 0; i-- {
			indices[i] += h.Stride[i]
			if indices[i] < h.Start[i]+h.Count[i] {
				break
			}
			if i == 0 {
				return readOffset, nil
			}
			indices[i] = h.Start[i]
		}
	}
}

func (d *dataset) readPoints(data interface{}, points [][]uint64) error {
	rank := len(d.dspace.Dims)
	dataSize := uint64(len(points)) * uint64(d.dtype.Size)
	buf := make([]byte, dataSize)

	fullBuf := make([]byte, d.dspace.Size()*uint64(d.dtype.Size))
	if d.layout.LayoutClass == format.LayoutChunked {
		if err := d.readChunked(&fullBuf); err != nil {
			return err
		}
	} else {
		if _, err := d.file.reader.Seek(int64(d.layout.DataAddress), 0); err != nil {
			return err
		}
		if d.dtype.Compression != "" {
			compressedBuf := make([]byte, d.layout.DataSize)
			if _, err := d.file.reader.Read(compressedBuf); err != nil {
				return err
			}
			var decompressErr error
			fullBuf, decompressErr = decompressData(compressedBuf, d.dtype.Compression, int(len(fullBuf)))
			if decompressErr != nil {
				return decompressErr
			}
		} else {
			if _, err := d.file.reader.Read(fullBuf); err != nil {
				return err
			}
		}
	}

	readOffset := uint64(0)
	for _, point := range points {
		if len(point) != rank {
			return errors.New("hdf5: point dimensions mismatch")
		}

		fullIndex := uint64(0)
		stride := uint64(1)
		for i := rank - 1; i >= 0; i-- {
			fullIndex += point[i] * stride
			stride *= d.dspace.Dims[i]
		}

		srcOffset := fullIndex * uint64(d.dtype.Size)
		copy(buf[readOffset:readOffset+uint64(d.dtype.Size)], fullBuf[srcOffset:srcOffset+uint64(d.dtype.Size)])
		readOffset += uint64(d.dtype.Size)
	}

	return decodeData(buf, data, d.dtype)
}

func (d *dataset) readMask(data interface{}, mask []bool) error {
	dataSize := d.dspace.Size()
	if uint64(len(mask)) != dataSize {
		return errors.New("hdf5: mask length mismatch")
	}

	selectedCount := 0
	for _, v := range mask {
		if v {
			selectedCount++
		}
	}

	bufSize := uint64(selectedCount) * uint64(d.dtype.Size)
	buf := make([]byte, bufSize)

	fullBuf := make([]byte, dataSize*uint64(d.dtype.Size))
	if d.layout.LayoutClass == format.LayoutChunked {
		if err := d.readChunked(&fullBuf); err != nil {
			return err
		}
	} else {
		if _, err := d.file.reader.Seek(int64(d.layout.DataAddress), 0); err != nil {
			return err
		}
		if d.dtype.Compression != "" {
			compressedBuf := make([]byte, d.layout.DataSize)
			if _, err := d.file.reader.Read(compressedBuf); err != nil {
				return err
			}
			var decompressErr error
			fullBuf, decompressErr = decompressData(compressedBuf, d.dtype.Compression, int(len(fullBuf)))
			if decompressErr != nil {
				return decompressErr
			}
		} else {
			if _, err := d.file.reader.Read(fullBuf); err != nil {
				return err
			}
		}
	}

	readOffset := uint64(0)
	for i := uint64(0); i < dataSize; i++ {
		if mask[i] {
			srcOffset := i * uint64(d.dtype.Size)
			copy(buf[readOffset:readOffset+uint64(d.dtype.Size)], fullBuf[srcOffset:srcOffset+uint64(d.dtype.Size)])
			readOffset += uint64(d.dtype.Size)
		}
	}

	return decodeData(buf, data, d.dtype)
}

func (d *dataset) CreateReference(sel Selection) (RegionReference, error) {
	if d.closed {
		return RegionReference{}, ErrClosedDataset
	}

	return RegionReference{
		DatasetAddr: d.addr,
		Selection:   sel,
	}, nil
}

func (d *dataset) CreateObjectReference() (ObjectReference, error) {
	if d.closed {
		return ObjectReference{}, ErrClosedDataset
	}

	return ObjectReference{
		ObjectAddr: d.addr,
		ObjectType: ObjectTypeDataset,
	}, nil
}

func parseDatasetFromObjectHeader(oh *format.ObjectHeader, file *File, addr uint64) (*dataset, error) {
	dset := &dataset{
		name:   "",
		addr:   addr,
		file:   file,
		super:  file.super,
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
				dset.dtype.Size = dtMsg.Size
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
				chunkIndex, _ := format.ReadChunkIndex(file.reader, layoutMsg.DataAddress)
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

func (d *dataset) WriteSlice(data interface{}, sel Selection) error {
	if d.closed {
		return ErrClosedDataset
	}

	if len(sel.Hyperslabs) == 0 {
		return errors.New("hdf5: selection is empty")
	}

	rank := len(d.dspace.Dims)

	totalSize := uint64(0)
	for _, h := range sel.Hyperslabs {
		if len(h.Start) != rank || len(h.Count) != rank {
			return errors.New("hdf5: hyperslab dimensions mismatch")
		}
		slabSize := uint64(1)
		for i := range h.Count {
			slabSize *= h.Count[i]
		}
		totalSize += slabSize
	}

	selSpace := Dataspace{Dims: []uint64{totalSize}}
	encoded, err := encodeData(data, d.dtype, selSpace)
	if err != nil {
		return err
	}

	dataSize := d.dspace.Size() * uint64(d.dtype.Size)
	fullBuf := make([]byte, dataSize)

	if d.layout.LayoutClass == format.LayoutChunked {
		if d.chunkIndex == nil {
			var readErr error
			d.chunkIndex, readErr = format.ReadChunkIndex(d.file.reader, d.layout.DataAddress)
			if readErr != nil {
				return readErr
			}
		}

		var chunkSize uint64 = 1
		for _, dim := range d.layout.ChunkDims {
			chunkSize *= dim
		}
		chunkByteSize := chunkSize * uint64(d.dtype.Size)

		for i, entry := range d.chunkIndex.Entries {
			if entry.ChunkOffset == 0 {
				continue
			}

			if _, seekErr := d.file.reader.Seek(int64(entry.ChunkOffset), 0); seekErr != nil {
				return seekErr
			}

			chunkData := make([]byte, entry.ChunkSize)
			if _, readErr := d.file.reader.Read(chunkData); readErr != nil {
				return readErr
			}

			var storedChecksum uint32
			if d.dtype.Fletcher32 && len(chunkData) >= 4 {
				storedChecksum = binary.LittleEndian.Uint32(chunkData[len(chunkData)-4:])
				chunkData = chunkData[:len(chunkData)-4]
			}

			remainingDataSize := dataSize - uint64(i)*chunkByteSize
			expectedSize := int(chunkByteSize)
			if remainingDataSize < chunkByteSize {
				expectedSize = int(remainingDataSize)
			}

			chunkBuf, decompressErr := decompressData(chunkData, d.dtype.Compression, expectedSize)
			if decompressErr != nil {
				return decompressErr
			}

			if d.dtype.Fletcher32 {
				computedChecksum := computeFletcher32(chunkBuf)
				if computedChecksum != storedChecksum {
					return errors.New("hdf5: fletcher32 checksum mismatch")
				}
			}

			if d.dtype.Shuffle {
				chunkBuf = undoShuffle(chunkBuf, int(d.dtype.Size))
			}

			chunkOffset := uint64(i) * chunkByteSize
			if chunkOffset < dataSize {
				copySize := uint64(len(chunkBuf))
				if chunkOffset+copySize > dataSize {
					copySize = dataSize - chunkOffset
				}
				copy(fullBuf[chunkOffset:chunkOffset+copySize], chunkBuf[:copySize])
			}
		}
	} else {
		if _, err := d.file.reader.Seek(int64(d.layout.DataAddress), 0); err != nil {
			return err
		}
		if d.dtype.Compression != "" {
			compressedBuf := make([]byte, d.layout.DataSize)
			if _, err := d.file.reader.Read(compressedBuf); err != nil {
				return err
			}
			decompressed, err := decompressData(compressedBuf, d.dtype.Compression, int(dataSize))
			if err != nil {
				return err
			}
			copy(fullBuf, decompressed)
		} else {
			if _, err := d.file.reader.Read(fullBuf); err != nil && err != io.EOF {
				return err
			}
		}
	}

	writeOffset := uint64(0)
	for _, h := range sel.Hyperslabs {
		bytesWritten, err := d.writeSingleHyperslab(fullBuf, encoded[writeOffset:], h, rank, d.dtype, d.dspace.Dims)
		if err != nil {
			return err
		}
		writeOffset += bytesWritten
	}

	if d.layout.LayoutClass == format.LayoutChunked {
		var chunkSize uint64 = 1
		for _, dim := range d.layout.ChunkDims {
			chunkSize *= dim
		}
		chunkByteSize := chunkSize * uint64(d.dtype.Size)
		numChunks := (uint64(len(fullBuf)) + chunkByteSize - 1) / chunkByteSize

		if d.chunkIndex == nil {
			d.chunkIndex = &format.ChunkIndex{
				NumChunks: numChunks,
				ChunkDims: d.layout.ChunkDims,
				Entries:   make([]format.ChunkIndexEntry, numChunks),
			}
		} else if d.chunkIndex.NumChunks < numChunks {
			newEntries := make([]format.ChunkIndexEntry, numChunks)
			copy(newEntries, d.chunkIndex.Entries)
			d.chunkIndex.Entries = newEntries
			d.chunkIndex.NumChunks = numChunks
			d.chunkIndex.ChunkDims = d.layout.ChunkDims
		}

		for i := uint64(0); i < numChunks; i++ {
			start := i * chunkByteSize
			end := start + chunkByteSize
			if end > uint64(len(fullBuf)) {
				end = uint64(len(fullBuf))
			}
			chunkBuf := fullBuf[start:end]

			if d.dtype.Shuffle {
				chunkBuf = applyShuffle(chunkBuf, int(d.dtype.Size))
			}

			writeBuf, err := compressData(chunkBuf, d.dtype.Compression)
			if err != nil {
				return err
			}

			if d.dtype.Fletcher32 {
				checksum := computeFletcher32(chunkBuf)
				checksumBuf := make([]byte, 4)
				binary.LittleEndian.PutUint32(checksumBuf, checksum)
				writeBuf = append(writeBuf, checksumBuf...)
			}

			if d.chunkIndex.Entries[i].ChunkOffset == 0 {
				pos, _ := d.file.writer.Seek(0, 2)
				d.chunkIndex.Entries[i].ChunkOffset = uint64(pos)
			}

			if _, err := d.file.writer.Seek(int64(d.chunkIndex.Entries[i].ChunkOffset), 0); err != nil {
				return err
			}
			if _, err := d.file.writer.Write(writeBuf); err != nil {
				return err
			}

			d.chunkIndex.Entries[i].ChunkSize = uint64(len(writeBuf))

			if d.chunkCache != nil {
				d.chunkCache.Remove(uint64(i))
			}
		}

		if d.chunkIndex.NumChunks > 0 {
			pos, _ := d.file.writer.Seek(0, 2)
			d.layout.DataAddress = uint64(pos)
		}

		indexData, err := format.EncodeChunkIndex(d.chunkIndex)
		if err != nil {
			return err
		}
		d.layout.DataSize = uint64(len(indexData))

		if _, err := d.file.writer.Seek(int64(d.layout.DataAddress), 0); err != nil {
			return err
		}
		if _, _, err := format.WriteChunkIndex(d.file.writer, d.chunkIndex); err != nil {
			return err
		}

		layoutEncoded, err := format.EncodeLayoutMessage(d.layout.LayoutClass, d.layout.DataSize, d.layout.DataAddress, d.layout.ChunkDims)
		if err != nil {
			return err
		}

		newLayoutSize := uint16(10 + len(layoutEncoded))

		for i, msg := range d.header.Messages {
			if msg.Type == format.MessageDataset {
				sizeDiff := int64(newLayoutSize) - int64(d.header.Messages[i].Size)
				d.header.Messages[i].Content = layoutEncoded
				d.header.Messages[i].Size = newLayoutSize
				d.header.TotalHeaderSize += uint64(sizeDiff)
				break
			}
		}

		if _, err := d.file.writer.Seek(int64(d.addr), 0); err != nil {
			return err
		}
		if _, err := format.WriteObjectHeader(d.file.writer, d.super, d.header); err != nil {
			return err
		}

		if file, ok := d.file.writer.(*os.File); ok {
			if err := file.Sync(); err != nil {
				return err
			}
		}

		return nil
	}

	writeBuf := fullBuf
	if d.dtype.Shuffle {
		writeBuf = applyShuffle(writeBuf, int(d.dtype.Size))
	}

	compressed, err := compressData(writeBuf, d.dtype.Compression)
	if err != nil {
		return err
	}

	if d.dtype.Fletcher32 {
		checksum := computeFletcher32(writeBuf)
		checksumBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(checksumBuf, checksum)
		compressed = append(compressed, checksumBuf...)
	}

	if _, err := d.file.writer.Seek(int64(d.layout.DataAddress), 0); err != nil {
		return err
	}
	if _, err := d.file.writer.Write(compressed); err != nil {
		return err
	}

	d.layout.DataSize = uint64(len(compressed))
	return nil
}

func (d *dataset) writeSingleHyperslab(fullBuf, encoded []byte, h Hyperslab, rank int, dtype Datatype, dims []uint64) (uint64, error) {
	writeOffset := uint64(0)
	indices := make([]uint64, rank)
	copy(indices, h.Start)

	if h.Stride == nil {
		h.Stride = make([]uint64, rank)
		for i := range h.Stride {
			h.Stride[i] = 1
		}
	}

	for {
		fullIndex := uint64(0)
		stride := uint64(1)
		for i := rank - 1; i >= 0; i-- {
			fullIndex += indices[i] * stride
			stride *= dims[i]
		}

		dstOffset := fullIndex * uint64(dtype.Size)
		copy(fullBuf[dstOffset:dstOffset+uint64(dtype.Size)], encoded[writeOffset:writeOffset+uint64(dtype.Size)])
		writeOffset += uint64(dtype.Size)

		for i := rank - 1; i >= 0; i-- {
			indices[i] += h.Stride[i]
			if indices[i] < h.Start[i]+h.Count[i] {
				break
			}
			if i == 0 {
				return writeOffset, nil
			}
			indices[i] = h.Start[i]
		}
	}
}

func (d *dataset) Attributes() *AttributeManager {
	return &AttributeManager{obj: d}
}

func (d *dataset) Resize(newDims []uint64) error {
	if d.closed {
		return ErrClosedDataset
	}

	if len(newDims) != len(d.dspace.Dims) {
		return errors.New("hdf5: rank mismatch in resize")
	}

	for i, newDim := range newDims {
		if i < len(d.dspace.MaxDims) && d.dspace.MaxDims[i] != 0 && newDim > d.dspace.MaxDims[i] {
			return errors.New("hdf5: new dimension exceeds max dimension")
		}
	}

	oldSize := d.dspace.Size() * uint64(d.dtype.Size)
	d.dspace.Dims = newDims
	newSize := d.dspace.Size() * uint64(d.dtype.Size)

	if d.layout.LayoutClass == format.LayoutChunked {
		chunkByteSize := uint64(d.layout.ChunkSize) * uint64(d.dtype.Size)
		numChunks := (newSize + chunkByteSize - 1) / chunkByteSize

		if d.chunkIndex != nil && d.chunkIndex.NumChunks < numChunks {
			newEntries := make([]format.ChunkIndexEntry, numChunks)
			copy(newEntries, d.chunkIndex.Entries)
			d.chunkIndex.Entries = newEntries
			d.chunkIndex.NumChunks = numChunks
		}

		d.layout.DataSize = newSize

		if d.chunkIndex != nil && d.chunkIndex.NumChunks > 0 {
			indexData, _ := format.EncodeChunkIndex(d.chunkIndex)
			pos, _ := d.file.writer.Seek(0, 2)
			d.layout.DataAddress = uint64(pos)
			d.layout.DataSize = uint64(len(indexData))

			if _, err := d.file.writer.Seek(int64(d.layout.DataAddress), 0); err != nil {
				return err
			}
			if _, _, err := format.WriteChunkIndex(d.file.writer, d.chunkIndex); err != nil {
				return err
			}
		}
	} else {
		if newSize > oldSize {
			additionalData := make([]byte, newSize-oldSize)
			if _, err := d.file.writer.Seek(int64(d.layout.DataAddress+oldSize), 0); err != nil {
				return err
			}
			if _, err := d.file.writer.Write(additionalData); err != nil {
				return err
			}
			d.layout.DataSize = newSize
		} else {
			d.layout.DataSize = newSize
		}
	}

	dsEncoded, err := format.EncodeDataspaceMessage(d.dspace.Dims, d.dspace.MaxDims)
	if err != nil {
		return err
	}

	layoutEncoded, err := format.EncodeLayoutMessage(d.layout.LayoutClass, d.layout.DataSize, d.layout.DataAddress, d.layout.ChunkDims)
	if err != nil {
		return err
	}

	for i, msg := range d.header.Messages {
		if msg.Type == format.MessageDataspace {
			d.header.Messages[i].Content = dsEncoded
			d.header.Messages[i].Size = uint16(10 + len(dsEncoded))
		} else if msg.Type == format.MessageDataset {
			d.header.Messages[i].Content = layoutEncoded
			d.header.Messages[i].Size = uint16(10 + len(layoutEncoded))
		}
	}

	actualSize := uint64(4) + uint64(1) + uint64(1) + uint64(d.super.LengthSize) + uint64(2) + uint64(d.super.OffsetSize)*uint64(d.header.NumberOfMessages)
	for _, msg := range d.header.Messages {
		actualSize += uint64(msg.Size)
	}
	if d.header.TotalHeaderSize < actualSize {
		d.header.TotalHeaderSize = ((actualSize + 1023) / 1024) * 1024
	}

	if _, err := d.file.writer.Seek(int64(d.addr), 0); err != nil {
		return err
	}
	if _, err := format.WriteObjectHeader(d.file.writer, d.super, d.header); err != nil {
		return err
	}

	return nil
}

func (d *dataset) Flush() error {
	if d.closed {
		return ErrClosedDataset
	}

	if f, ok := d.file.writer.(*os.File); ok {
		return f.Sync()
	}

	return nil
}

func (d *dataset) copyAttributes(dest *dataset) {
	srcAttrs := d.Attributes()
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

func (d *dataset) Close() error {
	if d.closed {
		return nil
	}
	d.closed = true
	if d.chunkCache != nil {
		d.chunkCache.Clear()
	}
	globalFileCache.Remove(d.addr)
	return nil
}

func encodeData(data interface{}, dtype Datatype, space Dataspace) ([]byte, error) {
	rv := reflect.ValueOf(data)

	if dtype.Class == DatatypeVarLength {
		if rv.Kind() != reflect.Slice {
			return nil, ErrInvalidData
		}

		totalSize := uint64(0)
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i)
			if elem.Kind() != reflect.Slice && elem.Kind() != reflect.Array {
				return nil, ErrInvalidData
			}
			totalSize += 4 + uint64(elem.Len())*uint64(dtype.BaseType.Size)
		}

		buf := make([]byte, totalSize)
		_, err := encodeSliceData(buf, rv, dtype)
		return buf, err
	}

	dataSize := space.Size() * uint64(dtype.Size)
	buf := make([]byte, dataSize)

	switch rv.Kind() {
	case reflect.Slice:
		return encodeSliceData(buf, rv, dtype)
	case reflect.Array:
		return encodeSliceData(buf, rv.Slice(0, rv.Len()), dtype)
	default:
		return encodeScalarData(buf, rv, dtype)
	}
}

func encodeSliceData(buf []byte, rv reflect.Value, dtype Datatype) ([]byte, error) {
	switch dtype.Class {
	case DatatypeInteger:
		switch dtype.Size {
		case 1:
			for i := 0; i < rv.Len(); i++ {
				switch rv.Index(i).Kind() {
				case reflect.Int, reflect.Int8:
					buf[i] = uint8(rv.Index(i).Int())
				case reflect.Uint, reflect.Uint8:
					buf[i] = uint8(rv.Index(i).Uint())
				}
			}
		case 2:
			for i := 0; i < rv.Len(); i++ {
				switch rv.Index(i).Kind() {
				case reflect.Int, reflect.Int16:
					binary.LittleEndian.PutUint16(buf[2*i:], uint16(rv.Index(i).Int()))
				case reflect.Uint, reflect.Uint16:
					binary.LittleEndian.PutUint16(buf[2*i:], uint16(rv.Index(i).Uint()))
				}
			}
		case 4:
			for i := 0; i < rv.Len(); i++ {
				switch rv.Index(i).Kind() {
				case reflect.Int, reflect.Int32:
					binary.LittleEndian.PutUint32(buf[4*i:], uint32(rv.Index(i).Int()))
				case reflect.Uint, reflect.Uint32:
					binary.LittleEndian.PutUint32(buf[4*i:], uint32(rv.Index(i).Uint()))
				}
			}
		case 8:
			for i := 0; i < rv.Len(); i++ {
				switch rv.Index(i).Kind() {
				case reflect.Int, reflect.Int64:
					binary.LittleEndian.PutUint64(buf[8*i:], uint64(rv.Index(i).Int()))
				case reflect.Uint, reflect.Uint64:
					binary.LittleEndian.PutUint64(buf[8*i:], rv.Index(i).Uint())
				}
			}
		}
	case DatatypeFloat:
		switch dtype.Size {
		case 4:
			for i := 0; i < rv.Len(); i++ {
				binary.LittleEndian.PutUint32(buf[4*i:], math.Float32bits(float32(rv.Index(i).Float())))
			}
		case 8:
			for i := 0; i < rv.Len(); i++ {
				binary.LittleEndian.PutUint64(buf[8*i:], math.Float64bits(rv.Index(i).Float()))
			}
		}
	case DatatypeString:
		for i := 0; i < rv.Len(); i++ {
			str := rv.Index(i).String()
			copy(buf[i*int(dtype.Size):], str)
		}
	case DatatypeCompound:
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i)
			if elem.Kind() == reflect.Complex64 || elem.Kind() == reflect.Complex128 {
				if dtype.Size == 8 {
					c64 := complex64(elem.Complex())
					binary.LittleEndian.PutUint32(buf[8*i:], math.Float32bits(real(c64)))
					binary.LittleEndian.PutUint32(buf[8*i+4:], math.Float32bits(imag(c64)))
				} else if dtype.Size == 16 {
					c128 := elem.Complex()
					binary.LittleEndian.PutUint64(buf[16*i:], math.Float64bits(real(c128)))
					binary.LittleEndian.PutUint64(buf[16*i+8:], math.Float64bits(imag(c128)))
				}
			} else if elem.Kind() != reflect.Struct {
				return nil, ErrInvalidData
			} else {
				elemOffset := i * int(dtype.Size)
				for _, field := range dtype.Fields {
					fieldVal := elem.FieldByName(field.Name)
					if !fieldVal.IsValid() {
						continue
					}

					fieldBuf := buf[elemOffset+int(field.Offset) : elemOffset+int(field.Offset)+int(field.Datatype.Size)]
					_, err := encodeScalarData(fieldBuf, fieldVal, field.Datatype)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	case DatatypeEnum:
		baseSize := uint32(4)
		if dtype.BaseType != nil {
			baseSize = dtype.BaseType.Size
		}
		for i := 0; i < rv.Len(); i++ {
			val := rv.Index(i).Int()
			switch baseSize {
			case 1:
				buf[i] = uint8(val)
			case 2:
				binary.LittleEndian.PutUint16(buf[2*i:], uint16(val))
			case 4:
				binary.LittleEndian.PutUint32(buf[4*i:], uint32(val))
			case 8:
				binary.LittleEndian.PutUint64(buf[8*i:], uint64(val))
			}
		}
	case DatatypeArray:
		if dtype.BaseType == nil {
			return nil, ErrInvalidData
		}
		elemSize := dtype.BaseType.Size
		elemKind := rv.Type().Elem().Kind()
		if elemKind == reflect.Slice || elemKind == reflect.Array {
			elemSize = dtype.Size
			for i := 0; i < rv.Len(); i++ {
				elemBuf := buf[i*int(elemSize) : (i+1)*int(elemSize)]
				elem := rv.Index(i)
				if _, err := encodeSliceData(elemBuf, elem, *dtype.BaseType); err != nil {
					return nil, err
				}
			}
		} else {
			for i := 0; i < rv.Len(); i++ {
				elemBuf := buf[i*int(elemSize) : (i+1)*int(elemSize)]
				if _, err := encodeScalarData(elemBuf, rv.Index(i), *dtype.BaseType); err != nil {
					return nil, err
				}
			}
		}
	case DatatypeVarLength:
		if dtype.BaseType == nil {
			return nil, ErrInvalidData
		}
		offset := 0
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i)
			if elem.Kind() != reflect.Slice && elem.Kind() != reflect.Array {
				return nil, ErrInvalidData
			}
			itemCount := uint32(elem.Len())
			if offset+4 > len(buf) {
				return nil, ErrInvalidData
			}
			binary.LittleEndian.PutUint32(buf[offset:offset+4], itemCount)
			offset += 4

			elemDataSize := uint64(itemCount) * uint64(dtype.BaseType.Size)
			if offset+int(elemDataSize) > len(buf) {
				return nil, ErrInvalidData
			}
			elemBuf := buf[offset : offset+int(elemDataSize)]
			elemSlice := elem.Slice(0, elem.Len())
			if _, err := encodeSliceData(elemBuf, elemSlice, *dtype.BaseType); err != nil {
				return nil, err
			}
			offset += int(elemDataSize)
		}
	case DatatypeReference:
		for i := 0; i < rv.Len(); i++ {
			val := rv.Index(i).Uint()
			binary.LittleEndian.PutUint64(buf[8*i:], val)
		}
	default:
		return nil, ErrInvalidData
	}

	return buf, nil
}

func encodeScalarData(buf []byte, rv reflect.Value, dtype Datatype) ([]byte, error) {
	switch dtype.Class {
	case DatatypeInteger:
		switch dtype.Size {
		case 1:
			switch rv.Kind() {
			case reflect.Int, reflect.Int8:
				buf[0] = uint8(rv.Int())
			case reflect.Uint, reflect.Uint8:
				buf[0] = uint8(rv.Uint())
			}
		case 2:
			switch rv.Kind() {
			case reflect.Int, reflect.Int16:
				binary.LittleEndian.PutUint16(buf, uint16(rv.Int()))
			case reflect.Uint, reflect.Uint16:
				binary.LittleEndian.PutUint16(buf, uint16(rv.Uint()))
			}
		case 4:
			switch rv.Kind() {
			case reflect.Int, reflect.Int32:
				binary.LittleEndian.PutUint32(buf, uint32(rv.Int()))
			case reflect.Uint, reflect.Uint32:
				binary.LittleEndian.PutUint32(buf, uint32(rv.Uint()))
			}
		case 8:
			switch rv.Kind() {
			case reflect.Int, reflect.Int64:
				binary.LittleEndian.PutUint64(buf, uint64(rv.Int()))
			case reflect.Uint, reflect.Uint64:
				binary.LittleEndian.PutUint64(buf, rv.Uint())
			}
		}
	case DatatypeFloat:
		switch dtype.Size {
		case 4:
			binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(rv.Float())))
		case 8:
			binary.LittleEndian.PutUint64(buf, math.Float64bits(rv.Float()))
		}
	case DatatypeString:
		str := rv.String()
		copy(buf, str)
	default:
		return nil, ErrInvalidData
	}

	return buf, nil
}

func decodeData(buf []byte, data interface{}, dtype Datatype) error {
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Ptr {
		return ErrInvalidData
	}
	rv = rv.Elem()

	switch rv.Kind() {
	case reflect.Slice:
		return decodeSliceData(buf, rv, dtype)
	case reflect.Array:
		return decodeSliceData(buf, rv.Slice(0, rv.Len()), dtype)
	default:
		return decodeScalarData(buf, rv, dtype)
	}
}

func decodeSliceData(buf []byte, rv reflect.Value, dtype Datatype) error {
	elemKind := rv.Type().Elem().Kind()

	if dtype.Class != DatatypeVarLength {
		expectedLen := len(buf) / int(dtype.Size)
		if rv.IsNil() || rv.Len() == 0 {
			newSlice := reflect.MakeSlice(rv.Type(), expectedLen, expectedLen)
			rv.Set(newSlice)
		} else if rv.Len() > expectedLen {
			newSlice := reflect.MakeSlice(rv.Type(), expectedLen, expectedLen)
			rv.Set(newSlice)
		}
	}

	switch dtype.Class {
	case DatatypeVarLength:
		if dtype.BaseType == nil {
			return ErrInvalidData
		}
		count := 0
		offset := 0
		for offset < len(buf) {
			if offset+4 > len(buf) {
				break
			}
			itemCount := binary.LittleEndian.Uint32(buf[offset : offset+4])
			offset += 4 + int(itemCount)*int(dtype.BaseType.Size)
			count++
		}

		if rv.IsNil() || rv.Len() == 0 {
			newSlice := reflect.MakeSlice(rv.Type(), count, count)
			rv.Set(newSlice)
		} else if rv.Len() > count {
			newSlice := reflect.MakeSlice(rv.Type(), count, count)
			rv.Set(newSlice)
		}

		offset = 0
		for i := 0; i < rv.Len(); i++ {
			if offset+4 > len(buf) {
				return ErrInvalidData
			}
			itemCount := binary.LittleEndian.Uint32(buf[offset : offset+4])
			offset += 4

			elemType := rv.Index(i).Type()
			if elemType.Kind() == reflect.Slice {
				elemSlice := reflect.MakeSlice(elemType, int(itemCount), int(itemCount))
				elemDataSize := uint64(itemCount) * uint64(dtype.BaseType.Size)
				if offset+int(elemDataSize) > len(buf) {
					return ErrInvalidData
				}
				elemBuf := buf[offset : offset+int(elemDataSize)]
				if err := decodeSliceData(elemBuf, elemSlice, *dtype.BaseType); err != nil {
					return err
				}
				rv.Index(i).Set(elemSlice)
				offset += int(elemDataSize)
			}
		}
		return nil
	case DatatypeCompound:
		if rv.Len() == 0 {
			return nil
		}
		if elemKind != reflect.Struct {
			return ErrInvalidData
		}
		if len(dtype.Fields) == 0 {
			return ErrInvalidData
		}
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i)
			if elem.Kind() != reflect.Struct {
				return ErrInvalidData
			}
			elemOffset := i * int(dtype.Size)
			for _, field := range dtype.Fields {
				fieldVal := elem.FieldByName(field.Name)
				if !fieldVal.IsValid() {
					continue
				}

				fieldBuf := buf[elemOffset+int(field.Offset) : elemOffset+int(field.Offset)+int(field.Datatype.Size)]
				fieldDtype := field.Datatype

				var class DatatypeClass
				switch fieldDtype.Class {
				case 0:
					class = DatatypeInteger
				case 1:
					class = DatatypeFloat
				case 2:
					class = DatatypeString
				case 3:
					class = DatatypeCompound
				default:
					class = DatatypeClass(fieldDtype.Class)
				}

				if class != DatatypeInteger && class != DatatypeFloat && class != DatatypeString && class != DatatypeCompound {
					return ErrInvalidData
				}

				err := decodeScalarData(fieldBuf, fieldVal, Datatype{
					Class:     class,
					Size:      fieldDtype.Size,
					ByteOrder: ByteOrder(fieldDtype.ByteOrder),
					Sign:      Sign(fieldDtype.Sign),
					FloatBits: fieldDtype.FloatBits,
				})
				if err != nil {
					return err
				}
			}
		}
		return nil
	case DatatypeInteger:
		for i := 0; i < rv.Len(); i++ {
			var val int64
			switch dtype.Size {
			case 1:
				val = int64(buf[i])
			case 2:
				val = int64(binary.LittleEndian.Uint16(buf[2*i:]))
			case 4:
				val = int64(binary.LittleEndian.Uint32(buf[4*i:]))
			case 8:
				val = int64(binary.LittleEndian.Uint64(buf[8*i:]))
			}
			switch elemKind {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				rv.Index(i).SetInt(val)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				rv.Index(i).SetUint(uint64(val))
			case reflect.Float32, reflect.Float64:
				rv.Index(i).SetFloat(float64(val))
			}
		}
	case DatatypeFloat:
		for i := 0; i < rv.Len(); i++ {
			var val float64
			switch dtype.Size {
			case 4:
				val = float64(math.Float32frombits(binary.LittleEndian.Uint32(buf[4*i:])))
			case 8:
				val = math.Float64frombits(binary.LittleEndian.Uint64(buf[8*i:]))
			}
			switch elemKind {
			case reflect.Float32, reflect.Float64:
				rv.Index(i).SetFloat(val)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				rv.Index(i).SetInt(int64(val))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				rv.Index(i).SetUint(uint64(val))
			}
		}
	case DatatypeString:
		for i := 0; i < rv.Len(); i++ {
			str := string(buf[i*int(dtype.Size) : (i+1)*int(dtype.Size)])
			if idx := strings.Index(str, "\x00"); idx >= 0 {
				str = str[:idx]
			}
			rv.Index(i).SetString(str)
		}
	case DatatypeEnum:
		baseSize := uint32(4)
		if dtype.BaseType != nil {
			baseSize = dtype.BaseType.Size
		}
		for i := 0; i < rv.Len(); i++ {
			var val int64
			switch baseSize {
			case 1:
				val = int64(buf[i])
			case 2:
				val = int64(binary.LittleEndian.Uint16(buf[2*i:]))
			case 4:
				val = int64(binary.LittleEndian.Uint32(buf[4*i:]))
			case 8:
				val = int64(binary.LittleEndian.Uint64(buf[8*i:]))
			}
			rv.Index(i).SetInt(val)
		}
	case DatatypeArray:
		if dtype.BaseType == nil {
			return ErrInvalidData
		}
		elemSize := dtype.BaseType.Size
		elemKind := rv.Type().Elem().Kind()
		if elemKind == reflect.Slice || elemKind == reflect.Array {
			elemSize = dtype.Size
			for i := 0; i < rv.Len(); i++ {
				elemBuf := buf[i*int(elemSize) : (i+1)*int(elemSize)]
				elem := rv.Index(i)
				if elem.Kind() == reflect.Slice {
					numElements := int(elemSize) / int(dtype.BaseType.Size)
					newSlice := reflect.MakeSlice(elem.Type(), numElements, numElements)
					if err := decodeSliceData(elemBuf, newSlice, *dtype.BaseType); err != nil {
						return err
					}
					rv.Index(i).Set(newSlice)
				} else {
					if err := decodeScalarData(elemBuf, elem, *dtype.BaseType); err != nil {
						return err
					}
				}
			}
		} else {
			for i := 0; i < rv.Len(); i++ {
				elemBuf := buf[i*int(elemSize) : (i+1)*int(elemSize)]
				if err := decodeScalarData(elemBuf, rv.Index(i), *dtype.BaseType); err != nil {
					return err
				}
			}
		}
	case DatatypeReference:
		for i := 0; i < rv.Len(); i++ {
			val := binary.LittleEndian.Uint64(buf[8*i : 8*(i+1)])
			rv.Index(i).SetUint(val)
		}
	default:
		return ErrInvalidData
	}

	return nil
}

func decodeScalarData(buf []byte, rv reflect.Value, dtype Datatype) error {
	switch dtype.Class {
	case DatatypeInteger:
		switch dtype.Size {
		case 1:
			if rv.Kind() == reflect.Uint || rv.Kind() == reflect.Uint8 {
				rv.SetUint(uint64(buf[0]))
			} else {
				rv.SetInt(int64(buf[0]))
			}
		case 2:
			val := binary.LittleEndian.Uint16(buf)
			if rv.Kind() == reflect.Uint || rv.Kind() == reflect.Uint16 {
				rv.SetUint(uint64(val))
			} else {
				rv.SetInt(int64(val))
			}
		case 4:
			val := binary.LittleEndian.Uint32(buf)
			if rv.Kind() == reflect.Uint || rv.Kind() == reflect.Uint32 {
				rv.SetUint(uint64(val))
			} else {
				rv.SetInt(int64(val))
			}
		case 8:
			val := binary.LittleEndian.Uint64(buf)
			if rv.Kind() == reflect.Uint || rv.Kind() == reflect.Uint64 {
				rv.SetUint(val)
			} else {
				rv.SetInt(int64(val))
			}
		}
	case DatatypeFloat:
		switch dtype.Size {
		case 4:
			rv.SetFloat(float64(math.Float32frombits(binary.LittleEndian.Uint32(buf))))
		case 8:
			rv.SetFloat(math.Float64frombits(binary.LittleEndian.Uint64(buf)))
		}
	case DatatypeString:
		str := string(buf)
		if idx := len(str); idx > 0 {
			for i := 0; i < idx; i++ {
				if buf[i] == 0 {
					str = str[:i]
					break
				}
			}
		}
		rv.SetString(str)
	default:
		return ErrInvalidData
	}

	return nil
}

type virtualDataset struct {
	name   string
	group  *group
	layout VirtualLayout
	closed bool
}

func (d *virtualDataset) Name() string {
	return d.name
}

func (d *virtualDataset) Parent() Group {
	return d.group
}

func (d *virtualDataset) File() *File {
	return d.group.file
}

func (d *virtualDataset) Datatype() Datatype {
	return d.layout.Datatype
}

func (d *virtualDataset) Dataspace() Dataspace {
	return d.layout.Dataspace
}

func (d *virtualDataset) Read(data interface{}) error {
	if d.closed {
		return ErrClosedDataset
	}

	dataSize := d.layout.Dataspace.Size() * uint64(d.layout.Datatype.Size)
	buf := make([]byte, dataSize)

	for _, source := range d.layout.Sources {
		extFile, err := OpenFile(source.FilePath, ReadOnly)
		if err != nil {
			return err
		}

		extDset, err := extFile.GetDataset(source.DatasetPath)
		if err != nil {
			extFile.Close()
			return err
		}

		var sourceData interface{}
		var sourceSpace Dataspace
		if len(source.Selection.Hyperslabs) > 0 {
			err = extDset.ReadSlice(&sourceData, source.Selection)
			sourceSpace = extDset.Dataspace()
			for _, h := range source.Selection.Hyperslabs {
				sourceSpace.Dims = h.Count
				break
			}
		} else {
			err = extDset.Read(&sourceData)
			sourceSpace = extDset.Dataspace()
		}

		sourceEncoded, encodeErr := encodeData(sourceData, d.layout.Datatype, sourceSpace)

		extFile.Close()

		if err != nil {
			return err
		}
		if encodeErr != nil {
			return encodeErr
		}

		var dstOffset uint64
		if len(source.TargetStart) > 0 {
			rank := len(d.layout.Dataspace.Dims)
			stride := uint64(1)
			for i := rank - 1; i >= 0; i-- {
				start := uint64(0)
				if i < len(source.TargetStart) {
					start = source.TargetStart[i]
				}
				dstOffset += start * stride
				stride *= d.layout.Dataspace.Dims[i]
			}
		} else if len(source.Selection.Hyperslabs) > 0 {
			rank := len(d.layout.Dataspace.Dims)
			stride := uint64(1)
			for i := rank - 1; i >= 0; i-- {
				dstOffset += source.Selection.Hyperslabs[0].Start[i] * stride
				stride *= d.layout.Dataspace.Dims[i]
			}
		}

		dstOffset *= uint64(d.layout.Datatype.Size)
		dataSize := uint64(len(sourceEncoded))
		if dstOffset+dataSize > uint64(len(buf)) {
			dataSize = uint64(len(buf)) - dstOffset
		}
		copy(buf[dstOffset:dstOffset+dataSize], sourceEncoded[:dataSize])
	}

	return decodeData(buf, data, d.layout.Datatype)
}

func (d *virtualDataset) Write(data interface{}) error {
	return errors.New("hdf5: virtual datasets are read-only")
}

func (d *virtualDataset) ReadSlice(data interface{}, sel Selection) error {
	if d.closed {
		return ErrClosedDataset
	}

	if len(sel.Hyperslabs) == 0 {
		return errors.New("hdf5: selection is empty")
	}

	rank := len(d.layout.Dataspace.Dims)

	totalSize := uint64(0)
	for _, h := range sel.Hyperslabs {
		if len(h.Start) != rank || len(h.Count) != rank {
			return errors.New("hdf5: hyperslab dimensions mismatch")
		}
		slabSize := uint64(1)
		for i := range h.Count {
			slabSize *= h.Count[i]
		}
		totalSize += slabSize
	}

	dataSize := totalSize * uint64(d.layout.Datatype.Size)
	buf := make([]byte, dataSize)

	fullBuf := make([]byte, d.layout.Dataspace.Size()*uint64(d.layout.Datatype.Size))
	if err := d.Read(&fullBuf); err != nil {
		return err
	}

	for i, h := range sel.Hyperslabs {
		sourceOffset := uint64(0)
		stride := uint64(1)
		for j := rank - 1; j >= 0; j-- {
			sourceOffset += h.Start[j] * stride
			stride *= d.layout.Dataspace.Dims[j]
		}
		sourceOffset *= uint64(d.layout.Datatype.Size)

		slabSize := uint64(1)
		for j := range h.Count {
			slabSize *= h.Count[j]
		}
		slabSize *= uint64(d.layout.Datatype.Size)

		if sourceOffset+slabSize > uint64(len(fullBuf)) {
			slabSize = uint64(len(fullBuf)) - sourceOffset
		}

		dstOffset := uint64(0)
		if i > 0 {
			for k := 0; k < i; k++ {
				prevSlabSize := uint64(1)
				for j := range sel.Hyperslabs[k].Count {
					prevSlabSize *= sel.Hyperslabs[k].Count[j]
				}
				dstOffset += prevSlabSize * uint64(d.layout.Datatype.Size)
			}
		}

		copy(buf[dstOffset:dstOffset+slabSize], fullBuf[sourceOffset:sourceOffset+slabSize])
	}

	return decodeData(buf, data, d.layout.Datatype)
}

func (d *virtualDataset) WriteSlice(data interface{}, sel Selection) error {
	return errors.New("hdf5: virtual datasets are read-only")
}

func (d *virtualDataset) Resize(newDims []uint64) error {
	return errors.New("hdf5: virtual datasets cannot be resized")
}

func (d *virtualDataset) CreateReference(sel Selection) (RegionReference, error) {
	return RegionReference{}, errors.New("hdf5: virtual datasets do not support references")
}

func (d *virtualDataset) CreateObjectReference() (ObjectReference, error) {
	return ObjectReference{}, errors.New("hdf5: virtual datasets do not support object references")
}

func (d *virtualDataset) Attributes() *AttributeManager {
	return &AttributeManager{obj: d}
}

func (d *virtualDataset) Flush() error {
	return nil
}

func (d *virtualDataset) Close() error {
	if d.closed {
		return nil
	}
	d.closed = true
	return nil
}
