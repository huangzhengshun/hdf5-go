package hdf5

import (
	"encoding/binary"
	"errors"
	"os"
	"sync"

	"github.com/huangzhengshun/hdf5-go/format"
)

var atomicWriteMutexes sync.Map

func (d *dataset) WriteAtomic(data interface{}) error {
	if d.closed {
		return ErrClosedDataset
	}

	if d.layout.LayoutClass != format.LayoutChunked {
		return errors.New("hdf5: atomic write only supported for chunked datasets")
	}

	mutex := getOrCreateMutex(d.addr)
	mutex.Lock()
	defer mutex.Unlock()

	return d.Write(data)
}

func (d *dataset) writeChunkAtomic(chunkIndex uint64, data []byte) error {
	if d.closed {
		return ErrClosedDataset
	}

	mutex := getOrCreateMutex(d.addr)
	mutex.Lock()
	defer mutex.Unlock()

	if d.chunkIndex == nil {
		var err error
		d.chunkIndex, err = format.ReadChunkIndex(d.file.reader, d.layout.DataAddress)
		if err != nil {
			return err
		}
	}

	if uint64(len(d.chunkIndex.Entries)) <= chunkIndex {
		return errors.New("hdf5: chunk index out of range")
	}

	cacheData := make([]byte, len(data))
	copy(cacheData, data)

	if d.dtype.Shuffle {
		data = applyShuffle(data, int(d.dtype.Size))
	}

	writeBuf, err := compressData(data, d.dtype.Compression)
	if err != nil {
		return err
	}

	if d.dtype.Fletcher32 {
		checksum := computeFletcher32(data)
		checksumBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(checksumBuf, checksum)
		writeBuf = append(writeBuf, checksumBuf...)
	}

	if d.chunkIndex.Entries[chunkIndex].ChunkOffset == 0 {
		pos, _ := d.file.writer.Seek(0, 2)
		d.chunkIndex.Entries[chunkIndex].ChunkOffset = uint64(pos)
	}

	if _, err := d.file.writer.Seek(int64(d.chunkIndex.Entries[chunkIndex].ChunkOffset), 0); err != nil {
		return err
	}
	if _, err := d.file.writer.Write(writeBuf); err != nil {
		return err
	}

	d.chunkIndex.Entries[chunkIndex].ChunkSize = uint64(len(writeBuf))

	if d.chunkCache != nil {
		d.chunkCache.Set(chunkIndex, cacheData)
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

	for i, msg := range d.header.Messages {
		if msg.Type == format.MessageDataset {
			d.header.Messages[i].Content = layoutEncoded
			d.header.Messages[i].Size = uint16(10 + len(layoutEncoded))
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

func getOrCreateMutex(key uint64) *sync.Mutex {
	if v, ok := atomicWriteMutexes.Load(key); ok {
		return v.(*sync.Mutex)
	}

	mutex := &sync.Mutex{}
	if v, ok := atomicWriteMutexes.LoadOrStore(key, mutex); ok {
		return v.(*sync.Mutex)
	}
	return mutex
}
