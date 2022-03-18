package file

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func glob(path string) []string {
	files, err := filepath.Glob(path)
	if err != nil {
		return make([]string, 0)
	}
	return files
}

func copyFromFilePosition(path string, position int64, writer io.Writer) (int64, error) {
	fs, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("issue opening file: %s", err)
	}

	defer fs.Close()

	fileStats, err := fs.Stat()
	if err != nil {
		return 0, fmt.Errorf("issue stating file: %s", err)
	}

	// If the file size is smaller than the current position, assume it has changed
	if fileStats.Size() == 0 {
		return 0, nil
	} else if fileStats.Size() < position {
		position = 0
	}

	// Jump to offset
	_, err = fs.Seek(position, 0)
	if err != nil {
		return 0, fmt.Errorf("issue seaking to position in file: %s", err)
	}

	// Read file
	scanner := bufio.NewScanner(fs)
	for scanner.Scan() {
		line := scanner.Bytes()
		_, err = writer.Write(line)
		if err != nil {
			return 0, fmt.Errorf("issue writting to file: %s", err)
		}
	}

	// Current position
	currentPosition, err := fs.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, fmt.Errorf("issue reading current position in file: %s", err)
	}

	return currentPosition, nil
}
