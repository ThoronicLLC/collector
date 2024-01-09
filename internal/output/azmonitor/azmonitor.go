package azmonitor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"

	"github.com/ThoronicLLC/collector/pkg/core"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/azingest"
)

var OutputName = "azure_monitor"

type Config struct {
	TenantID             string `json:"tenant_id" validate:"required"`
	ClientID             string `json:"client_id" validate:"required"`
	ClientSecret         string `json:"client_secret" validate:"required"`
	CollectionEndpoint   string `json:"collection_endpoint" validate:"required"`
	CollectionRuleID     string `json:"collection_rule_id" validate:"required"`
	CollectionStreamName string `json:"collection_stream_name" validate:"required"`
}

type azureMonitorOutput struct {
	config     Config
	ctx        context.Context
	cancelFunc context.CancelFunc
}

type logAnalyticsDefaultLog struct {
	Message string `json:"message"`
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

		// Setup context
		ctx, cancelFn := context.WithCancel(context.Background())

		return &azureMonitorOutput{
			config:     conf,
			ctx:        ctx,
			cancelFunc: cancelFn,
		}, nil
	}
}

func (azMon *azureMonitorOutput) Write(inputFile string) (int, error) {
	// Make upload buffer and line count
	uploadBuffer := make([]interface{}, 0)
	uploadBufferByteSize := 0
	lineCount := 0
	emptyLines := 0

	// Set up the credential using azidentity
	cred, err := azidentity.NewClientSecretCredential(azMon.config.TenantID, azMon.config.ClientID, azMon.config.ClientSecret, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to obtain a credential: %w", err)
	}

	client, err := azingest.NewClient(azMon.config.CollectionEndpoint, cred, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create a client: %w", err)
	}

	// Open file
	file, err := os.Open(inputFile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Start reading from the file with a reader.
	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, core.MaxLogSize)
	scanner.Buffer(buffer, core.MaxLogSize)
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine != "" {
			var tmpLog interface{}
			var tmpLogBytes []byte

			// If the value is a valid json, serialize it, if not, set as message to default type
			if json.Valid([]byte(trimmedLine)) {
				err = json.Unmarshal([]byte(trimmedLine), &tmpLog)
				if err != nil {
					return 0, fmt.Errorf("issue unmarshalling line: %s", err)
				}
			} else {
				tmpLog = logAnalyticsDefaultLog{
					Message: trimmedLine,
				}
			}

			tmpLogBytes, err = json.Marshal(tmpLog)
			if err != nil {
				return 0, fmt.Errorf("issue marshalling log: %s", err)
			}

			if uploadBufferByteSize+len(tmpLogBytes) >= (1 * 1024 * 1024) {
				log.Debugf("buffer limit reached, uploading 1MB worth of data (%d log entries)", lineCount)
				lineCount = 1

				// Do upload
				err = azMon.upload(client, uploadBuffer)
				if err != nil {
					return 0, fmt.Errorf("issue uploading log: %s", err)
				}

				// Clear upload buffer and add new line
				uploadBuffer = make([]interface{}, 0)
				uploadBuffer = append(uploadBuffer, tmpLog)

				// Reset upload buffer byte size
				uploadBufferByteSize = len(tmpLogBytes)
			} else {
				lineCount++
				uploadBufferByteSize += len(tmpLogBytes)
				uploadBuffer = append(uploadBuffer, tmpLog)
			}
		} else {
			emptyLines++
		}
	}

	// Upload any remaining data
	if len(uploadBuffer) > 0 {
		log.Debugf("uploading remaining buffer data (%d log entries)", lineCount)
		err = azMon.upload(client, uploadBuffer)
		if err != nil {
			return 0, fmt.Errorf("issue uploading logs to log analytics: %s", err)
		}
	}

	// Debug print with empty line count
	if emptyLines > 0 {
		log.Debugf("ignored %d empty log entries", emptyLines)
	}

	return lineCount, nil
}

func (azMon *azureMonitorOutput) upload(client *azingest.Client, buffer []interface{}) error {
	// Marshal buffer to json
	jsonBuffer, err := json.Marshal(buffer)
	if err != nil {
		return fmt.Errorf("issue marshalling buffer to json: %s", err)
	}

	// Upload the data
	_, err = client.Upload(azMon.ctx, azMon.config.CollectionRuleID, azMon.config.CollectionStreamName, jsonBuffer, nil)
	if err != nil {
		return fmt.Errorf("failed to upload data: %w", err)
	}

	return nil
}
