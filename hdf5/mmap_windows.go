//go:build windows
// +build windows

package hdf5

import (
	"os"
	"syscall"
	"unsafe"
)

func mmapFile(file *os.File, size int64) ([]byte, error) {
	if size <= 0 {
		return nil, nil
	}

	hFileMapping, err := syscall.CreateFileMapping(syscall.Handle(file.Fd()), nil, syscall.PAGE_READWRITE, uint32(size>>32), uint32(size), nil)
	if err != nil {
		return nil, err
	}

	addr, err := syscall.MapViewOfFile(hFileMapping, syscall.FILE_MAP_READ|syscall.FILE_MAP_WRITE, 0, 0, uintptr(size))
	if err != nil {
		syscall.CloseHandle(hFileMapping)
		return nil, err
	}

	return unsafe.Slice((*byte)(unsafe.Pointer(addr)), size), nil
}

func munmapFile(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return syscall.UnmapViewOfFile(uintptr(unsafe.Pointer(&data[0])))
}


