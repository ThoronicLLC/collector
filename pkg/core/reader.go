package core

import (
	"bufio"
	"fmt"
	"os"
)

type LineProcessor func(string)

func FileReader(filePath string, processFunc LineProcessor) error {
	// Open File
	fs, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("issue opening file: %s", err)
	}

	// Defer file close
	defer fs.Close()

	fileStats, err := fs.Stat()
	if err != nil {
		return fmt.Errorf("issue stating file: %s", err)
	}

	// Get max file length
	maxSize := int(fileStats.Size())

	// Read file
	scanner := bufio.NewScanner(fs)
	buffer := make([]byte, 0, maxSize)
	scanner.Buffer(buffer, maxSize)
	for scanner.Scan() {
		line := scanner.Text()
		processFunc(line)
	}

	return nil
}
