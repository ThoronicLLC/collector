package log_analytics

import (
	"bufio"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"time"
)

var OutputName = "log_analytics"

type Config struct {
	LogType     string `json:"log_type" validate:"required"`
	WorkspaceID string `json:"workspace_id" validate:"required"`
	PrimaryKey  string `json:"primary_key" validate:"required"`
	DateField   string `json:"date_field,omitempty"`
}

type logAnalyticsOutput struct {
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
		conf := Config{
			DateField: "Timestamp",
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

		return &logAnalyticsOutput{
			config:     conf,
			ctx:        ctx,
			cancelFunc: cancelFn,
		}, nil
	}
}

func (l *logAnalyticsOutput) Write(inputFile string) (int, error) {
	// Make upload buffer and line count
	uploadBuffer := make([]interface{}, 0)
	uploadBufferByteSize := 0
	lineCount := 0
	emptyLines := 0

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

			if uploadBufferByteSize+len(tmpLogBytes) >= (25 * 1024 * 1024) {
				log.Debugf("buffer limit reached, uploading 25MB worth of data (%d log entries)", lineCount)
				lineCount = 1

				// Do upload
				err = logAnalyticsUpload(uploadBuffer, l.config.LogType, l.config.WorkspaceID, l.config.PrimaryKey, l.config.DateField)
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
		err = logAnalyticsUpload(uploadBuffer, l.config.LogType, l.config.WorkspaceID, l.config.PrimaryKey, l.config.DateField)
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

func logAnalyticsBuildSignature(message, secret string) (string, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, keyBytes)
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

func logAnalyticsUpload(data []interface{}, logName, workspaceID, key, dateField string) error {
	// Marshal data
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	dateString := time.Now().UTC().Format(time.RFC1123)
	dateString = strings.Replace(dateString, "UTC", "GMT", -1)

	stringToHash := "POST\n" + strconv.Itoa(len(dataBytes)) + "\napplication/json\n" + "x-ms-date:" + dateString + "\n/api/logs"

	hashedString, err := logAnalyticsBuildSignature(stringToHash, key)
	if err != nil {
		return err
	}

	signature := fmt.Sprintf("SharedKey %s:%s", workspaceID, hashedString)
	uri := fmt.Sprintf("https://%s.ods.opinsights.azure.com/api/logs?api-version=2016-04-01", workspaceID)

	request := resty.New().SetRetryCount(3).R()
	request.SetHeader("Log-Type", logName)
	request.SetHeader("Authorization", signature)
	request.SetHeader("Content-Type", "application/json")
	request.SetHeader("x-ms-date", dateString)
	request.SetHeader("time-generated-field", dateField)

	// Set body and post
	request.SetBody(dataBytes)
	response, err := request.Post(uri)

	// Handle error
	if err != nil {
		return err
	}

	// Handle response error
	if response.IsError() {
		return fmt.Errorf("response returned: %s", response.Status())
	}

	return nil
}
