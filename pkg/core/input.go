package core

type InputHandler func(config []byte) Input

type Input interface {
	Run(errorHandler ErrorHandler, state State, processPipe chan<- PipelineResults)
	Stop()
}
