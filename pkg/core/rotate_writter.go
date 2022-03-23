package core

import (
	"fmt"
	"github.com/google/uuid"
	"os"
	"strings"
	"sync"
	"time"
)

type RotateWriter struct {
	lock     sync.Mutex
	filename string // should be set to the actual filename
	fp       *os.File
	fileOpen bool
	size     int64
	maxSize  int64
}

// NewRotateWriter creates a new RotateWriter. Return nil if error occurs during setup.
func NewRotateWriter(filename string, maxSize int64) (*RotateWriter, error) {
	// Create writer
	writer := &RotateWriter{
		filename: filename,
		maxSize:  maxSize,
	}

	// Open new or existing
	err := writer.openExistingOrNew()
	if err != nil {
		return nil, err
	}

	return writer, nil
}

// Write implements io.Writer.
func (w *RotateWriter) Write(p []byte) (n int, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.write(p)
}

func (w *RotateWriter) WriteString(s string) (n int, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.write([]byte(s))
}

// Close implements io.Closer, and closes the current logfile.
func (w *RotateWriter) Close() error {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.close()
}

// openExistingOrNew opens the logfile if it exists.  If there is no such file, a new file is created.
func (w *RotateWriter) openExistingOrNew() error {
	fileSize := int64(0)
	if info, err := osStat(w.filename); err == nil {
		fileSize = info.Size()
	}

	file, err := os.OpenFile(w.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.fp = file
	w.size = fileSize
	w.fileOpen = true
	return nil
}

func (w *RotateWriter) write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	// Append newline to byte
	pStringWithNewline := strings.TrimSpace(string(p)) + "\n"
	pBytes := []byte(pStringWithNewline)
	pBytesLength := int64(len(pBytes))

	// Rotate file if max size is enabled
	if w.maxSize > 0 && ((w.size + pBytesLength) > w.maxSize) {
		err = w.rotate()
		if err != nil {
			return 0, fmt.Errorf("issue while rotating file: %s", err)
		}
	}

	// Write new line
	n, err = w.fp.Write(pBytes)
	w.size += int64(n)
	return n, err
}

// close will close the file if it is open.
func (w *RotateWriter) close() error {
	if w.fp == nil {
		w.fileOpen = false
		return nil
	}

	err := w.fp.Sync()
	if err != nil {
		return err
	}

	err = w.fp.Close()
	if err != nil {
		return err
	}

	w.fp = nil
	w.size = 0
	w.fileOpen = false
	return err
}

// Perform the actual act of rotating and reopening file.
func (w *RotateWriter) rotate() error {
	// Close existing file if open
	err := w.close()
	if err != nil {
		return err
	}

	// Rename dest file if it already exists
	_, err = os.Stat(w.filename)
	if err == nil {
		err = os.Rename(w.filename, fmt.Sprintf("%s.%d.%s", w.filename, time.Now().Unix(), uuid.New().String()))
		if err != nil {
			return err
		}
	}

	// Create a file.
	return w.openExistingOrNew()
}
