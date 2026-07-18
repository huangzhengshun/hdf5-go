//go:build linux || darwin || freebsd || openbsd
// +build linux darwin freebsd openbsd

package hdf5

import (
	"os"
	"syscall"
)

func lockFile(file *os.File, blocking bool) error {
	flags := syscall.LOCK_EX
	if !blocking {
		flags |= syscall.LOCK_NB
	}

	err := syscall.Flock(int(file.Fd()), flags)
	return err
}

func unlockFile(file *os.File) error {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	return err
}

