package cel

import (
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/core"
	log "github.com/sirupsen/logrus"
	"io"
	"strings"
)

type Action string

const (
	ActionAccept Action = "accept"
	ActionReject Action = "reject"
)

type celConfig struct {
	Rules  []string `json:"rules"`
	Action string   `json:"action"`
}

type celProcessor struct {
	config []byte
	logger *log.Entry
}

func Handler() core.ProcessHandler {
	return func(config []byte) core.Processor {
		return &celProcessor{
			config: config,
			logger: log.WithField("processor", "cel"),
		}
	}
}

func (processor *celProcessor) Process(inputFile string, writer io.Writer) error {
	// Unmarshal config
	var conf celConfig
	err := json.Unmarshal(processor.config, &conf)
	if err != nil {
		return fmt.Errorf("issue unmarshalling CEL config: %s", err)
	}

	// Validate config
	err = validate(conf)
	if err != nil {
		return fmt.Errorf("issue validating CEL config: %s", err)
	}

	// Set action
	currentAction := ActionAccept
	if conf.Action == string(ActionReject) {
		currentAction = ActionReject
	}

	// Use the file reader utility to pass our function
	err = core.FileReader(inputFile, func(s string) {
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
		result := ruleDetection(cleanLine, conf.Rules, processor.logger)

		// If the result was true and the action is accept, write log
		// If the result was false and the action is reject, write log
		if result && currentAction == ActionAccept {
			_, _ = writer.Write([]byte(s))
		} else if !result && currentAction == ActionReject {
			_, _ = writer.Write([]byte(s))
		}
	})
	if err != nil {
		return fmt.Errorf("issue reading file: %s", err)
	}

	return nil
}

func validate(config celConfig) error {
	// Action should only be "accept" or "reject"
	if config.Action != "accept" && config.Action != "reject" && config.Action != "" {
		return fmt.Errorf("invalid CEL action: %s", config.Action)
	}

	// Validate rules
	for _, v := range config.Rules {
		err := validateRule(v)
		if err != nil {
			return fmt.Errorf("invalid CEL rule: %s; error: %s", v, err)
		}
	}

	return nil
}
