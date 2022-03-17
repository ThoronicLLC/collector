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
	size          int64
	file          *os.File
	fileOpen      bool
	mu            sync.Mutex
	previousFile  *os.File
	previousCount int
	WriteCount    int

	millCh    chan bool
	startMill sync.Once
}

var (
	// os_Stat exists so it can be mocked out by tests.
	osStat = os.Stat
)

func NewTmpWriter() (*TmpWriter, error) {
	writer := &TmpWriter{}
	err := writer.openNew()
	return writer, err
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
	if l.file == nil {
		if err = l.openExistingOrNew(); err != nil {
			return 0, err
		}
	}

	if len(p) == 0 {
		return 0, nil
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

// close closes the file if it is open.
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

	l.previousFile = l.file
	l.previousCount = l.WriteCount
	l.file = nil
	l.fileOpen = false
	return err
}

// Rotate causes TmpWriter to close the existing log file and immediately create a
// new one.
func (l *TmpWriter) Rotate() (int, string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Get file name and write count
	rotatedFileName := l.file.Name()
	rotatedWriteCount := l.WriteCount

	// Do not rotate if there is no writes to the file
	if rotatedWriteCount > 0 {
		return rotatedWriteCount, rotatedFileName, l.rotate()
	}
	return rotatedWriteCount, rotatedFileName, nil
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
	if err := l.openNew(); err != nil {
		return err
	}
	return nil
}

// openNew opens a new log file for writing. This methods assumes the last file has already been closed.
func (l *TmpWriter) openNew() error {
	// we use truncate here because this should only get called when we've moved
	// the file ourselves. if someone else creates the file in the meantime,
	// just wipe out the contents.
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

// openExistingOrNew opens the logfile if it exists.  If there is no such file, a new file is created.
func (l *TmpWriter) openExistingOrNew() error {
	filename := l.filename()
	info, err := osStat(filename)
	if os.IsNotExist(err) {
		return l.openNew()
	}
	if err != nil {
		return fmt.Errorf("error getting log file info: %s", err)
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// if we fail to open the old log file for some reason, just ignore
		// it and open a new log file.
		return l.openNew()
	}
	l.file = file
	l.size = info.Size()
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

// CurrentFile returns current log file pointer
func (l *TmpWriter) CurrentFile() *os.File {
	return l.file
}

// PreviousFile returns previous log file pointer
func (l *TmpWriter) PreviousFile() *os.File {
	return l.previousFile
}

// PreviousCount returns previous file write count
func (l *TmpWriter) PreviousCount() int {
	return l.previousCount
}

// DeleteCurrentFile deletes the current log file and update struct
func (l *TmpWriter) DeleteCurrentFile() (err error) {
	if l.file != nil && fileExists(l.file.Name()) {
		if err = os.Remove(l.file.Name()); err != nil {
			return err
		}
	}
	l.file = nil
	l.fileOpen = false
	return nil
}

// DeletePreviousFile deletes the previous log file and update struct
func (l *TmpWriter) DeletePreviousFile() (err error) {
	if l.previousFile != nil && fileExists(l.previousFile.Name()) {
		if err = os.Remove(l.previousFile.Name()); err != nil {
			return err
		}
	}
	l.previousFile = nil
	return nil
}

// Exit closes the open file and cleans up any current or previous log files
func (l *TmpWriter) Exit() (err error) {
	if l.fileOpen {
		if err = l.close(); err != nil {
			return err
		}
	}
	if l.file != nil {
		if err = l.DeleteCurrentFile(); err != nil {
			return err
		}
	}
	if l.previousFile != nil {
		if err = l.DeletePreviousFile(); err != nil {
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

// fileExists will return true if the file at the supplied path exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
