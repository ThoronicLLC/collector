package bfp

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
)

var DefaultMaxSize = 1 * 1024 * 1024 // 1MB

type BufferedProcessor struct {
	maxSize int
	logger  *slog.Logger
}

func New(options ...Option) *BufferedProcessor {
	br := &BufferedProcessor{
		maxSize: DefaultMaxSize,
		logger:  slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("pkg", "bfp"),
	}

	for _, opts := range options {
		br = opts(br)
	}

	return br
}

type Option func(reader *BufferedProcessor) *BufferedProcessor

// WithLogger will configure the logger for the reader
func WithLogger(logger *slog.Logger) Option {
	return func(reader *BufferedProcessor) *BufferedProcessor {
		reader.logger = logger
		return reader
	}
}

// WithMaxSize will configure max size in bytes for the HTTP requests
func WithMaxSize(maxSize int) Option {
	return func(reader *BufferedProcessor) *BufferedProcessor {
		reader.maxSize = maxSize
		return reader
	}
}

func (b *BufferedProcessor) Process(name string, process func(bufferedFile string) error) (n int, err error) {
	// Open the input file for reading
	inputFile, err := os.Open(name)
	if err != nil {
		return 0, fmt.Errorf("error opening input file: %w", err)
	}
	defer inputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	var currentSize int
	var tempFile *os.File
	var writer *bufio.Writer

	// Function to create a new temporary file
	createTempFile := func() error {
		if tempFile != nil {
			if err := writer.Flush(); err != nil {
				return fmt.Errorf("error flushing to temporary file: %w", err)
			}
			tempName := tempFile.Name()
			if currentSize > 0 {
				if err := process(tempName); err != nil {
					return fmt.Errorf("error processing temporary file: %w", err)
				}
				n++
			}
			if err := tempFile.Close(); err != nil {
				return fmt.Errorf("error closing temporary file: %w", err)
			}
			if err := os.Remove(tempName); err != nil {
				return fmt.Errorf("error removing temporary file: %w", err)
			}
		}

		tempFile, err = os.CreateTemp("", "tempfile_*.txt")
		if err != nil {
			return fmt.Errorf("error creating temporary file: %w", err)
		}
		writer = bufio.NewWriter(tempFile)
		currentSize = 0
		return nil
	}

	// Create the first temporary file
	if err := createTempFile(); err != nil {
		return n, err
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		line = line + "\n"
		lineSize := len(line)

		// If the line does not fit in the current temporary file, create a new one
		if currentSize+lineSize > b.maxSize {
			if err := createTempFile(); err != nil {
				return n, err
			}
		}

		// Write the line to the current temporary file
		if _, err := writer.WriteString(line); err != nil {
			return n, fmt.Errorf("error writing to temporary file: %w", err)
		}
		currentSize += lineSize
	}

	// Check for any errors that occurred during the scan
	if err := scanner.Err(); err != nil {
		return n, fmt.Errorf("error scanning input file: %w", err)
	}

	// Flush and process the last temporary file
	if err := writer.Flush(); err != nil {
		return n, fmt.Errorf("error flushing to temporary file: %w", err)
	}

	tempName := tempFile.Name()

	// Process if there is any data left over
	if currentSize > 0 {
		if err := process(tempName); err != nil {
			return n, fmt.Errorf("error processing temporary file: %w", err)
		}
		n++
	}
	if err := tempFile.Close(); err != nil {
		return n, fmt.Errorf("error closing temporary file: %w", err)
	}
	if err := os.Remove(tempName); err != nil {
		return n, fmt.Errorf("error removing temporary file: %w", err)
	}

	return n, nil
}
