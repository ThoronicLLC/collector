package cli

import (
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/collector"
	"github.com/ThoronicLLC/collector/pkg/core"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

func Run(configPath string) error {
	// State manager
	stateManager := NewFileState(configPath)

	// Instance configs
	instanceConfigs := make(map[string]core.Config, 0)

	// Load config files
	files, err := filepath.Glob(filepath.Join(configPath, "*.conf"))
	if err != nil {
		return fmt.Errorf("issue reading config directory: %s", err)
	}

	// Load all the required configs
	for _, v := range files {
		// Read the file
		fileData, err := ioutil.ReadFile(v)
		if err != nil {
			log.Errorf("issue reading file: %s", err)
			continue
		}

		// Marshal config
		var tmpCfg core.Config
		err = json.Unmarshal(fileData, &tmpCfg)
		if err != nil {
			log.Errorf("invalid config file: %s", v)
			continue
		}

		// Get just the filename as the config
		id := strings.TrimPrefix(strings.Replace(v, configPath, "", 1), "/")
		instanceConfigs[id] = tmpCfg
	}

	// Initialize the collector config
	collectorConfig := collector.Config{
		SaveState:    defaultSaveStateFunc(stateManager),
		LoadState:    defaultLoadStateFunc(stateManager),
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

	// Setup context wait group go routine for closing application
	for k, v := range instanceConfigs {
		instanceID := k
		currentConfig := v
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = c.Start(instanceID, currentConfig)
			if err != nil {
				log.Errorf("%s", err)
			}
		}()
	}

	wg.Wait()

	return nil
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
		fmt.Println("")
		log.Infof("gracefully shutting down... Send an additional CTRL+C for a forced shutdown")
		// Execute safe cancel function
		appCollector.StopAll()

		// Wait for additional CTRL+C for force closing
		for {
			select {
			case <-c:
				fmt.Println("")
				log.Warnf("additional CTRL+C received. Forced shutdown started...")
				os.Exit(0)
				return
			}
		}

	}()

	return
}
