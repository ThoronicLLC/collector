package stdout

import (
	"bufio"
	"github.com/ThoronicLLC/collector/pkg/core"
	log "github.com/sirupsen/logrus"
	"os"
)

type Output struct{}

type stdoutOutput struct{}

func Handler() core.OutputHandler {
	return func(config []byte) core.Output {
		return &stdoutOutput{}
	}
}

func (s stdoutOutput) Write(inputFile string) (int, error) {
	// Open file
	fs, err := os.Open(inputFile)
	if err != nil {
		return 0, err
	}

	// Stat file
	info, err := fs.Stat()
	if err != nil {
		return 0, err
	}

	// Set max size
	var maxSize int
	scanner := bufio.NewScanner(fs)
	maxSize = int(info.Size())

	// Setup buffer
	buffer := make([]byte, 0, maxSize)

	// Setup count
	count := 0

	// Scan
	scanner.Buffer(buffer, maxSize)
	for scanner.Scan() {
		line := scanner.Text()
		log.Infof("%s", line)
		count++
	}

	return count, nil
}
