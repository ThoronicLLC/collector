package gcs

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/ThoronicLLC/collector/pkg/core/variable_replacer"
	"github.com/google/uuid"
	"google.golang.org/api/option"
	"io"
	"os"
	"time"
)

var OutputName = "gcs"

type Config struct {
	Bucket          string          `json:"bucket" validate:"required"`
	Path            string          `json:"path" validate:"required"`
	Credentials     json.RawMessage `json:"credentials,omitempty"`
	CredentialsPath string          `json:"credentials_path"`
	Composite       bool            `json:"composite"`
}

type gcsOutput struct {
	config     Config
	ctx        context.Context
	cancelFunc context.CancelFunc
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

		// Setup context
		ctx, cancelFn := context.WithCancel(context.Background())

		return &gcsOutput{
			config:     conf,
			ctx:        ctx,
			cancelFunc: cancelFn,
		}, nil
	}
}

func (g *gcsOutput) Write(inputFile string) (int, error) {
	// Get current time
	currentTime := time.Now()

	// Open file
	fs, err := os.Open(inputFile)
	if err != nil {
		return 0, fmt.Errorf("issue opening file: %s", err)
	}
	defer fs.Close()

	// Setup new client
	opts := make([]option.ClientOption, 0)
	if g.config.Credentials != nil && len(g.config.Credentials) > 0 && string(g.config.Credentials) != "null" {
		opts = append(opts, option.WithCredentialsJSON(g.config.Credentials))
	} else if g.config.CredentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(g.config.CredentialsPath))
	}

	// Setup storage client
	client, err := storage.NewClient(g.ctx, opts...)
	if err != nil {
		return 0, fmt.Errorf("issue initializing storage client: %s", err)
	}

	// Build file name
	fileName := variable_replacer.VariableReplacer(currentTime, g.config.Path)
	firstFileName := fileName

	// Generate temporary file name if composition is enabled
	if g.config.Composite {
		firstFileName = fmt.Sprintf("%s.%s.tmp", fileName, uuid.New().String())
	}

	// Initialize the cloud file and writer
	googleCloudStorageFile := client.Bucket(g.config.Bucket).Object(firstFileName)
	gcsFileWriter := googleCloudStorageFile.NewWriter(g.ctx)

	// Upload the file
	if _, err = io.Copy(gcsFileWriter, fs); err != nil {
		return 0, fmt.Errorf("issue copying stream: %s", err)
	}

	// Handle google cloud storage file closure errors
	if err = gcsFileWriter.Close(); err != nil {
		return 0, fmt.Errorf("issue closing google file writer: %s", err)
	}

	// Handle source file closure errors
	if err = fs.Close(); err != nil {
		return 0, fmt.Errorf("issue closing local file reader: %s", err)
	}

	// If file composition is enabled, do it
	if g.config.Composite {
		compositeFile := client.Bucket(g.config.Bucket).Object(fileName)

		// Check if composite file already exists
		_, err = compositeFile.Attrs(g.ctx)

		// If composite file does not exist, move file; if it does, create a composition and remove the new file.
		if err == storage.ErrObjectNotExist {
			// Copy file to composite file since it does not exist
			if _, err = compositeFile.CopierFrom(googleCloudStorageFile).Run(g.ctx); err != nil {
				return 0, fmt.Errorf("issue moving google cloud storage file: %s", err)
			}
		} else {
			// Copy data from new file and the old composite file into a new composition
			composer := compositeFile.ComposerFrom(compositeFile, googleCloudStorageFile)
			if _, err = composer.Run(g.ctx); err != nil {
				return 0, fmt.Errorf("issue create composite file: %s", err)
			}
		}

		// Delete new file now it has been moved/appended to the composite file
		if err = googleCloudStorageFile.Delete(g.ctx); err != nil {
			return 0, fmt.Errorf("issue deleting temporary cloud file: %s", err)
		}
	}

	// Handle storage client closure errors
	if err = client.Close(); err != nil {
		return 0, fmt.Errorf("issue closing storage client: %s", err)
	}

	return 0, nil
}

func validateCredentialsOrPath(credentials json.RawMessage, path string) error {
	if credentials != nil && len(credentials) > 0 && string(credentials) != "null" {
		return nil
	} else if path != "" {
		return nil
	}

	return fmt.Errorf("missing credentials")
}
