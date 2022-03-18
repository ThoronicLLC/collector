package cel

import (
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/core"
	log "github.com/sirupsen/logrus"
	"io"
	"strings"
)

var ProcessorName = "cel"

type Config struct {
	Rules  []string `json:"rules" validate:"required"`
	Action string   `json:"action" validate:"oneof=accept reject"`
}

type celProcessor struct {
	config Config
	logger *log.Entry
}

func Handler() core.ProcessHandler {
	return func(config []byte) (core.Processor, error) {
		// Set config defaults
		conf := Config{
			Action: "accept",
		}

		// Unmarshal config
		err := json.Unmarshal(config, &conf)
		if err != nil {
			return nil, fmt.Errorf("issue unmarshalling CEL config: %s", err)
		}

		// Validate config
		err = core.ValidateStruct(&conf)
		if err != nil {
			return nil, err
		}

		return &celProcessor{
			config: conf,
			logger: log.WithField("processor", "cel"),
		}, nil
	}
}

func (processor *celProcessor) Process(inputFile string, writer io.Writer) error {
	// Use the file reader utility to pass our function
	err := core.FileReader(inputFile, func(s string) {
		// Clean line of any extra spaces for CEL detection
		cleanLine := strings.TrimSpace(s)

		// Return if clean line is empty
		if cleanLine == "" {
			if log.IsLevelEnabled(log.DebugLevel) {
				processor.logger.Debugf("line not valid json: %s", cleanLine)
			}
			return
		}

		// Return if line is not json
		if !json.Valid([]byte(cleanLine)) {
			if log.IsLevelEnabled(log.DebugLevel) {
				processor.logger.Errorf("line not valid json: %s", cleanLine)
			}
			return
		}

		// Run the rule detection with the configured rules
		result := ruleDetection(cleanLine, processor.config.Rules, processor.logger)

		// If the result was true and the action is accept, write log
		// If the result was false and the action is reject, write log
		if result && processor.config.Action == "accept" {
			_, _ = writer.Write([]byte(s))
		} else if !result && processor.config.Action == "reject" {
			_, _ = writer.Write([]byte(s))
		}
	})
	if err != nil {
		return fmt.Errorf("issue reading file: %s", err)
	}

	return nil
}
