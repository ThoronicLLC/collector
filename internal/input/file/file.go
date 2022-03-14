package file

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/core"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type fileConfig struct {
	Path     string `json:"path"`
	Delete   bool   `json:"delete"`
	Schedule int    `json:"schedule"`
}

type fileInput struct {
	Config     []byte
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func Handler() core.InputHandler {
	return func(config []byte) core.Input {
		// Setup context
		ctx, cancelFn := context.WithCancel(context.Background())

		return &fileInput{
			Config:     config,
			ctx:        ctx,
			cancelFunc: cancelFn,
		}
	}
}

// Run will execute the input with the supplied context and state and return results
func (input *fileInput) Run(errorHandler core.ErrorHandler, state core.State, processPipe chan<- core.PipelineResults) {
	// Validate and Load config
	var conf fileConfig
	err := json.Unmarshal(input.Config, &conf)
	if err != nil {
		errorHandler(true, fmt.Errorf("issue unmarshalling file config: %s", err))
		return
	}

	// Validate and load state
	currentState := loadState(state)

	for {
		select {
		case <-input.ctx.Done():
			return
		case <-time.After(time.Duration(conf.Schedule) * time.Second):
			// Create temp file
			tmpFile, err := core.NewTmpWriter()
			if err != nil {
				errorHandler(false, fmt.Errorf("issue opening a new temp file writer: %s", err))
				continue
			}

			// Copy current state to a new object for modification
			newState := currentState

			// Collect (should be blocking for streaming)
			files := glob(conf.Path)
			for _, v := range files {
				log.Debugf("getting file: %s", v)

				// Get existing file position
				filePosition := getFilePastStatePosition(v, currentState)

				// Get results and offset
				offset, err := copyFromFilePosition(v, filePosition, tmpFile)
				if err != nil {
					errorHandler(false, err)
					continue
				}

				// If delete is enabled, remove the file. If not, update file state to keep track of data already processed
				if conf.Delete {
					err = os.Remove(v)
					if err != nil {
						errorHandler(false, err)
					}
				} else {
					newState = updateFileState(v, newState, offset)
				}
			}

			// Get results file name and size
			path := tmpFile.CurrentFile().Name()
			linesWritten := tmpFile.WriteCount
			err = tmpFile.Close()
			if err != nil {
				errorHandler(false, fmt.Errorf("issue closing file: %s", err))
				continue
			}

			// Marshal new state
			newStateBytes, err := json.Marshal(newState)
			if err != nil {
				errorHandler(false, fmt.Errorf("issue marshalling new state: %s", err))
				continue
			}

			// Setup pipeline results for next stage
			result := core.PipelineResults{
				FilePath:    path,
				ResultCount: linesWritten,
				State:       newStateBytes,
				RetryCount:  0,
			}

			// Pipe results to next stage
			processPipe <- result

			// Update current state to the new state since successful run
			currentState = newState
		}
	}
}

func (input *fileInput) Stop() {
	input.cancelFunc()
}
