package core

import "io"

type ProcessHandler func(config []byte) Processor

type Processor interface {
	Process(inputFile string, writer io.Writer) error
}
