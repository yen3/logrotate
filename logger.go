package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const sepMark = byte('\n')

type PathMetadata struct {
	Path      string
	BaseDir   string
	Basename  string
	Extension string
}

type File struct {
	*os.File
	PathInfo       *PathMetadata
	MaxFileSize    int64 // unit: bytes
	MaxBackupFiles int

	currentFileSize int64 // unit: bytes
}

func IsFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func GetFileSize(path string) int64 {
	stat, err := os.Stat(path)

	if os.IsNotExist(err) {
		return 0
	}

	return stat.Size()
}

func GenerateLogFilename(index int, dir string, basename string, extension string) string {
	if index == 0 {
		return fmt.Sprintf("%s/%s%s", dir, basename, extension)
	}

	return fmt.Sprintf("%s/%s-%d%s", dir, basename, index, extension)
}

func NewPathMetadata(path string) *PathMetadata {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Println(err)
		return nil
	}
	baseDir := filepath.Dir(absPath)
	extension := filepath.Ext(absPath)
	basename := filepath.Base(absPath)
	basename = basename[0 : len(basename)-len(extension)]

	return &PathMetadata{
		Path:      absPath,
		BaseDir:   baseDir,
		Basename:  basename,
		Extension: extension,
	}
}

func NewLogger(path string, maxFileSize int64, maxBackupFiles int) (*File, error) {
	lg := &File{
		PathInfo:        NewPathMetadata(path),
		MaxFileSize:     maxFileSize,
		MaxBackupFiles:  maxBackupFiles,
		currentFileSize: 0,
	}

	if IsFileExists(lg.PathInfo.Path) {
		lg.currentFileSize = GetFileSize(lg.PathInfo.Path)

		if lg.currentFileSize >= lg.MaxFileSize {
			// If the filesize is too big, delete the file rather than rotate it.
			// Otherwise, do normal file rotation
			if lg.currentFileSize >= lg.MaxFileSize*10 {
				if err := os.Remove(lg.PathInfo.Path); err != nil {
					return nil, err
				}
			} else {
				if err := lg.rotateFiles(); err != nil {
					return nil, err
				}
			}

			lg.currentFileSize = 0
		}
	}

	if err := lg.openLogFile(); err != nil {
		return nil, err
	}

	return lg, nil
}
func (lr *File) listPreviousLogFiles() ([]string, error) {
	var logPaths []string
	for i := 0; i <= lr.MaxBackupFiles-1; i++ {
		prevLogPath := GenerateLogFilename(i, lr.PathInfo.BaseDir, lr.PathInfo.Basename, lr.PathInfo.Extension)

		if IsFileExists(prevLogPath) {
			logPaths = append(logPaths, prevLogPath)
		}
	}

	return logPaths, nil
}

func (lr *File) rotateFiles() error {
	logPaths, err := lr.listPreviousLogFiles()
	if err != nil {
		return err
	}

	// No backup files
	if logPaths == nil {
		return nil
	}

	// If backup files is less then the limitation, generate a new filename
	if len(logPaths) < lr.MaxBackupFiles {
		newPath := GenerateLogFilename(len(logPaths), lr.PathInfo.BaseDir, lr.PathInfo.Basename, lr.PathInfo.Extension)
		logPaths = append(logPaths, newPath)
	}

	// Rotate files
	for i := len(logPaths) - 2; i >= 0; i-- {
		//fmt.Println(logPaths[i], logPaths[i+1])
		if err = os.Rename(logPaths[i], logPaths[i+1]); err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (lr *File) openLogFile() (err error) {
	lr.File, err = os.OpenFile(lr.PathInfo.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	return
}

func (lr *File) Close() (err error) {
	return lr.File.Close()
}

func (lr *File) truncateWriteFile(b []byte) ([]byte, error) {
	buf := b
	sepIndex := bytes.IndexByte(buf, sepMark)

	// Write the remaining part of the entry or maybe it's a new entry
	// and the entry is in the buffer totally.
	//
	// If the sepIndex is -1, it means the whole buffer can not form an
	// entry. Continue to write the file until the sep is found then
	// rotate
	if sepIndex != -1 {
		// Write the remaining part of the entry
		temp_buf := buf[0 : sepIndex+1]
		if _, err := lr.File.Write(temp_buf); err != nil {
			return nil, err
		}

		if err := lr.File.Close(); err != nil {
			return nil, err
		}

		// Rotate the file to backup file
		if err := lr.rotateFiles(); err != nil {
			return nil, err
		}

		// Reopen the file and reinit the filesize counter.
		lr.openLogFile()
		lr.currentFileSize = 0

		// The buf would be written to next file.
		// The writing action would be effect in the below code block.
		buf = buf[sepIndex+1 : len(buf)]
		if len(buf) == 0 {
			return nil, nil
		}
	}

	return buf, nil
}

func (lr *File) Write(b []byte) (int, error) {
	buf := b
	var err error

	// Ready to rotate file
	if lr.currentFileSize >= lr.MaxFileSize {
		buf, err = lr.truncateWriteFile(b)
		if err != nil {
			return -1, err
		}
	}

	// Write the file
	writeBytes, err := lr.File.Write(buf)
	if err != nil {
		return writeBytes, err
	}

	// Update filesize
	lr.currentFileSize += int64(writeBytes)

	return writeBytes, nil
}

func main() {
	fmt.Println("Hello World")
}
