package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"sync"
	"time"
)

var InputName = "sqs"

type Config struct {
	QueueUrl        string `json:"queue_url" validate:"required"`
	Region          string `json:"region" validate:"required"`
	AccessKeyID     string `json:"access_key_id" validate:"required"`
	SecretAccessKey string `json:"secret_access_key" validate:"required"`
	PollFrequency   int    `json:"poll_frequency" validate:"required|int|min:10"`
	FlushFrequency  int    `json:"flush_frequency" validate:"required|int|min:10"`
}

type sqsInput struct {
	config     Config
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func Handler() core.InputHandler {
	return func(config []byte) (core.Input, error) {
		// Set config defaults
		conf := defaultConfig()

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

		return &sqsInput{
			config:     conf,
			ctx:        ctx,
			cancelFunc: cancelFn,
		}, nil
	}
}

func (s *sqsInput) Run(errorHandler core.ErrorHandler, state core.State, processPipe chan<- core.PipelineResults) {
	// Setup local variables
	tmpWriter, err := core.NewTmpWriter()
	if err != nil {
		errorHandler(true, err)
		return
	}

	// Initialize AWS creds
	creds := credentials.NewStaticCredentials(s.config.AccessKeyID, s.config.SecretAccessKey, "")
	_, err = creds.Get()
	if err != nil {
		errorHandler(true, fmt.Errorf("invalid aws credentials: %s", err))
		return
	}

	// Initialize AWS config
	awsConf := aws.NewConfig().WithRegion(s.config.Region).WithCredentials(creds)

	// Initialize AWS session
	awsSession, err := session.NewSession(awsConf)
	if err != nil {
		errorHandler(true, fmt.Errorf("issue creating AWS session: %s", err))
		return
	}

	sqsService := sqs.New(awsSession)

	// Setup wait group
	var wg sync.WaitGroup
	flushCtx, flushCancelFn := context.WithCancel(s.ctx)

	// Start pub sub receiver go routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer flushCancelFn()
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				// Long poll SQS
				output, err := sqsService.ReceiveMessageWithContext(s.ctx, &sqs.ReceiveMessageInput{
					QueueUrl:            aws.String(s.config.QueueUrl),
					MaxNumberOfMessages: aws.Int64(10000),
					WaitTimeSeconds:     aws.Int64(int64(s.config.PollFrequency)),
				})
				if err != nil {
					errorHandler(true, fmt.Errorf("failed to fetch sqs messages: %s", err))
					return
				}

				// Loop through received messages and skip those with an empty body
				if output != nil {
					for _, message := range output.Messages {
						if message != nil {
							if message.Body != nil {
								safeBody := derefString(message.Body)
								if safeBody != "" {
									_, err = tmpWriter.Write([]byte(safeBody))
									if err != nil {
										errorHandler(false, fmt.Errorf("issue writing sqs message to tmp file: %s", err))
									}
								}
							}
						}
					}
				}
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
			case <-time.After(time.Duration(s.config.FlushFrequency) * time.Second):
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

func (s *sqsInput) Stop() {
	s.cancelFunc()
}

func derefString(s *string) string {
	if s != nil {
		return *s
	}

	return ""
}

func defaultConfig() Config {
	return Config{
		PollFrequency:  20,
		FlushFrequency: 300,
	}
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
