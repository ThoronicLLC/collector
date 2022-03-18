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
	ProjectID       string          `json:"project_id"`
	TopicID         string          `json:"topic_id"`
	Credentials     json.RawMessage `json:"credentials,omitempty"`
	CredentialsPath string          `json:"credentials_path"`
}

type pubSubOutput struct {
	Config []byte
	ctx    context.Context
}

func Handler() core.OutputHandler {
	return func(config []byte) core.Output {
		return &pubSubOutput{
			Config: config,
			ctx:    context.Background(),
		}
	}
}

func (p *pubSubOutput) Write(inputFile string) (int, error) {
	// Serialize config
	var conf Config
	err := json.Unmarshal(p.Config, &conf)
	if err != nil {
		return 0, fmt.Errorf("issue unmarshalling config: %s", err)
	}

	// Setup new client
	opts := make([]option.ClientOption, 0)
	if conf.Credentials != nil && len(conf.Credentials) > 0 && string(conf.Credentials) != "null" {
		opts = append(opts, option.WithCredentialsJSON(conf.Credentials))
	} else if conf.CredentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(conf.CredentialsPath))
	}

	// Setup PubSub client
	client, err := pubsub.NewClient(p.ctx, conf.ProjectID, opts...)
	if err != nil {
		return 0, fmt.Errorf("issue setting up pub sub client: %s", err)
	}

	// Setup line variables
	lineCount := 0
	emptyLines := 0
	topicClient := client.Topic(conf.TopicID)

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
