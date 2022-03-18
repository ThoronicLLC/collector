package core

type InputHandler func(config []byte) (Input, error)

type Input interface {
	Run(errorHandler ErrorHandler, state State, processPipe chan<- PipelineResults)
	Stop()
}
