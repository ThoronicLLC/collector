package main

import (
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/collector"
	"github.com/ThoronicLLC/collector/pkg/core"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	// TODO: In memory state map (change this to file state map)
	stateMap := NewStateMap()

	collectorConfig := collector.Config{
		SaveState:    defaultSaveStateFunc(stateMap),
		LoadState:    defaultLoadStateFunc(stateMap),
		ErrorHandler: defaultErrorHandler(),
	}

	// Setup collector
	c, err := collector.New(collectorConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Setup close handler
	setupCloseHandler(c)

	// Setup wait group
	var wg sync.WaitGroup

	// Load configs
	fileConfig := &core.Config{
		Input: core.PluginConfig{
			Name:     "file",
			Settings: []byte(`{"path":"/tmp/test2/*.log", "schedule":15}`),
		},
		Processors: []core.PluginConfig{
			{Name: "cel", Settings: []byte(`{"rules": ["has(event.hello)", "has(event.key)", "has(event.joe)"], "action": "reject"}`)},
		},
		Outputs: []core.PluginConfig{
			{Name: "stdout", Settings: nil},
		},
	}

	configs := make([]*core.Config, 0)
	configs = append(configs, fileConfig)

	// Setup context wait group go routine for closing application
	for i, v := range configs {
		currentConfig := v
		wg.Add(1)
		configCount := i
		go func() {
			defer wg.Done()
			err = c.Start(fmt.Sprintf("file_%d", configCount), currentConfig)
			if err != nil {
				log.Errorf("%s", err)
			}
		}()
	}

	go func() {
		for {
			<-time.After(30 * time.Second)
			status, err := c.Status("file_0")
			if err != nil {
				log.Errorf("unable to find status for instance %s: %s", "file", err)
				continue
			}
			if status != nil {
				statusBytes, _ := json.Marshal(status)
				log.Infof("current status: %s", string(statusBytes))
			}
		}
	}()

	go func() {
		for {
			<-time.After(45 * time.Second)
			state, ok := stateMap.Load("file_0")
			if ok {
				log.Infof("current state: %s", string(state))
			}
		}
	}()

	wg.Wait()
}

func defaultSaveStateFunc(stateMap *StateMap) core.SaveStateFunc {
	return func(id string, state core.State) error {
		stateMap.Store(id, state)
		return nil
	}
}

func defaultLoadStateFunc(stateMap *StateMap) core.LoadStateFunc {
	return func(id string) core.State {
		if val, ok := stateMap.Load(id); ok {
			return val
		}
		return nil
	}
}

func defaultErrorHandler() core.ErrorHandler {
	return func(critical bool, err error) {
		log.Errorf("%s", err)
	}
}

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS.
func setupCloseHandler(appCollector *collector.Collector) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		// Wait for first CTRL+C
		<-c
		log.Infof("gracefully shutting down... Send an additional CTRL+C for a forced shutdown")
		// Execute safe cancel function
		appCollector.StopAll()

		// Wait for additional CTRL+C for force closing
		for {
			select {
			case <-c:
				log.Warnf("additional CTRL+C received. Forced shutdown started...")
				os.Exit(0)
				return
			}
		}

	}()

	return
}
