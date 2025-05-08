//go:build !windows
// +build !windows

package file

import (
	"fmt"
	"netleap/pkg/logs"
	"os"

	"golang.org/x/sys/unix"
)

// Acquire lock
func (fl *FileLocker) Acquire() error {
	f, err := os.Create(fl.Path)
	if err != nil {
		return err
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		f.Close()
		return fmt.Errorf("failed to lock file: %s %s", fl.Path, err)
	}
	fl.File = f
	return nil
}

// Release lock
func (fl *FileLocker) Release() {
	if err := unix.Flock(int(fl.File.Fd()), unix.LOCK_UN); err != nil {
		logs.Errorf("failed to unlock file: %s %s", fl.Path, err)
	}
	if err := fl.File.Close(); err != nil {
		logs.Errorf("failed to close file: %s %s", fl.Path, err)
	}
	if err := os.Remove(fl.Path); err != nil {
		logs.Errorf("failed to remove file: %s %s", fl.Path, err)
	}
}
