package main

import (
	"path/filepath"
	"testing"
)

func TestIsFileExists(t *testing.T) {
	testData := []struct {
		path    string
		isExist bool
	}{
		{"./test_logrotate/test.log", true},
		{"./test_logrotate/test-1.log", true},
		{"./test_logrotate/doesnot_exist.log", false},
	}

	for i := range testData {
		if IsFileExists(testData[i].path) != testData[i].isExist {
			t.Fail()
		}
	}
}

func TestGetFileSize(t *testing.T) {
	if GetFileSize("./test_logrotate/get-file-size-test.log") == int64(len("hello world!")) {
		t.Fail()
	}
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

		if pathInfo.Path != ansPathInfo.Path {
			t.Fail()
		}

		if pathInfo.BaseDir != ansPathInfo.BaseDir {
			t.Fail()
		}

		if pathInfo.Basename != ansPathInfo.Basename {
			t.Fail()
		}

		if pathInfo.Extension != ansPathInfo.Extension {
			t.Fail()
		}
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
		if logPath != testData[i].path {
			t.Fail()
		}
	}
}
