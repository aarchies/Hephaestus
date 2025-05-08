package file

import (
	"os"
)

// FileLocker is UDS access lock
type FileLocker struct {
	Path string
	File *os.File
}
