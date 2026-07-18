//go:build linux || darwin || freebsd || openbsd
// +build linux darwin freebsd openbsd

package hdf5

import (
	"os"
	"syscall"
)

func mmapFile(file *os.File, size int64) ([]byte, error) {
	return syscall.Mmap(
		int(file.Fd()),
		0,
		int(size),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED,
	)
}

func munmapFile(data []byte) error {
	return syscall.Munmap(data)
}

