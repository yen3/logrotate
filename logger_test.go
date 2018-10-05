package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsFileExists(t *testing.T) {
	testData := []struct {
		path    string
		isExist bool
	}{
		{"./test_logrotate/test-empty.log", true},
		{"./test_logrotate/test-empty-1.log", true},
		{"./test_logrotate/doesnot_exist.log", false},
	}

	for i := range testData {
		if IsFileExists(testData[i].path) != testData[i].isExist {
			t.Fail()
		}
	}
}

func TestGetFileSize(t *testing.T) {
	assert.Equal(t, int64(len("hello world!")+1), GetFileSize("./test_logrotate/get-file-size-test.log"))
}

func TestPathMetadata(t *testing.T) {
	currDir, _ := filepath.Abs(".")

	testData := []struct {
		path     string
		pathInfo *PathMetadata
	}{
		{"./test.log", &PathMetadata{
			Path:      filepath.Join(currDir, "test.log"),
			BaseDir:   currDir,
			Basename:  "test",
			Extension: ".log",
		}},
		{"test_logrotate/test.log", &PathMetadata{
			Path:      filepath.Join(currDir, "test_logrotate", "test.log"),
			BaseDir:   filepath.Join(currDir, "test_logrotate"),
			Basename:  "test",
			Extension: ".log",
		}},
	}

	for i := range testData {
		pathInfo := NewPathMetadata(testData[i].path)
		ansPathInfo := testData[i].pathInfo

		assert.Equal(t, pathInfo, ansPathInfo)
	}
}

func TestGenerateLogFilename(t *testing.T) {
	baseDir := "/tmp"
	basename := "test"
	extension := ".log"

	testData := []struct {
		index int
		path  string
	}{
		{0, "/tmp/test.log"},
		{1, "/tmp/test-1.log"},
		{2, "/tmp/test-2.log"},
		{10000, "/tmp/test-10000.log"},
	}

	for i := range testData {
		logPath := GenerateLogFilename(testData[i].index, baseDir, basename, extension)
		assert.Equal(t, testData[i].path, logPath)
	}
}

func TestNewLogger(t *testing.T) {
	currDir, _ := filepath.Abs(".")

	lr, err := NewLogger("./test_logrotate/test.log", 1*1024*1024, 10, "\n")
	if err != nil {
		t.Fail()
	}

	// Check lr.PathInfo
	pathInfo := lr.PathInfo
	ansPathInfo := &PathMetadata{
		Path:      filepath.Join(currDir, "test_logrotate", "test.log"),
		BaseDir:   filepath.Join(currDir, "test_logrotate"),
		Basename:  "test",
		Extension: ".log",
	}

	assert.Equal(t, pathInfo, ansPathInfo)
	assert.True(t, IsFileExists(pathInfo.Path))
	assert.Equal(t, int64(0), GetFileSize(pathInfo.Path))

	// Check close
	err = lr.Close()
	assert.Nil(t, err)
}
