package core

type OutputHandler func(config []byte) (Output, error)

type Output interface {
	Write(inputFile string) (int, error)
}
