package kafka

import (
  "bufio"
  "context"
  "encoding/json"
  "fmt"
  "github.com/ThoronicLLC/collector/internal/integrations/kafka"
  "github.com/ThoronicLLC/collector/pkg/core"
  log "github.com/sirupsen/logrus"
  "os"
  "strings"
)

var OutputName = "kafka"

type Config struct {
  Brokers    []string         `json:"brokers" validate:"required|minLen:1"`
  Topic      string           `json:"topic" validate:"required"`
  AuthConfig kafka.AuthConfig `json:"auth_config"`
}

type kafkaOutput struct {
  config Config
  ctx    context.Context
}

func Handler() core.OutputHandler {
  return func(config []byte) (core.Output, error) {
    // Set config defaults
    conf := Config{}

    // Unmarshal config
    err := json.Unmarshal(config, &conf)
    if err != nil {
      return nil, fmt.Errorf("issue unmarshalling file config: %s", err)
    }

    // Validate config
    err = core.ValidateStruct(&conf)
    if err != nil {
      return nil, err
    }

    ctx := context.Background()

    return &kafkaOutput{
      config: conf,
      ctx:    ctx,
    }, nil
  }
}

func (p *kafkaOutput) Write(inputFile string) (int, error) {
  // Set up line variables
  lineCount := 0
  emptyLines := 0

  // Set up writer
  writer, err := kafka.NewWriter(kafka.WriterConfig{
    Ctx:        p.ctx,
    AuthConfig: p.config.AuthConfig,
    Brokers:    p.config.Brokers,
    Topic:      p.config.Topic,
  })
  if err != nil {
    return 0, fmt.Errorf("issue setting up kafka writer: %s", err)
  }
  defer writer.Close()

  // Open file
  file, err := os.Open(inputFile)
  if err != nil {
    return 0, fmt.Errorf("issue opening input file: %s", err)
  }
  defer file.Close()

  // Start reading from the file with a reader.
  scanner := bufio.NewScanner(file)
  buffer := make([]byte, 0, core.MaxLogSize)
  scanner.Buffer(buffer, core.MaxLogSize)
  for scanner.Scan() {
    line := scanner.Text()
    trimmedLine := strings.TrimSpace(line)

    // Check empty lines
    if trimmedLine == "" {
      emptyLines++
      continue
    }

    // Write to kafka topic
    err = writer.WriteMessage([]byte(trimmedLine))
    if err != nil {
      log.Errorf("issue publishing line to kafka: %s", err)
      continue
    }

    // Debug print with empty line count
    if emptyLines > 0 {
      log.Debugf("ignored %d empty log entries", emptyLines)
    }

    lineCount++
  }

  return lineCount, nil
}
