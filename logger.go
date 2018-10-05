package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

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
	EndEntryMark   string // TODO: implement with it

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

func NewLogger(path string, maxFileSize int64, maxBackupFiles int, endEntryMark string) (*File, error) {
	lg := &File{
		PathInfo:        NewPathMetadata(path),
		MaxFileSize:     maxFileSize,
		MaxBackupFiles:  maxBackupFiles,
		EndEntryMark:    endEntryMark,
		currentFileSize: 0,
	}

	if IsFileExists(lg.PathInfo.Path) {
		lg.currentFileSize = GetFileSize(lg.PathInfo.Path)

		if lg.currentFileSize >= lg.MaxFileSize {
			// If the filesize is too big, delete the file rather than rotate it.
			// Otherwise, do normal file rotation
			if lg.currentFileSize >= lg.MaxFileSize*10 {
				err := os.Remove(lg.PathInfo.Path)
				if err != nil {
					return nil, err
				}
			} else {
				lg.rotateFiles()
			}

			lg.currentFileSize = 0
		}
	}

	lg.openLogFile()

	return lg, nil
}

func GenerateLogFilename(index int, dir string, basename string, extension string) string {
	if index == 0 {
		return fmt.Sprintf("%s/%s%s", dir, basename, extension)
	}

	return fmt.Sprintf("%s/%s-%d%s", dir, basename, index, extension)
}

func ListLogFiles(pathInfo *PathMetadata, maxLogFiles int) ([]string, error) {
	var logPaths []string
	for i := 0; i <= maxLogFiles-1; i++ {
		prevLogPath := GenerateLogFilename(i, pathInfo.BaseDir, pathInfo.Basename, pathInfo.Extension)

		if IsFileExists(prevLogPath) {
			logPaths = append(logPaths, prevLogPath)
		}
	}

	return logPaths, nil
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
		fmt.Println(logPaths[i], logPaths[i+1])
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

func (lr *File) Write(b []byte) (int, error) {
	writeBytes, err := lr.File.Write(b)

	if err != nil {
		return writeBytes, err
	}

	lr.currentFileSize += int64(writeBytes)
	if lr.currentFileSize > lr.MaxFileSize {
		lr.File.Close()
		lr.rotateFiles()
		lr.openLogFile()
	}

	return writeBytes, nil
}
