package core

type PipelineResults struct {
	FilePath    string
	ResultCount int
	State       State
	RetryCount  int
}
