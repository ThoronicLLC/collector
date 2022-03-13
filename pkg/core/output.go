package core

type OutputHandler func(config []byte) Output

type Output interface {
	Write(inputFile string) (int, error)
}
