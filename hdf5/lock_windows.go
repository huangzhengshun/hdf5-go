//go:build windows
// +build windows

package hdf5

import (
	"os"

	"golang.org/x/sys/windows"
)

func lockFile(file *os.File, blocking bool) error {
	flags := uint32(windows.LOCKFILE_EXCLUSIVE_LOCK)
	if !blocking {
		flags |= windows.LOCKFILE_FAIL_IMMEDIATELY
	}

	overlapped := &windows.Overlapped{}
	err := windows.LockFileEx(windows.Handle(file.Fd()), flags, 0, 1, 0, overlapped)
	return err
}

func unlockFile(file *os.File) error {
	overlapped := &windows.Overlapped{}
	err := windows.UnlockFileEx(windows.Handle(file.Fd()), 0, 1, 0, overlapped)
	return err
}

