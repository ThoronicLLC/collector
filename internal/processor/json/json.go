package json

import (
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/core"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"strings"
)

var ProcessorName = "json"

type Config struct {
	Add     []AddAction     `json:"add"`
	Remove  []RemoveAction  `json:"remove"`
	Replace []ReplaceAction `json:"replace"`
}

type AddAction struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RemoveAction struct {
	Key string `json:"key"`
}

type ReplaceAction struct {
	Key      string      `json:"key"`
	Value    interface{} `json:"value"`
	NewValue interface{} `json:"new_value"`
}

type jsonProcessor struct {
	config Config
	logger *log.Entry
}

func Handler() core.ProcessHandler {
	return func(config []byte) (core.Processor, error) {
		// Set config defaults
		conf := Config{}

		// Unmarshal config
		err := json.Unmarshal(config, &conf)
		if err != nil {
			return nil, fmt.Errorf("issue unmarshalling JSON config: %s", err)
		}

		// Validate config
		err = core.ValidateStruct(&conf)
		if err != nil {
			return nil, err
		}

		return &jsonProcessor{
			config: conf,
			logger: log.WithField("processor", "cel"),
		}, nil
	}
}

func (processor *jsonProcessor) Process(inputFile string, writer io.Writer) error {
	// Use the file reader utility to pass our function
	err := core.FileReader(inputFile, func(s string) {
		// Clean line of any extra spaces for CEL detection
		logLine := strings.TrimSpace(s)

		// Return if clean line is empty
		if logLine == "" {
			if log.IsLevelEnabled(log.DebugLevel) {
				processor.logger.Debugf("line not valid json: %s", logLine)
			}
			return
		}

		// Return if line is not json
		if !json.Valid([]byte(logLine)) {
			if log.IsLevelEnabled(log.DebugLevel) {
				processor.logger.Errorf("line not valid json: %s", logLine)
			}
			return
		}

		// Run add actions
		var err error
		for _, action := range processor.config.Add {
			logLine, err = sjson.Set(logLine, action.Key, action.Value)
			if err != nil {
				processor.logger.Errorf("issue running add action: %s", err)
				continue
			}
		}

		// Run remove actions
		for _, action := range processor.config.Remove {
			result := gjson.Get(logLine, action.Key)
			if result.Exists() {
				logLine, err = sjson.Delete(logLine, action.Key)
				if err != nil {
					processor.logger.Errorf("issue running remove action: %s", err)
					continue
				}
			}
		}

		// Run replace actions
		for _, action := range processor.config.Replace {
			result := gjson.Get(logLine, action.Key)
			if result.Exists() && result.Value() == action.Value {
				logLine, err = sjson.Set(logLine, action.Key, action.NewValue)
				if err != nil {
					processor.logger.Errorf("issue running remove action: %s", err)
					continue
				}
			}
		}

		// Write log line to output
		_, _ = writer.Write([]byte(logLine))
	})
	if err != nil {
		return fmt.Errorf("issue reading file: %s", err)
	}

	return nil
}
