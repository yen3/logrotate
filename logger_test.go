package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	assert.Equal(t, int64(0), GetFileSize("./test_logrotate/test-empty.log"))
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

func TestNewLoggerEmptyLogFile(t *testing.T) {
	currDir, _ := filepath.Abs(".")

	lr, err := NewLogger("./test_logrotate/test.log", 1*1024*1024, 10)
	assert.Nil(t, err)

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

	// Remove the test file
	os.Remove(pathInfo.Path)
}

func TestNewLoggerExistingFile(t *testing.T) {
	// Open
	lr, err := NewLogger("./test_logrotate/test-existing.log", 1*1024*1024, 10)
	assert.Nil(t, err)

	assert.True(t, IsFileExists(lr.PathInfo.Path))
	assert.Equal(t, int64(0), GetFileSize(lr.PathInfo.Path))

	// Try to write log
	msg := "hello world!\n"
	lr.Write([]byte(msg))

	// Close
	err = lr.Close()
	assert.Nil(t, err)

	// Check write
	assert.Equal(t, int64(len(msg)), GetFileSize(lr.PathInfo.Path))
}

func TestNewLoggerExistingFileRotate(t *testing.T) {
	// Open
	lr, err := NewLogger("./test_logrotate/test-existing-rotate.log", 20, 10)
	assert.Nil(t, err)

	assert.True(t, IsFileExists(lr.PathInfo.Path))
	assert.Equal(t, int64(0), GetFileSize(lr.PathInfo.Path))

	// Try to write log
	msg := "123456789\n"

	// First write, it will rotate in next write
	lr.Write([]byte(msg))
	lr.Write([]byte(msg))
	// Write to file then rotate
	lr.Write([]byte(msg))

	// Write to second file
	lr.Write([]byte(msg))

	// Close
	err = lr.Close()
	assert.Nil(t, err)

	// Check file
	assert.Equal(t, int64(len(msg)), GetFileSize(lr.PathInfo.Path))
	assert.Equal(t, int64(len(msg)*3), GetFileSize("./test_logrotate/test-existing-rotate-1.log"))
}

func TestLogRotate(t *testing.T) {
	// Open
	lr, err := NewLogger("./test_logrotate/test-rotate.log", 20, 10)
	assert.Nil(t, err)

	// Try to write log
	msg := "1234\n5678\n"

	// First write, it will rotate in next write
	lr.Write([]byte(msg))
	lr.Write([]byte(msg))
	// Write to file then rotate
	lr.Write([]byte(msg))

	// Write to second file
	lr.Write([]byte(msg))

	// Close
	err = lr.Close()
	assert.Nil(t, err)

	// Check file content
	assert.Equal(t, int64(15), GetFileSize(lr.PathInfo.Path))
	assert.Equal(t, int64(25), GetFileSize("./test_logrotate/test-rotate-1.log"))
}

func TestLogRotateNotFindSep(t *testing.T) {
	// Open
	lr, err := NewLogger("./test_logrotate/test-rotate-nosep.log", 20, 10)
	assert.Nil(t, err)

	// Try to write log
	msg := "1234\n5678\n"

	// First write, it will rotate in next write
	lr.Write([]byte(msg))
	lr.Write([]byte(msg))
	// Write to file then rotate
	lr.Write([]byte("12345678"))

	// Write to second file
	lr.Write([]byte(msg))

	// Close
	err = lr.Close()
	assert.Nil(t, err)

	//// Check file content
	assert.Equal(t, int64(5), GetFileSize(lr.PathInfo.Path))
	assert.Equal(t, int64(33), GetFileSize("./test_logrotate/test-rotate-nosep-1.log"))
}

func TestLogRotateReplace(t *testing.T) {
	// Open
	lr, err := NewLogger("./test_logrotate/test-rotate-replace.log", 20, 2)
	assert.Nil(t, err)

	msg := "ABCDEFGH"

	// When write 7 logs. Since each file can contain 3 entry, the writing
	// action would generate three files. The 1, 2 and 3 log would be
	// elimimate, The first log file would contain log 7, and the second log
	// file would contain log 4, 5 and 6.
	for i := 1; i <= 7; i++ {
		lr.Write([]byte(fmt.Sprintf("%s%d\n", msg, i)))
	}

	// Check file log
	raw_log, err := ioutil.ReadFile("./test_logrotate/test-rotate-replace.log")
	assert.Nil(t, err)

	raw_log_2, err := ioutil.ReadFile("./test_logrotate/test-rotate-replace-1.log")
	assert.Nil(t, err)

	first_log := string(raw_log)
	second_log := string(raw_log_2)

	assert.Equal(t, first_log, "ABCDEFGH7\n")
	assert.Equal(t, second_log, "ABCDEFGH4\nABCDEFGH5\nABCDEFGH6\n")
}

func TestNewLoggerRemoveBigLogFile(t *testing.T) {
	// Create test log file
	path := "./test_logrotate/test-big.log"

	data := make([]byte, 101)
	for i := range data {
		data[i] = byte('a')
	}
	data = append(data, byte('\n'))

	err := ioutil.WriteFile(path, data, 0644)
	assert.Nil(t, err)

	// Open
	lr, err := NewLogger(path, 10, 2)
	assert.Nil(t, err)
	lr.Close()

	// Check file size
	assert.Equal(t, int64(0), GetFileSize(path))
}

func TestNewLoggerRotatePreLogFile(t *testing.T) {
	// Create test log file
	path := "./test_logrotate/test-rotate-first.log"

	data := make([]byte, 20)
	for i := range data {
		data[i] = byte('a')
	}
	data = append(data, byte('\n'))

	err := ioutil.WriteFile(path, data, 0644)
	assert.Nil(t, err)

	// Open
	lr, err := NewLogger(path, 20, 2)
	assert.Nil(t, err)
	lr.Close()

	// Check file size
	assert.Equal(t, int64(0), GetFileSize(path))
	assert.Equal(t, int64(21), GetFileSize("./test_logrotate/test-rotate-first-1.log"))
}

func TestNewLoggerAppendExistedFile(t *testing.T) {
	// Create test log file
	path := "./test_logrotate/test-append.log"
	data := make([]byte, 5)
	for i := range data {
		data[i] = byte('a')
	}
	data = append(data, byte('\n'))

	err := ioutil.WriteFile(path, data, 0644)
	assert.Nil(t, err)

	// Open
	lr, err := NewLogger(path, 20, 2)
	assert.Nil(t, err)

	// Write file
	lr.Write([]byte("Hello World!\n"))

	lr.Close()

	// Check file content
	raw_log, err := ioutil.ReadFile(path)
	assert.Nil(t, err)
	logContent := string(raw_log)

	assert.Equal(t, int64(19), GetFileSize(path))
	assert.Equal(t, "aaaaa\nHello World!\n", logContent)
}

func WriteToRotateFile(logReader *os.File, lr *File, writeFinished chan bool, t *testing.T) {
	for {
		buf := make([]byte, 128)

		readSize, err := logReader.Read(buf)
		if err != nil {
			lr.Close()

			// The write pipe is closed
			if err == io.EOF {
				logReader.Close()
				lr.Close()
				writeFinished <- true
				return
			}

			// Should not happen
			t.Fail()
			writeFinished <- false
			return
		}

		buf = buf[0:readSize]
		lr.Write(buf)
	}
}

func TestStarProcessLog(t *testing.T) {
	lr, err := NewLogger("./test_logrotate/test-subprocess.log", 20, 10)
	assert.Nil(t, err)

	// Create /dev/null file
	var devnull *os.File
	if devnull == nil {
		devnull, err = os.OpenFile(os.DevNull, os.O_RDWR, 0755)
		if err != nil {
			panic(err)
		}
	}

	logReader, logWriter, err := os.Pipe()
	assert.Nil(t, err)

	var procAttr os.ProcAttr
	procAttr.Files = []*os.File{devnull, logWriter, logWriter}
	writeFinished := make(chan bool)
	go WriteToRotateFile(logReader, lr, writeFinished, t)

	args := []string{"bash", "-c", "for i in $(seq 20); do echo \"Hello World$i\"; done"}
	args[0], err = exec.LookPath(args[0])
	assert.Nil(t, err)

	p, err := os.StartProcess(args[0], args, &procAttr)
	assert.Nil(t, err)

	// Wait for the process
	_, err = p.Wait()
	assert.Nil(t, err)

	// Close the pipe
	logWriter.Close()

	// Wait until the reading is finished, and avoid deadlock
	select {
	case writeState := <-writeFinished:
		assert.True(t, writeState)
	case <-time.After(100 * time.Second):
		fmt.Println("timeout 10")
	}

	// Check the file content
	files := []string{
		"./test_logrotate/test-subprocess-3.log",
		"./test_logrotate/test-subprocess-2.log",
		"./test_logrotate/test-subprocess-1.log",
		"./test_logrotate/test-subprocess.log",
	}
	for i := 1; i < len(files); i++ {
		// test-subprocess-3.log may not generate in some situation.
		assert.True(t, IsFileExists(files[i]))
	}

	var data string
	for i := range files {
		if IsFileExists(files[i]) {
			raw_data, err := ioutil.ReadFile(files[i])
			assert.Nil(t, err)
			data = data + string(raw_data)
		}
	}

	var ansData string
	for i := 1; i <= 20; i++ {
		ansData = ansData + fmt.Sprintf("Hello World%d\n", i)
	}

	assert.True(t, strings.Contains(ansData, data))
	assert.True(t, strings.Contains(data, "Hello World20\n"))
}

//func TestSpawnProcess(t *testing.T) {
//args := []string{"bash", "-c", "for i in $(seq 20); do echo \"Hello World$i\"; done"}
//err := SpawnProcess(args, "./test_logrotate/test-spawn-process.log", 20, 10)

//assert.Nil(t, err)
//}
