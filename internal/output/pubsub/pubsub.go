package pubsub

import (
	"bufio"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/core"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"os"
	"strings"
)

var OutputName = "pubsub"

type Config struct {
	ProjectID       string          `json:"project_id" validate:"required"`
	TopicID         string          `json:"topic_id" validate:"required"`
	Credentials     json.RawMessage `json:"credentials,omitempty"`
	CredentialsPath string          `json:"credentials_path"`
}

type pubSubOutput struct {
	config Config
	ctx    context.Context
}

func Handler() core.OutputHandler {
	return func(config []byte) (core.Output, error) {
		// Set config defaults
		var conf Config

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

		// Validate credentials
		err = validateCredentialsOrPath(conf.Credentials, conf.CredentialsPath)
		if err != nil {
			return nil, err
		}

		return &pubSubOutput{
			config: conf,
			ctx:    context.Background(),
		}, nil
	}
}

func (p *pubSubOutput) Write(inputFile string) (int, error) {
	// Setup new client
	opts := make([]option.ClientOption, 0)
	if p.config.Credentials != nil && len(p.config.Credentials) > 0 && string(p.config.Credentials) != "null" {
		opts = append(opts, option.WithCredentialsJSON(p.config.Credentials))
	} else if p.config.CredentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(p.config.CredentialsPath))
	}

	// Setup PubSub client
	client, err := pubsub.NewClient(p.ctx, p.config.ProjectID, opts...)
	if err != nil {
		return 0, fmt.Errorf("issue setting up pub sub client: %s", err)
	}

	// Setup line variables
	lineCount := 0
	emptyLines := 0
	topicClient := client.Topic(p.config.TopicID)

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

		// Setup messages
		msg := &pubsub.Message{
			Data: []byte(trimmedLine),
		}

		// Setup subscription
		_, err = topicClient.Publish(p.ctx, msg).Get(p.ctx)
		if err != nil {
			log.Errorf("issue publishing line to pubsub: %s", err)
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

func validateCredentialsOrPath(credentials json.RawMessage, path string) error {
	if credentials != nil && len(credentials) > 0 && string(credentials) != "null" {
		return nil
	} else if path != "" {
		return nil
	}

	return fmt.Errorf("missing credentials")
}
