package msgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/internal/integrations/msgraph"
	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/tidwall/pretty"
	"net/url"
	"time"
)

var InputName = "msgraph"

type Config struct {
	TenantID     string `json:"tenant_id" validate:"required"`
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
	Schedule     int    `json:"schedule" validate:"required|min:0"`

	// This is not required. We use the default/global endpoint which will work for most users
	// https://docs.microsoft.com/en-us/graph/deployments#microsoft-graph-and-graph-explorer-service-root-endpoints
	GraphEndpoint string `json:"graph_endpoint,omitempty"`
	LoginEndpoint string `json:"login_endpoint,omitempty"`
}

type msgraphInput struct {
	config     Config
	ctx        context.Context
	cancelFunc context.CancelFunc
	client     *msgraph.Client
}

type msgraphState struct {
	LastTimestamp int64 `json:"last_timestamp"`
}

func Handler() core.InputHandler {
	return func(config []byte) (core.Input, error) {
		// Set config defaults
		conf := Config{
			Schedule: 60,
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
		client := msgraph.NewClient(conf.TenantID, conf.ClientID, conf.ClientSecret, msgraph.GraphScope.String())

		// Setup other MS login endpoint
		if conf.LoginEndpoint != "" {
			err = client.SetLoginEndpoint(conf.LoginEndpoint)
			if err != nil {
				return nil, err
			}
		}

		// Setup other MS graph endpoint
		if conf.GraphEndpoint != "" {
			err = client.SetGraphEndpoint(conf.GraphEndpoint)
			if err != nil {
				return nil, err
			}
		}

		// Setup context
		ctx, cancelFn := context.WithCancel(context.Background())

		return &msgraphInput{
			config:     conf,
			ctx:        ctx,
			cancelFunc: cancelFn,
			client:     client,
		}, nil
	}
}

// Run will execute the input with the supplied context and state and return results
func (input *msgraphInput) Run(errorHandler core.ErrorHandler, state core.State, processPipe chan<- core.PipelineResults) {
	// Test login
	isValid := input.client.Ping()
	if !isValid {
		errorHandler(true, fmt.Errorf("failed to ping microsoft graph client"))
		return
	}

	// Validate and load state
	currentState, err := loadState(state)
	if err != nil {
		errorHandler(true, err)
		return
	}

	for {
		select {
		case <-input.ctx.Done():
			return
		case <-time.After(time.Duration(input.config.Schedule) * time.Second):
			currentTime := time.Now()
			currentTimeUnix := currentTime.Unix()
			pastTimeUnix := currentState.LastTimestamp
			pastTime := time.Unix(pastTimeUnix, 0)

			// Create temp file
			tmpFile, err := core.NewTmpWriter()
			if err != nil {
				errorHandler(false, fmt.Errorf("issue opening a new temp file writer: %s", err))
				continue
			}

			// Copy current state to a new object for modification
			newState := currentState

			// Convert times to strings
			gtTime := pastTime.UTC().Format("2006-01-02T15:04:05Z")
			leTime := currentTime.UTC().Format("2006-01-02T15:04:05Z")

			// Get alerts and loop through if there are new lines
			params := url.Values{}
			params.Set("$top", "1000")
			params.Set("$filter", fmt.Sprintf("createdDateTime gt %s and createdDateTime le %s", gtTime, leTime))
			hasError := false
			for {
				// Get security alerts with params
				response, err := input.client.SecurityAlerts(params)
				if err != nil {
					errorHandler(false, fmt.Errorf("issue getting graph security alerts: %s", err))
					hasError = true
					break
				}

				// Loop through all responses
				for _, v := range response.Value {
					event, err := json.Marshal(v)
					if err != nil {
						errorHandler(false, fmt.Errorf("issue marshalling alert: %s", err))
						hasError = true
						break
					}
					_, err = tmpFile.Write(pretty.Ugly(event))
					if err != nil {
						errorHandler(false, fmt.Errorf("issue writing alert to temp file: %s", err))
						hasError = true
						break
					}
				}

				if response.NextLink == "" {
					break
				}

				// Parse next link
				nextLink, err := url.Parse(response.NextLink)

				// Handle error
				if err != nil {
					errorHandler(false, fmt.Errorf(""))
					hasError = true
					break
				}

				// Parse query params
				nextLinkParams, _ := url.ParseQuery(nextLink.RawQuery)

				// Get skip token
				skipToken := nextLinkParams.Get("$skiptoken")

				// Set skip token
				params.Set("$skiptoken", skipToken)
			}

			// Restart if an error occurred (don't want a broken partial state)
			if hasError {
				continue
			}

			// Get results file name and size
			path := tmpFile.CurrentFile().Name()
			linesWritten := tmpFile.WriteCount
			err = tmpFile.Close()
			if err != nil {
				errorHandler(false, fmt.Errorf("issue closing file: %s", err))
				continue
			}

			// Set new timestamp
			newState.LastTimestamp = currentTimeUnix

			// Marshal new state
			newStateBytes, err := json.Marshal(newState)
			if err != nil {
				errorHandler(false, fmt.Errorf("issue marshalling new state: %s", err))
				continue
			}

			// Setup pipeline results for next stage
			result := core.PipelineResults{
				FilePath:    path,
				ResultCount: linesWritten,
				State:       newStateBytes,
				RetryCount:  0,
			}

			// Pipe results to next stage
			processPipe <- result

			// Update current state to the new state since successful run
			currentState = newState
		}
	}
}

func (input *msgraphInput) Stop() {
	input.cancelFunc()
}

func loadState(state core.State) (*msgraphState, error) {
	if state == nil {
		return defaultState(), nil
	}

	var s msgraphState
	err := json.Unmarshal(state, &s)
	if err != nil {
		return nil, fmt.Errorf("issue unmarshalling state: %s", err)
	}

	return &s, nil
}

func defaultState() *msgraphState {
	return &msgraphState{LastTimestamp: time.Now().Unix()}
}
