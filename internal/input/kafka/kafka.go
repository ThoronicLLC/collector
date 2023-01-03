package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/internal/integrations/kafka"
	"github.com/ThoronicLLC/collector/pkg/core"
	kafkago "github.com/segmentio/kafka-go"
	"sync"
	"time"
)

var InputName = "kafka"

type Config struct {
	Brokers        []string         `json:"brokers" validate:"required"`
	Topic          string           `json:"topic" validate:"required"`
	GroupID        string           `json:"group_id" validate:"required"`
	MinBytes       int              `json:"min_bytes"`
	MaxBytes       int              `json:"max_bytes"`
	AuthConfig     kafka.AuthConfig `json:"auth_config"`
	IncludeHeaders bool             `json:"include_headers"`
	FlushFrequency int              `json:"flush_frequency" validate:"required|min:0"`
}

type kafkaInput struct {
	config     Config
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func Handler() core.InputHandler {
	return func(config []byte) (core.Input, error) {
		// Set config defaults
		conf := Config{
			MinBytes:       10e2, // 1KB
			MaxBytes:       10e6, // 10MB
			FlushFrequency: 300,
			IncludeHeaders: false,
		}

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

		// Setup context
		ctx, cancelFn := context.WithCancel(context.Background())

		return &kafkaInput{
			config:     conf,
			ctx:        ctx,
			cancelFunc: cancelFn,
		}, nil
	}
}

func (k *kafkaInput) Run(errorHandler core.ErrorHandler, state core.State, processPipe chan<- core.PipelineResults) {
	// Setup local variables
	tmpWriter, err := core.NewTmpWriter()
	if err != nil {
		errorHandler(true, err)
		return
	}

	reader, err := kafka.NewReader(kafka.ReaderConfig{
		Ctx:        k.ctx,
		AuthConfig: k.config.AuthConfig,
		Brokers:    k.config.Brokers,
		Topic:      k.config.Topic,
		GroupID:    k.config.GroupID,
		MinBytes:   k.config.MinBytes,
		MaxBytes:   k.config.MaxBytes,
	})

	if err != nil {
		errorHandler(true, err)
		return
	}

	// Setup wait group
	var wg sync.WaitGroup
	flushCtx, flushCancelFn := context.WithCancel(k.ctx)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer flushCancelFn()
		for {
			m, err := reader.ReadMessage()
			if err != nil {
				if err == context.Canceled {
					return
				} else {
					errorHandler(false, fmt.Errorf("error reading message: %w", err))
					continue
				}
			}

			// Get message value
			messageValue := m.Value

			// If headers should be included, and the message is json, add them to the message
			if k.config.IncludeHeaders {
				newMessageValue, err := addHeadersToJsonMessages(m)
				if err != nil {
					errorHandler(false, fmt.Errorf("unable to add kafka headers to message: %w", err))
				} else {
					messageValue = newMessageValue
				}
			}

			_, writeErr := tmpWriter.Write(messageValue)
			if writeErr != nil {
				errorHandler(false, fmt.Errorf("error writing to tmp file: %w", writeErr))
			}
		}
	}()

	// Start timed process sync go routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-flushCtx.Done():
				return
			case <-time.After(time.Duration(k.config.FlushFrequency) * time.Second):
				err = flush(tmpWriter, processPipe)
				if err != nil {
					errorHandler(false, fmt.Errorf("issue flushing file: %s", err))
				}
			}
		}
	}()

	wg.Wait()

	// Close the reader
	err = reader.Close()
	if err != nil {
		errorHandler(false, fmt.Errorf("error closing reader: %w", err))
	}

	// Flush any remaining data to pipeline before return
	err = flush(tmpWriter, processPipe)
	if err != nil {
		errorHandler(false, fmt.Errorf("issue flushing file: %s", err))
	}
}

func (k *kafkaInput) Stop() {
	k.cancelFunc()
}

func flush(tmpFile *core.TmpWriter, processPipe chan<- core.PipelineResults) error {
	// Rotate the temp writer
	count, fileName, err := tmpFile.Rotate()
	if err != nil {
		return err
	}

	// Only send on if there are results
	processPipe <- core.PipelineResults{
		FilePath:    fileName,
		ResultCount: count,
		State:       nil,
		RetryCount:  0,
	}

	return nil
}

func addHeadersToJsonMessages(message kafkago.Message) ([]byte, error) {
	// Check if message is json
	var jsonMessage map[string]interface{}
	err := json.Unmarshal(message.Value, &jsonMessage)
	if err != nil {
		// Message is not json, so it will not add headers
		return nil, err
	}

	// Get headers
	var headers map[string]interface{}
	for _, header := range message.Headers {
		// Unmarshal header value from byte array to something that is JSON serializable
		var v interface{}
		if err := json.Unmarshal(header.Value, &v); err != nil {
			// Skip headers that are not JSON serializable
			continue
		}
		headers[header.Key] = v
	}

	// Set headers in json message
	jsonMessage["@headers"] = headers

	// Marshal json message
	newMessage, err := json.Marshal(jsonMessage)
	if err != nil {
		return nil, fmt.Errorf("error marshalling new json message: %w", err)
	}

	return newMessage, nil
}
