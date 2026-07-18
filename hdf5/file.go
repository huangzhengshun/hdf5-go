package hdf5

import (
	"errors"
	"io"
	"os"

	"github.com/hdf5-go/hdf5/format"
)

type FreeSpaceInfo struct {
	TotalFreeSpace  uint64
	FreeSpaceBlocks int
	SmallestBlock   uint64
	LargestBlock    uint64
}

type SpaceAllocationStrategy uint8

const (
	SpaceStrategyDefault SpaceAllocationStrategy = 0
	SpaceStrategyFIFO    SpaceAllocationStrategy = 1
	SpaceStrategyBest    SpaceAllocationStrategy = 2
	SpaceStrategyWorst   SpaceAllocationStrategy = 3
)

type memoryBuffer struct {
	data []byte
	pos  int64
}

func (m *memoryBuffer) Read(p []byte) (n int, err error) {
	if m.pos >= int64(len(m.data)) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.pos:])
	m.pos += int64(n)
	return
}

func (m *memoryBuffer) Write(p []byte) (n int, err error) {
	endPos := m.pos + int64(len(p))
	if endPos > int64(len(m.data)) {
		newData := make([]byte, endPos)
		copy(newData, m.data)
		m.data = newData
	}
	copy(m.data[m.pos:], p)
	m.pos = endPos
	return len(p), nil
}

func (m *memoryBuffer) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = m.pos + offset
	case io.SeekEnd:
		newPos = int64(len(m.data)) + offset
	default:
		return 0, errors.New("invalid whence")
	}
	if newPos < 0 {
		return 0, errors.New("negative position")
	}
	m.pos = newPos
	return newPos, nil
}

func OpenFile(name string, mode FileMode) (*File, error) {
	return openFileWithBuffer(name, mode, nil)
}

func CreateMemoryFile() (*File, error) {
	return openFileWithBuffer("", Create, &memoryBuffer{})
}

func openFileWithBuffer(name string, mode FileMode, buf *memoryBuffer) (*File, error) {
	f := &File{name: name, mode: mode}

	if buf != nil {
		f.reader = buf
		f.writer = buf
	} else {
		switch mode {
		case ReadOnly:
			file, err := os.Open(name)
			if err != nil {
				return nil, err
			}
			f.reader = file
			f.writer = nil
		case ReadWrite:
			file, err := os.OpenFile(name, os.O_RDWR, 0644)
			if err != nil {
				return nil, err
			}
			f.reader = file
			f.writer = file
		case Create:
			file, err := os.Create(name)
			if err != nil {
				return nil, err
			}
			f.reader = file
			f.writer = file
		case CreateExclusive:
			file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
			if err != nil {
				return nil, err
			}
			f.reader = file
			f.writer = file
		case Append:
			file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				return nil, err
			}
			f.reader = file
			f.writer = file
		default:
			return nil, os.ErrInvalid
		}
	}

	if mode == ReadOnly || mode == ReadWrite || mode == Append {
		sb, err := format.ReadSuperblock(f.reader)
		if err != nil {
			f.reader.(*os.File).Close()
			return nil, err
		}
		f.super = sb
		rootOH, err := format.ReadObjectHeader(f.reader, sb, sb.RootGroupAddr)
		if err != nil {
			f.reader.(*os.File).Close()
			return nil, err
		}
		links := parseLinksFromObjectHeader(rootOH)
		f.root = &group{
			name:   "/",
			addr:   sb.RootGroupAddr,
			file:   f,
			header: rootOH,
			super:  sb,
			links:  links,
		}
	} else {
		sb := &format.Superblock{
			Version:              format.SuperblockVersion2,
			FreeSpaceManagerAddr: 0,
			RootGroupAddr:        0,
			FileConsistencyFlags: 0,
			OffsetSize:           8,
			LengthSize:           8,
			GroupLeafNodeK:       16,
			GroupInternalNodeK:   16,
		}
		f.super = sb

		rootOH := &format.ObjectHeader{
			Version:          format.ObjectHeaderVersion2,
			Flags:            0,
			TotalHeaderSize:  1024,
			NumberOfMessages: 0,
			Messages:         []format.ObjectHeaderMessage{},
		}

		if err := format.WriteSuperblock(f.writer, sb); err != nil {
			f.writer.(*os.File).Close()
			return nil, err
		}

		rootAddr, err := format.WriteObjectHeader(f.writer, sb, rootOH)
		if err != nil {
			f.writer.(*os.File).Close()
			return nil, err
		}

		padding := make([]byte, 1024)
		if _, err := f.writer.Write(padding); err != nil {
			f.writer.(*os.File).Close()
			return nil, err
		}

		sb.RootGroupAddr = rootAddr
		if _, err := f.writer.Seek(0, 0); err != nil {
			f.writer.(*os.File).Close()
			return nil, err
		}
		if err := format.WriteSuperblock(f.writer, sb); err != nil {
			f.writer.(*os.File).Close()
			return nil, err
		}

		if _, err := f.writer.Seek(0, 2); err != nil {
			f.writer.(*os.File).Close()
			return nil, err
		}

		f.root = &group{
			name:   "/",
			addr:   rootAddr,
			file:   f,
			header: rootOH,
			super:  sb,
		}
	}

	return f, nil
}

func (f *File) Close() error {
	if f.closed {
		return nil
	}
	f.closed = true

	if f.root != nil {
		f.root.Close()
	}

	if f.writer != nil {
		if err := f.flushFreeSpace(); err != nil {
			return err
		}
	}

	if f.reader != nil {
		if mbuf, ok := f.reader.(*mmapBuffer); ok {
			munmapFile(mbuf.data)
			return mbuf.file.Close()
		}
		if file, ok := f.reader.(*os.File); ok {
			if f.writer != nil {
				file.Sync()
			}
			return file.Close()
		}
	}

	return nil
}

func (f *File) flushFreeSpace() error {
	if f.writer == nil {
		return nil
	}

	endPos, _ := f.writer.Seek(0, 2)
	f.super.FreeSpaceManagerAddr = uint64(endPos)

	if _, err := f.writer.Seek(0, 0); err != nil {
		return err
	}
	if err := format.WriteSuperblock(f.writer, f.super); err != nil {
		return err
	}

	_, err := f.writer.Seek(endPos, 0)
	return err
}

type mmapBuffer struct {
	data []byte
	pos  int64
	file *os.File
}

func (m *mmapBuffer) Read(p []byte) (n int, err error) {
	if m.pos >= int64(len(m.data)) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.pos:])
	m.pos += int64(n)
	return
}

func (m *mmapBuffer) Write(p []byte) (n int, err error) {
	endPos := m.pos + int64(len(p))
	if endPos > int64(len(m.data)) {
		newSize := endPos * 2
		if err := munmapFile(m.data); err != nil {
			return 0, err
		}
		if err := m.file.Truncate(newSize); err != nil {
			return 0, err
		}
		newData, err := mmapFile(m.file, newSize)
		if err != nil {
			return 0, err
		}
		m.data = newData
	}
	copy(m.data[m.pos:], p)
	m.pos = endPos
	return len(p), nil
}

func (m *mmapBuffer) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = m.pos + offset
	case io.SeekEnd:
		newPos = int64(len(m.data)) + offset
	default:
		return 0, errors.New("invalid whence")
	}
	if newPos < 0 {
		return 0, errors.New("negative position")
	}
	m.pos = newPos
	return newPos, nil
}

func OpenMmapFile(name string, mode FileMode) (*File, error) {
	f := &File{name: name, mode: mode}

	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	size := info.Size()
	if size == 0 {
		size = 4096
		if err := file.Truncate(size); err != nil {
			file.Close()
			return nil, err
		}
	}

	data, err := mmapFile(file, size)
	if err != nil {
		file.Close()
		return nil, err
	}

	buf := &mmapBuffer{
		data: data,
		pos:  0,
		file: file,
	}

	f.reader = buf
	f.writer = buf

	if mode == ReadOnly || mode == ReadWrite || mode == Append || (mode == Create && info.Size() > 0) {
		sb, err := format.ReadSuperblock(buf)
		if err != nil {
			munmapFile(data)
			file.Close()
			return nil, err
		}
		f.super = sb
		rootOH, err := format.ReadObjectHeader(buf, sb, sb.RootGroupAddr)
		if err != nil {
			munmapFile(data)
			file.Close()
			return nil, err
		}
		links := parseLinksFromObjectHeader(rootOH)
		f.root = &group{
			name:   "/",
			addr:   sb.RootGroupAddr,
			file:   f,
			header: rootOH,
			super:  sb,
			links:  links,
		}
	} else {
		sb := &format.Superblock{
			Version:              format.SuperblockVersion2,
			FreeSpaceManagerAddr: 0,
			RootGroupAddr:        0,
			FileConsistencyFlags: 0,
			OffsetSize:           8,
			LengthSize:           8,
			GroupLeafNodeK:       16,
			GroupInternalNodeK:   16,
		}
		f.super = sb

		rootOH := &format.ObjectHeader{
			Version:          format.ObjectHeaderVersion2,
			Flags:            0,
			TotalHeaderSize:  1024,
			NumberOfMessages: 0,
			Messages:         []format.ObjectHeaderMessage{},
		}

		if err := format.WriteSuperblock(buf, sb); err != nil {
			munmapFile(data)
			file.Close()
			return nil, err
		}

		rootAddr, err := format.WriteObjectHeader(buf, sb, rootOH)
		if err != nil {
			munmapFile(data)
			file.Close()
			return nil, err
		}

		padding := make([]byte, 1024)
		if _, err := buf.Write(padding); err != nil {
			munmapFile(data)
			file.Close()
			return nil, err
		}

		sb.RootGroupAddr = rootAddr
		if _, err := buf.Seek(0, 0); err != nil {
			munmapFile(data)
			file.Close()
			return nil, err
		}
		if err := format.WriteSuperblock(buf, sb); err != nil {
			munmapFile(data)
			file.Close()
			return nil, err
		}

		if _, err := buf.Seek(0, 2); err != nil {
			munmapFile(data)
			file.Close()
			return nil, err
		}

		f.root = &group{
			name:   "/",
			addr:   rootAddr,
			file:   f,
			header: rootOH,
			super:  sb,
		}
	}

	return f, nil
}

func (f *File) Bytes() ([]byte, error) {
	if buf, ok := f.reader.(*memoryBuffer); ok {
		return buf.data, nil
	}
	return nil, errors.New("hdf5: Bytes() only supported for memory files")
}

func (f *File) ReclaimFreeSpace() error {
	if f.writer == nil {
		return ErrReadOnly
	}

	endPos, _ := f.writer.Seek(0, 2)
	if f.super.FreeSpaceManagerAddr > 0 && f.super.FreeSpaceManagerAddr < uint64(endPos) {
		if file, ok := f.writer.(*os.File); ok {
			if err := file.Truncate(int64(f.super.FreeSpaceManagerAddr)); err != nil {
				return err
			}
		}
		f.super.FreeSpaceManagerAddr = 0
		if _, err := f.writer.Seek(0, 0); err != nil {
			return err
		}
		return format.WriteSuperblock(f.writer, f.super)
	}

	return nil
}

func (f *File) GetFreeSpaceInfo() (FreeSpaceInfo, error) {
	if f.closed {
		return FreeSpaceInfo{}, ErrClosedFile
	}

	endPos, _ := f.reader.Seek(0, 2)
	totalSize := uint64(endPos)

	usedSize := f.super.FreeSpaceManagerAddr
	if usedSize == 0 {
		usedSize = totalSize
	}

	freeSpace := totalSize - usedSize

	return FreeSpaceInfo{
		TotalFreeSpace:  freeSpace,
		FreeSpaceBlocks: 1,
		SmallestBlock:   freeSpace,
		LargestBlock:    freeSpace,
	}, nil
}

func (f *File) CompactFile() error {
	if f.writer == nil {
		return ErrReadOnly
	}

	if err := f.ReclaimFreeSpace(); err != nil {
		return err
	}

	if file, ok := f.writer.(*os.File); ok {
		return file.Sync()
	}

	return nil
}

func (f *File) SetSpaceAllocationStrategy(strategy SpaceAllocationStrategy) error {
	if f.closed {
		return ErrClosedFile
	}

	f.super.FileConsistencyFlags = uint32(strategy)

	if _, err := f.writer.Seek(0, 0); err != nil {
		return err
	}
	return format.WriteSuperblock(f.writer, f.super)
}

func (f *File) GetSpaceAllocationStrategy() SpaceAllocationStrategy {
	return SpaceAllocationStrategy(f.super.FileConsistencyFlags)
}

func (f *File) GetUsedSpace() (uint64, error) {
	if f.closed {
		return 0, ErrClosedFile
	}

	return f.super.FreeSpaceManagerAddr, nil
}

func (f *File) GetTotalSpace() (uint64, error) {
	if f.closed {
		return 0, ErrClosedFile
	}

	endPos, _ := f.reader.Seek(0, 2)
	return uint64(endPos), nil
}


