package core

type PipelineResults struct {
	FilePath   string
	State      State
	RetryCount int
}
