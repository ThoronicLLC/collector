package collector

import (
	"errors"
	"fmt"
	"github.com/ThoronicLLC/collector/internal/app"
	"github.com/ThoronicLLC/collector/internal/app/manager"
	"github.com/ThoronicLLC/collector/pkg/core"
	log "github.com/sirupsen/logrus"
	"sync"
)

type Collector struct {
	registeredInputs     map[string]core.InputHandler
	registeredProcessors map[string]core.ProcessHandler
	registeredOutputs    map[string]core.OutputHandler
	runningInstances     *instanceManagerMap
	errorHandler         core.ErrorHandler
	saveState            core.SaveStateFunc
	loadState            core.LoadStateFunc
}

type Config struct {
	SaveState    core.SaveStateFunc
	LoadState    core.LoadStateFunc
	ErrorHandler core.ErrorHandler
}

// New initializes a new collector instance with state management and error handling
func New(config Config) (*Collector, error) {
	// Register default
	c := &Collector{
		errorHandler:     config.ErrorHandler,
		saveState:        config.SaveState,
		loadState:        config.LoadState,
		runningInstances: NewInstanceManagerMap(),
	}

	// Register default inputs
	for k, v := range app.AddInternalInputs() {
		err := c.RegisterInput(k, v)
		if err != nil {
			return nil, err
		}
	}

	// Register default processors
	for k, v := range app.AddInternalProcessors() {
		err := c.RegisterProcessor(k, v)
		if err != nil {
			return nil, err
		}
	}

	// Register default outputs
	for k, v := range app.AddInternalOutputs() {
		err := c.RegisterOutput(k, v)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Collector) Start(id string, config core.Config) error {
	// Check if an instance already exists
	if _, exists := c.runningInstances.Get(id); exists {
		return errors.New("instance with same ID already exists")
	}

	// Debug log
	log.Debugf("starting collector: %s", id)

	// Setup wait group
	var wg sync.WaitGroup
	wg.Add(1)

	// Run go routine
	go func() {
		defer wg.Done()

		// Setup input
		inputHandler, exists := c.registeredInputs[config.Input.Name]
		if !exists {
			c.errorHandler(true, fmt.Errorf("invalid input type: %s", config.Input.Name))
			return
		}
		input, err := inputHandler(config.Input.Settings)
		if err != nil {
			c.errorHandler(true, fmt.Errorf("invalid input config: %s", err))
			return
		}

		// Setup processors
		processors := make([]core.Processor, 0)
		for _, v := range config.Processors {
			if processHandler, exists := c.registeredProcessors[v.Name]; !exists {
				c.errorHandler(true, fmt.Errorf("invalid processor type: %s", v.Name))
				return
			} else {
				configuredProcessor, err := processHandler(v.Settings)
				if err != nil {
					c.errorHandler(true, fmt.Errorf("invalid processor config: %s", err))
					return
				}
				processors = append(processors, configuredProcessor)
			}
		}

		// Setup outputs
		outputs := make([]core.Output, 0)
		for _, v := range config.Outputs {
			if outputHandler, exists := c.registeredOutputs[v.Name]; !exists {
				c.errorHandler(true, fmt.Errorf("invalid output type: %s", v.Name))
				return
			} else {
				configuredOutput, err := outputHandler(v.Settings)
				if err != nil {
					c.errorHandler(true, fmt.Errorf("invalid output config: %s", err))
					return
				}
				outputs = append(outputs, configuredOutput)
			}
		}

		managerConfig := manager.Config{
			ID:           id,
			Input:        input,
			Processors:   processors,
			Outputs:      outputs,
			SaveState:    c.saveState,
			LoadState:    c.loadState,
			ErrorHandler: c.errorHandler,
		}

		// Setup instance manager
		instance := manager.New(managerConfig)
		c.runningInstances.Set(id, instance)

		// Run instance with an instance manager
		instance.Run()
		log.Infof("gracefully stopped instance with id: %s", id)
	}()

	// Blocking
	wg.Wait()

	// Delete instance from map
	c.runningInstances.Delete(id)

	return nil
}

func (c *Collector) Stop(id string) error {
	// Check if an instance already exists
	if currentInstance, exists := c.runningInstances.Get(id); !exists {
		return fmt.Errorf("an instance with that ID does not exist")
	} else {
		currentInstance.Stop()
	}

	return nil
}

func (c *Collector) Status(id string) (*manager.Status, error) {
	// Check if an instance exists
	if currentInstance, exists := c.runningInstances.Get(id); !exists {
		return nil, fmt.Errorf("an instance with that ID does not exist")
	} else {
		return currentInstance.Status(), nil
	}
}

func (c *Collector) List() []string {
	instanceList := make([]string, 0)
	for _, v := range c.runningInstances.ListKeys() {
		instanceList = append(instanceList, v)
	}
	return instanceList
}

func (c *Collector) ListStatus() []*manager.Status {
	instanceStatusList := make([]*manager.Status, 0)
	for _, v := range c.runningInstances.List() {
		instanceStatusList = append(instanceStatusList, v.Status())
	}
	return instanceStatusList
}

func (c *Collector) StopAll() {
	for _, v := range c.runningInstances.List() {
		v.Stop()
	}
}

func (c *Collector) RegisterInput(name string, input core.InputHandler) error {
	if c.registeredInputs == nil {
		c.registeredInputs = make(map[string]core.InputHandler, 0)
	}

	if _, exists := c.registeredInputs[name]; exists {
		return fmt.Errorf("input with specified name already exists")
	}
	c.registeredInputs[name] = input
	return nil
}

func (c *Collector) RegisterProcessor(name string, processor core.ProcessHandler) error {
	if c.registeredProcessors == nil {
		c.registeredProcessors = make(map[string]core.ProcessHandler, 0)
	}

	if _, exists := c.registeredProcessors[name]; exists {
		return fmt.Errorf("processor with specified name already exists")
	}
	c.registeredProcessors[name] = processor
	return nil
}

func (c *Collector) RegisterOutput(name string, output core.OutputHandler) error {
	if c.registeredOutputs == nil {
		c.registeredOutputs = make(map[string]core.OutputHandler, 0)
	}

	if _, exists := c.registeredOutputs[name]; exists {
		return fmt.Errorf("output with specified name already exists")
	}
	c.registeredOutputs[name] = output
	return nil
}
