package core

type ProcessHandler func(config []byte) Processor

type Processor interface {
	Process(inputFile string) (string, error)
}
