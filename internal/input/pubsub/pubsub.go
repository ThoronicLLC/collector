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

type Config struct {
	ProjectID       string          `json:"project_id"`
	SubscriptionID  string          `json:"subscription_id"`
	Credentials     json.RawMessage `json:"credentials,omitempty"`
	CredentialsPath string          `json:"credentials_path"`
	FlushFrequency  int             `json:"flush_frequency"`
}

type pubSubInput struct {
	Config     []byte
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func Handler() core.InputHandler {
	return func(config []byte) core.Input {
		// Setup context
		ctx, cancelFn := context.WithCancel(context.Background())

		return &pubSubInput{
			Config:     config,
			ctx:        ctx,
			cancelFunc: cancelFn,
		}
	}
}

func (p *pubSubInput) Run(errorHandler core.ErrorHandler, state core.State, processPipe chan<- core.PipelineResults) {
	// Serialize config
	var conf Config
	err := json.Unmarshal(p.Config, &conf)
	if err != nil {
		errorHandler(true, fmt.Errorf("issue unmarshalling config: %s", err))
		return
	}

	// If flush frequency is not set, set to default of 300 seconds (5 minutes)
	if conf.FlushFrequency == 0 {
		conf.FlushFrequency = 300
	}

	// Setup local variables
	tmpWriter, err := core.NewTmpWriter()
	if err != nil {
		errorHandler(true, err)
		return
	}

	// Setup new client
	opts := make([]option.ClientOption, 0)
	if conf.Credentials != nil && len(conf.Credentials) > 0 && string(conf.Credentials) != "null" {
		opts = append(opts, option.WithCredentialsJSON(conf.Credentials))
	} else if conf.CredentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(conf.CredentialsPath))
	}

	client, err := pubsub.NewClient(p.ctx, conf.ProjectID, opts...)
	if err != nil {
		errorHandler(true, fmt.Errorf("issue setting up pub sub client: %s", err))
		return
	}

	// Setup subscription
	subscription := client.Subscription(conf.SubscriptionID)

	// Setup wait group
	var wg sync.WaitGroup

	// Start timed process sync go routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-p.ctx.Done():
				return
			case <-time.After(time.Duration(conf.FlushFrequency) * time.Second):
				// Rotate the temp writer
				count, fileName, rErr := tmpWriter.Rotate()
				if rErr != nil {
					errorHandler(true, fmt.Errorf("issue rotating temp file: %s", rErr))
					p.cancelFunc()
					continue
				}

				// Only send on if there are results
				if count > 0 {
					processPipe <- core.PipelineResults{
						FilePath:    fileName,
						ResultCount: count,
						State:       nil,
						RetryCount:  0,
					}
				}
			}
		}
	}()

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

		if rErr != nil {
			errorHandler(false, fmt.Errorf("error receiving subscription messages: %s", rErr))
		}
	}()

	wg.Wait()
}

func (p *pubSubInput) Stop() {
	p.cancelFunc()
}
