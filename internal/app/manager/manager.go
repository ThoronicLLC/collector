package manager

import (
	"github.com/ThoronicLLC/collector/pkg/core"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

type Manager struct {
	status       *Status
	id           string
	config       core.Config
	input        core.Input
	processors   []core.Processor
	outputs      []core.Output
	saveState    core.SaveStateFunc
	loadState    core.LoadStateFunc
	errorHandler core.ErrorHandler
	processPipe  chan core.PipelineResults
	outputPipe   chan core.PipelineResults
}

type Config struct {
	ID           string
	Input        core.Input
	Processors   []core.Processor
	Outputs      []core.Output
	SaveState    core.SaveStateFunc
	LoadState    core.LoadStateFunc
	ErrorHandler core.ErrorHandler
}

type Status struct {
	Running                  bool
	Errors                   []error
	LastSuccessfulRun        time.Time
	HasErrors                bool
	ErrorsSinceSuccessfulRun int
}

func New(config Config) *Manager {
	// Setup initial manager
	mng := &Manager{
		status:      &Status{},
		id:          config.ID,
		input:       config.Input,
		processors:  config.Processors,
		outputs:     config.Outputs,
		saveState:   config.SaveState,
		loadState:   config.LoadState,
		processPipe: make(chan core.PipelineResults, 20),
		outputPipe:  make(chan core.PipelineResults, 20),
	}

	// Add a local error handler that also updated internal status
	var localErrorHandler core.ErrorHandler = func(critical bool, err error) {
		if err != nil {
			config.ErrorHandler(critical, err)
		}
		mng.statusHandler(err)
	}
	mng.errorHandler = localErrorHandler

	return mng
}

// Run should be run as a go routine as it blocks until the manager context is closed
func (manager *Manager) Run() {
	// Load state
	state := manager.loadState(manager.id)
	manager.status.Running = true
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		manager.input.Run(manager.errorHandler, state, manager.processPipe)
		close(manager.processPipe)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		manager.processHandler(manager.errorHandler)
		close(manager.outputPipe)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		manager.outputHandler()
	}()

	wg.Wait()

	manager.status.Running = false
}

func (manager *Manager) ID() string {
	return manager.id
}

func (manager *Manager) Status() *Status {
	return manager.status
}

func (manager *Manager) Stop() {
	manager.input.Stop()
}

func (manager *Manager) processHandler(errorHandler core.ErrorHandler) {
	for {
		res, ok := <-manager.processPipe
		if ok == false {
			log.Debugf("process handler pipeline closed for: %s", manager.id)
			break
		}

		// Setup current file tracker
		currentFile := res.FilePath
		var err error

		// Loop through and run processors
		for _, v := range manager.processors {
			// Set old results for future deleting
			oldResults := currentFile

			currentFile, err = v.Process(currentFile)
			if err != nil {
				errorHandler(false, err)
				break
			}

			// Delete old results after each process step
			err = os.Remove(oldResults)
			if err != nil {
				errorHandler(false, err)
			}
		}

		manager.outputPipe <- core.PipelineResults{
			FilePath: currentFile,
			State:    res.State,
		}
	}
}

func (manager *Manager) outputHandler() {
	for {
		res, ok := <-manager.outputPipe
		if !ok {
			log.Debugf("output handler pipeline closed for: %s", manager.id)
			break
		}

		// Setup current file tracker
		currentFile := res.FilePath

		// Send data to outputs
		outputError := false
		anyWritten := false
		var err error
		for _, v := range manager.outputs {
			// Start writing results to output
			_, err = v.Write(currentFile)
			if err != nil {
				outputError = true
				manager.errorHandler(false, err)
				continue
			}
			anyWritten = true
		}

		// If we have an output error and no output was written to, sleep and retry in a minute
		// else, at least one output was written to, we must be lossy
		//
		// Note: in order to guarantee at least once delivery, only one output should be set for each instance
		if outputError && !anyWritten && res.RetryCount < 3 {
			<-time.After(60 * time.Second)

			// Retry
			newRes := res
			newRes.RetryCount = res.RetryCount + 1
			manager.outputPipe <- newRes

			continue
		}

		// Delete old results
		err = os.Remove(currentFile)
		if err != nil {
			manager.errorHandler(false, err)
		}

		// Save state
		err = manager.saveState(manager.id, res.State)
		if err != nil {
			// This should never really happen. If we can't save state, we need to kill the instance or else we will get
			// duplicates
			manager.errorHandler(true, err)
			// TODO: Exit
			continue
		}

		// Update status
		manager.errorHandler(false, nil)
	}
}
