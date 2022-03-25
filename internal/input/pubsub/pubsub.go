package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/core"
	"google.golang.org/api/option"
	"sync"
	"time"
)

var InputName = "pubsub"

type Config struct {
	ProjectID       string          `json:"project_id" validate:"required"`
	SubscriptionID  string          `json:"subscription_id" validate:"required"`
	Credentials     json.RawMessage `json:"credentials,omitempty"`
	CredentialsPath string          `json:"credentials_path"`
	FlushFrequency  int             `json:"flush_frequency" validate:"required|min:0"`
}

type pubSubInput struct {
	config     Config
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func Handler() core.InputHandler {
	return func(config []byte) (core.Input, error) {
		// Set config defaults
		conf := Config{
			FlushFrequency: 300,
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

		// Validate credentials
		err = validateCredentialsOrPath(conf.Credentials, conf.CredentialsPath)
		if err != nil {
			return nil, err
		}

		// Setup context
		ctx, cancelFn := context.WithCancel(context.Background())

		return &pubSubInput{
			config:     conf,
			ctx:        ctx,
			cancelFunc: cancelFn,
		}, nil
	}
}

func (p *pubSubInput) Run(errorHandler core.ErrorHandler, state core.State, processPipe chan<- core.PipelineResults) {
	// Setup local variables
	tmpWriter, err := core.NewTmpWriter()
	if err != nil {
		errorHandler(true, err)
		return
	}

	// Setup new client
	opts := make([]option.ClientOption, 0)
	if p.config.Credentials != nil && len(p.config.Credentials) > 0 && string(p.config.Credentials) != "null" {
		opts = append(opts, option.WithCredentialsJSON(p.config.Credentials))
	} else if p.config.CredentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(p.config.CredentialsPath))
	}

	client, err := pubsub.NewClient(p.ctx, p.config.ProjectID, opts...)
	if err != nil {
		errorHandler(true, fmt.Errorf("issue setting up pub sub client: %s", err))
		return
	}

	// Setup subscription
	subscription := client.Subscription(p.config.SubscriptionID)

	// Setup wait group
	var wg sync.WaitGroup
	flushCtx, flushCancelFn := context.WithCancel(p.ctx)

	// Start pub sub receiver go routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		rErr := subscription.Receive(p.ctx, func(ctx context.Context, msg *pubsub.Message) {
			// Write new message data to tmp writer
			_, writeErr := tmpWriter.Write(msg.Data)
			if writeErr != nil {
				errorHandler(false, fmt.Errorf("issue writing pubsub message: %s", writeErr))
				msg.Nack()
			} else {
				msg.Ack()
			}
		})

		// Handle errors
		if rErr != nil {
			errorHandler(false, fmt.Errorf("error receiving subscription messages: %s", rErr))
		}

		// Cancel flush context
		flushCancelFn()
	}()

	// Start timed process sync go routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-flushCtx.Done():
				return
			case <-time.After(time.Duration(p.config.FlushFrequency) * time.Second):
				err = flush(tmpWriter, processPipe)
				if err != nil {
					errorHandler(false, fmt.Errorf("issue flushing file: %s", err))
				}
			}
		}
	}()

	wg.Wait()

	// Flush any remaining data to pipeline before return
	err = flush(tmpWriter, processPipe)
	if err != nil {
		errorHandler(false, fmt.Errorf("issue flushing file: %s", err))
	}
}

func (p *pubSubInput) Stop() {
	p.cancelFunc()
}

func validateCredentialsOrPath(credentials json.RawMessage, path string) error {
	if credentials != nil && len(credentials) > 0 && string(credentials) != "null" {
		return nil
	} else if path != "" {
		return nil
	}

	return fmt.Errorf("missing credentials")
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
