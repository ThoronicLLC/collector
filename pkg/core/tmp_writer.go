package core

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

// ensure we always implement io.WriteCloser
var _ io.WriteCloser = (*TmpWriter)(nil)

type TmpWriter struct {
	size       int64
	file       *os.File
	fileOpen   bool
	mu         sync.Mutex
	WriteCount int

	millCh    chan bool
	startMill sync.Once
}

var (
	// os_Stat exists so it can be mocked out by tests.
	osStat = os.Stat
)

func NewTmpWriter() (*TmpWriter, error) {
	writer := &TmpWriter{}
	return writer, nil
}

// Write implements io.Writer.
func (l *TmpWriter) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.write(p)
}

func (l *TmpWriter) WriteString(s string) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.write([]byte(s))
}

func (l *TmpWriter) write(p []byte) (n int, err error) {
	if len(p) == 0 || string(p) == "" {
		return 0, nil
	}

	if l.file == nil {
		if err = l.openNew(); err != nil {
			return 0, err
		}
	}

	// Append newline to byte
	pStringWithNewline := strings.TrimSpace(string(p)) + "\n"

	// Write new line
	n, err = l.file.Write([]byte(pStringWithNewline))
	l.size += int64(n)
	l.WriteCount += 1

	return n, err
}

// Close implements io.Closer, and closes the current logfile.
func (l *TmpWriter) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.close()
}

// close syncs and closes the file if it is open.
func (l *TmpWriter) close() error {
	if l.file == nil {
		l.fileOpen = false
		return nil
	}

	err := l.file.Sync()
	if err != nil {
		return err
	}

	err = l.file.Close()
	if err != nil {
		return err
	}

	l.file = nil
	l.fileOpen = false
	return err
}

// Rotate causes TmpWriter to close the existing log file and immediately create a
// new one.
func (l *TmpWriter) Rotate() (int, string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Handle empty file
	if l.file == nil {
		return 0, "", nil
	}

	// Get file name and write count
	rotatedFileName := l.file.Name()
	rotatedWriteCount := l.WriteCount
	return rotatedWriteCount, rotatedFileName, l.rotate()
}

// Size returns the current file size
func (l *TmpWriter) Size() int64 {
	return l.size
}

// rotate closes the current file and opens a new file
func (l *TmpWriter) rotate() error {
	if err := l.close(); err != nil {
		return err
	}
	return nil
}

// openNew opens a new log file for writing. This methods assumes the last file has already been closed.
func (l *TmpWriter) openNew() error {
	f, err := ioutil.TempFile(os.TempDir(), randomStringWithLength(32))
	if err != nil {
		return fmt.Errorf("can't open new logfile: %s", err)
	}
	l.file = f
	l.size = 0
	l.fileOpen = true
	l.WriteCount = 0
	return nil
}

// filename generates the name of the logfile from the current time.
func (l *TmpWriter) filename() string {
	file, _ := ioutil.TempFile(os.TempDir(), randomStringWithLength(32))
	name := file.Name()
	_ = file.Close()
	return name
}

// Name returns current log file pointer
func (l *TmpWriter) Name() string {
	if l.file != nil {
		return l.file.Name()
	}

	return ""
}

// Exit closes the open file and cleans up any current or previous log files
func (l *TmpWriter) Exit() (err error) {
	if l.fileOpen {
		if err = l.close(); err != nil {
			return err
		}
	}

	return nil
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomStringWithLength(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
