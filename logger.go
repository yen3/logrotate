package main

import (
	"fmt"
	"os"
	"sync"
)

type File struct {
	*os.File
	Path           string
	MaxFileSize    int64 // unit: Mbytes
	MaxBackupFiles int64
	writeLock      sync.Mutex
}

func (logger *File) Write(b []byte) (int, error) {
	logger.writeLock.Lock()
	defer logger.writeLock.Unlock()

	return logger.File.Write(b)
}

func main() {
	fmt.Println("vim-go")
}
