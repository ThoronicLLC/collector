package file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/ThoronicLLC/collector/pkg/core/variable_replacer"
	"os"
	"path/filepath"
	"time"
)

var OutputName = "file"

type Config struct {
	Path    string `json:"path" validate:"required"`
	MaxSize int    `json:"max_size" validate:"min:1048576"` // We have to set a sane max size, so 1MB should work
}

type fileOutput struct {
	config Config
}

func Handler() core.OutputHandler {
	return func(config []byte) (core.Output, error) {
		// Set config defaults
		var conf Config

		// Unmarshal config
		err := json.Unmarshal(config, &conf)
		if err != nil {
			return nil, fmt.Errorf("issue unmarshalling file config: %s", err)
		}

		// Validate config
		err = core.ValidateStruct(&conf)
		if err != nil {
			return nil, err
		}

		return &fileOutput{
			config: conf,
		}, nil
	}
}

func (f *fileOutput) Write(inputFile string) (int, error) {
	// Setup local vars
	currentTime := time.Now()
	lineCount := 0

	// Setup path name
	path := variable_replacer.VariableReplacer(currentTime, f.config.Path)

	// Check if directory exists, if not, create
	pathParentDirectory := filepath.Dir(path)
	if !directoryExists(pathParentDirectory) {
		err := os.MkdirAll(pathParentDirectory, 0755)
		if err != nil {
			return 0, fmt.Errorf("issue creating parent directories for path: %s", path)
		}
	}

	// Open input file
	fs, err := os.Open(inputFile)
	if err != nil {
		return 0, fmt.Errorf("issue opening file: %s", err)
	}
	defer fs.Close()

	// Open output writer
	rotateWriter, err := core.NewRotateWriter(path, int64(f.config.MaxSize))
	if err != nil {
		return 0, fmt.Errorf("issue opening rotate writer: %s", err)
	}
	defer rotateWriter.Close()

	// Write to output
	scanner := bufio.NewScanner(fs)
	buffer := make([]byte, 0, core.MaxLogSize)
	scanner.Buffer(buffer, core.MaxLogSize)
	for scanner.Scan() {
		_, err = rotateWriter.Write(scanner.Bytes())
		if err != nil {
			return lineCount, fmt.Errorf("issue writing to file: %s", err)
		}
		lineCount++
	}

	return lineCount, nil
}

func directoryExists(dir string) bool {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
