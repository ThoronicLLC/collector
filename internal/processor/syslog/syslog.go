package syslog

import (
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/influxdata/go-syslog/v3/rfc3164"
	"github.com/influxdata/go-syslog/v3/rfc5424"
	log "github.com/sirupsen/logrus"
	"io"
	"regexp"
	"strings"
)

var ProcessorName = "syslog"

type Config struct {
	Type string `json:"type" validate:"in:raw,rfc5424,rfc3164"`
}

type syslogProcessor struct {
	config Config
	logger *log.Entry
}

func Handler() core.ProcessHandler {
	return func(config []byte) (core.Processor, error) {
		// Set config defaults
		conf := Config{
			Type: "raw",
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

		return &syslogProcessor{
			config: conf,
			logger: log.WithField("processor", "syslog"),
		}, nil
	}
}

func (processor *syslogProcessor) Process(inputFile string, writer io.Writer) error {
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

		// Process
		switch processor.config.Type {
		case "raw":
			m, err := parseRaw(cleanLine)
			if err != nil {
				processor.logger.Errorf("issue parsing line: %s", err)
			} else {
				_, _ = writer.Write([]byte(m))
			}
		case "rfc5424":
			m, err := parseRfc5424(cleanLine)
			if err != nil {
				processor.logger.Errorf("issue parsing line: %s", err)
			} else {
				_, _ = writer.Write([]byte(m))
			}
		case "rfc3164":
			m, err := parseRfc3164(cleanLine)
			if err != nil {
				processor.logger.Errorf("issue parsing line: %s", err)
			} else {
				_, _ = writer.Write([]byte(m))
			}
		}
	})

	return err
}

func parseRaw(message string) (string, error) {
	r := regexp.MustCompile(`^<([0-9]+)>`)
	newMessage := r.ReplaceAllString(message, "")
	return newMessage, nil
}

func parseRfc5424(message string) (string, error) {
	parser := rfc5424.NewParser()
	m, err := parser.Parse([]byte(message))
	if err != nil {
		return "", err
	}

	// Check if valid
	if !m.Valid() {
		return "", fmt.Errorf("invalid RFC5424 message")
	}

	// Cast message
	syslogMessage, ok := m.(*rfc5424.SyslogMessage)
	if !ok {
		return "", fmt.Errorf("unable to cast RFC5424 message")
	}

	return derefString(syslogMessage.Message), nil
}

func parseRfc3164(message string) (string, error) {
	parser := rfc3164.NewParser()
	m, err := parser.Parse([]byte(message))
	if err != nil {
		return "", err
	}

	// Check if valid
	if !m.Valid() {
		return "", fmt.Errorf("invalid RFC3164 message")
	}

	// Cast message
	syslogMessage, ok := m.(*rfc3164.SyslogMessage)
	if !ok {
		return "", fmt.Errorf("unable to cast RFC5424 message")
	}

	return derefString(syslogMessage.Message), nil
}

func derefString(s *string) string {
	if s != nil {
		return *s
	}

	return ""
}
