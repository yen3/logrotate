package main

import (
	"fmt"
	"os"
	"sync"
)

type File struct {
	*os.File
	Path           string
	MaxFileSize    int32 // unit: bytes 
	MaxBackupFiles int32
	currentFileSize int64 // unit: bytes	
	writeLock      sync.Mutex
}

func NewLogger(path string, maxFileSize int32, maxBackupFiles int32) (File*, error) {
}

func (logger *File) Write(b []byte) (int, error) {
	logger.writeLock.Lock()
	defer logger.writeLock.Unlock()

	return logger.File.Write(b)
}

func main() {
	fmt.Println("vim-go")
}
