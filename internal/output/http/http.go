package http

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/bfp"
	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/go-resty/resty/v2"
	"os"
)

var OutputName = "http"

type Config struct {
	URL         string            `json:"url" validate:"required"`
	Headers     map[string]string `json:"headers"`
	MaxSize     int               `json:"max_size" validate:"int" message:"int:max_size must int"` // Max size in KB
	AsMultiPart bool              `json:"as_multi_part"`
	AsJson      bool              `json:"as_json"`
}

type httpOutput struct {
	config Config
	client *resty.Client
}

func Handler() core.OutputHandler {
	return func(config []byte) (core.Output, error) {
		// Set config defaults
		conf := Config{
			MaxSize: 128, // Default 128KB max size for requests
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

		// Setup client
		client := resty.New()
		for k, v := range conf.Headers {
			client.SetHeader(k, v)
		}

		// Return output
		return &httpOutput{
			config: conf,
			client: client,
		}, nil
	}
}

func (h *httpOutput) Write(inputFile string) (int, error) {
	bufferedPart := 1

	// Use a buffered reader to process a file in chunks
	bProc := bfp.New(bfp.WithMaxSize(h.config.MaxSize * 1024))

	// Submit the http requests as different parts based on max size
	lineCount, err := bProc.Process(inputFile, func(path string) error {
		// Opened the file
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("unable to open file: %s", err)
		}

		// Get request
		request := h.client.R()

		// Handle multipart
		if h.config.AsMultiPart {
			name := fmt.Sprintf("%s_part-%d.log", inputFile, bufferedPart)
			request = request.SetFileReader("file", name, file)
		} else {
			// Handle normal requests
			requestBody := make([]interface{}, 0)
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()

				// Handle json logs
				if h.config.AsJson {
					var jsonLog interface{}
					err := json.Unmarshal([]byte(line), &jsonLog)
					if err != nil {
						continue
					}
					requestBody = append(requestBody, jsonLog)
				} else {
					requestBody = append(requestBody, line)
				}
			}
			request = request.SetBody(requestBody)
		}

		// Conduct the request and handle errors
		response, err := request.Post(h.config.URL)
		if err != nil {
			return err
		}
		if response.IsError() {
			return fmt.Errorf("invalid response: %s", response.Status())
		}

		// Increment part
		bufferedPart++

		return nil
	})
	if err != nil {
		return lineCount, err
	}

	return lineCount, nil
}
