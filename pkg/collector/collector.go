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
	runningInstances     map[string]*manager.Manager
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
		errorHandler: config.ErrorHandler,
		saveState:    config.SaveState,
		loadState:    config.LoadState,
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

	// Setup instance map
	c.runningInstances = make(map[string]*manager.Manager, 0)

	return c, nil
}

func (c *Collector) Start(id string, config *core.Config) error {
	// Check if an instance already exists
	if _, exists := c.runningInstances[id]; exists {
		return errors.New("instance with same ID already exists")
	}

	// Setup wait group
	var wg sync.WaitGroup
	wg.Add(1)

	// Run go routine
	go func() {
		defer wg.Done()

		// Setup input
		inputHandler := c.registeredInputs[config.Input.Name]
		input := inputHandler(config.Input.Settings)

		// Setup processors
		processors := make([]core.Processor, 0)

		// Setup outputs
		outputs := make([]core.Output, 0)
		for _, v := range config.Outputs {
			if outputHandler, exists := c.registeredOutputs[v.Name]; !exists {
				c.errorHandler(true, fmt.Errorf("invalid output type: %s", v.Name))
			} else {
				configuredOutput := outputHandler(v.Settings)
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
		c.runningInstances[id] = instance

		// Run instance with an instance manager
		instance.Run()
		log.Infof("closing instance with id: %s", id)
	}()

	// Blocking
	wg.Wait()

	// Delete instance from map
	delete(c.runningInstances, id)

	return nil
}

func (c *Collector) Stop(id string) error {
	// Check if an instance already exists
	if currentInstance, exists := c.runningInstances[id]; !exists {
		return errors.New("an instance with that ID does not exist")
	} else {
		currentInstance.Stop()
	}

	return nil
}

func (c *Collector) Status(id string) (*manager.Status, error) {
	// Check if an instance exists
	if currentInstance, exists := c.runningInstances[id]; !exists {
		return nil, errors.New("an instance with that ID does not exist")
	} else {
		return currentInstance.Status(), nil
	}
}

func (c *Collector) List() []string {
	instanceList := make([]string, 0)
	for k := range c.runningInstances {
		instanceList = append(instanceList, k)
	}
	return instanceList
}

func (c *Collector) ListStatus() []*manager.Status {
	instanceStatusList := make([]*manager.Status, 0)
	for _, v := range c.runningInstances {
		instanceStatusList = append(instanceStatusList, v.Status())
	}
	return instanceStatusList
}

func (c *Collector) StopAll() {
	for _, v := range c.runningInstances {
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
