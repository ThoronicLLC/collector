package kv

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/jjeffery/kv"
	log "github.com/sirupsen/logrus"
)

var ProcessorName = "kv"

type Config struct {
	Type string `json:"type" validate:"in:raw,cef"`
}

type kvProcessor struct {
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
			return nil, fmt.Errorf("issue unmarshalling CEL config: %s", err)
		}

		// Validate config
		err = core.ValidateStruct(&conf)
		if err != nil {
			return nil, err
		}

		return &kvProcessor{
			config: conf,
			logger: log.WithField("processor", "syslog"),
		}, nil
	}
}

func (processor *kvProcessor) Process(inputFile string, writer io.Writer) error {
	// Use the file reader utility to pass our function
	err := core.FileReader(inputFile, func(s string) {
		// Clean line of any extra spaces for CEL detection
		cleanLine := strings.TrimSpace(s)

		// Return if clean line is empty
		if cleanLine == "" {
			if log.IsLevelEnabled(log.DebugLevel) {
				processor.logger.Debugf("line is empty: %s", cleanLine)
			}
			return
		}

		switch processor.config.Type {
		case "raw":
			msg, err := parseKV(cleanLine)
			if err != nil {
				processor.logger.Errorf("issue parsing line: %s", err)
			} else {
				_, _ = writer.Write(msg)
			}
		case "cef":
			msg, err := parseCef(cleanLine)
			if err != nil {
				processor.logger.Errorf("issue parsing line: %s", err)
			} else {
				_, _ = writer.Write(msg)
			}
		}

	})

	return err
}

// parseKeyValue will take a key value formatted string and convert it into a key value map
func parseKeyValue(event string, cef bool) (map[string]string, error) {
	// Clear out empty key values
	reg := regexp.MustCompile("[a-zA-Z0-9]+=[ ]")
	newEvent := reg.ReplaceAllString(event, " ")

	// fix ending
	if newEvent[len(newEvent)-1] == '=' {
		reg := regexp.MustCompile("[ ][a-zA-Z0-9]+=$")
		newEvent = reg.ReplaceAllString(newEvent, "")
	}

	// Use KeyValue library to parse
	text, list := kv.Parse([]byte(newEvent))

	// If return text then an error occurred during parsing
	if len(text) > 0 {
		return nil, fmt.Errorf(`invalid key value format at: "%s"`, string(text))
	}

	// Convert from list to a map
	elementMap := make(map[string]string)
	for i := 0; i < len(list); i += 2 {
		key := list[i].(string)
		value := list[i+1].(string)
		if cef {
			elementMap[cefEscapeExtension(key)] = cefEscapeExtension(value)
		} else {
			elementMap[key] = value
		}
	}

	return elementMap, nil
}

// ConstructKeyValue will take a key value formatted string and convert it into a key value json object
func parseKV(event string) ([]byte, error) {
	// Parse key value string
	result, err := parseKeyValue(event, false)

	// Handle errors
	if err != nil {
		return nil, err
	}

	// Marshal JSON string
	jsonString, err := json.Marshal(result)

	// Handle errors
	if err != nil {
		return nil, err
	}

	return jsonString, nil
}
