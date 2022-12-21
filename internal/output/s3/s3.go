package s3

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/ThoronicLLC/collector/pkg/core/variable_replacer"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"os"
	"time"
)

var OutputName = "s3"

type Config struct {
	Bucket          string `json:"bucket" validate:"required"`
	Region          string `json:"region" validate:"required"`
	AccessKeyID     string `json:"access_key_id" validate:"required"`
	SecretAccessKey string `json:"secret_access_key" validate:"required"`
	Path            string `json:"path" validate:"required"`
	MaxRetries      int    `json:"max_retries" validate:"int|min:0"`
	GZip            bool   `json:"gzip" validate:"bool"`
}

type s3Output struct {
	config Config
}

func Handler() core.OutputHandler {
	return func(config []byte) (core.Output, error) {
		// Set config defaults
		conf := Config{
			MaxRetries: 3,
			GZip:       false,
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

		return &s3Output{
			config: conf,
		}, nil
	}
}

func (s *s3Output) Write(inputFile string) (int, error) {
	currentFile := inputFile

	// Handle gzip option
	if s.config.GZip {
		// Create a temp file
		newTmpFile, err := os.CreateTemp("", "gzip-output-")
		if err != nil {
			return 0, fmt.Errorf("unable to create temp file for gzip: %w", err)
		}
		defer os.Remove(newTmpFile.Name())

		// Set up a new gzip writer and file reader
		gzipWriter := gzip.NewWriter(newTmpFile)
		tmpFs, err := os.Open(inputFile)
		if err != nil {
			return 0, fmt.Errorf("unable to open input file for gzip: %w", err)
		}

		// Copy the file to the gzip writer
		_, err = io.Copy(gzipWriter, tmpFs)
		if err != nil {
			return 0, fmt.Errorf("unable to copy input file to gzip: %w", err)
		}

		// Close the gzip writer
		err = gzipWriter.Close()
		if err != nil {
			return 0, fmt.Errorf("unable to close gzip writer: %w", err)
		}

		// Close the file reader
		err = tmpFs.Close()
		if err != nil {
			return 0, fmt.Errorf("unable to close input file: %w", err)
		}

		// Sync the temp file
		err = newTmpFile.Sync()
		if err != nil {
			return 0, fmt.Errorf("unable to sync temp file: %w", err)
		}

		// Set the current file to the temp file
		currentFile = newTmpFile.Name()

		// Close the temp file
		err = newTmpFile.Close()
		if err != nil {
			return 0, fmt.Errorf("unable to close temp file: %w", err)
		}
	}

	// Open file
	fs, err := os.Open(currentFile)
	if err != nil {
		return 0, fmt.Errorf("issue opening file: %s", err)
	}
	defer fs.Close()

	// Stat file
	fileInfo, err := fs.Stat()
	if err != nil {
		return 0, fmt.Errorf("issue getting file stat: %s", err)
	}

	// Initialize AWS creds
	creds := credentials.NewStaticCredentials(s.config.AccessKeyID, s.config.SecretAccessKey, "")
	_, err = creds.Get()
	if err != nil {
		return 0, fmt.Errorf("invalid credentials: %s", err)
	}

	// Initialize AWS config
	awsConf := aws.NewConfig().WithRegion(s.config.Region).WithCredentials(creds)
	if s.config.MaxRetries > 0 {
		awsConf = awsConf.WithMaxRetries(s.config.MaxRetries)
	}

	// Initialize AWS session
	awsSession, err := session.NewSession(awsConf)
	if err != nil {
		return 0, fmt.Errorf("issue creating AWS session: %s", err)
	}

	partSize, err := getMaxPartSize(fileInfo.Size())
	if err != nil {
		return 0, err
	}

	// Create an uploader with the session and custom options
	uploader := s3manager.NewUploader(awsSession, func(u *s3manager.Uploader) {
		u.PartSize = partSize // We calculate part size based on file size
	})

	currentTime := time.Now()

	// Build file name
	fileName := variable_replacer.VariableReplacer(currentTime, s.config.Path)
	if s.config.GZip {
		fileName = fileName + ".gz"
	}

	// Create upload input
	uploadInput := &s3manager.UploadInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(fileName),
		Body:   fs,
	}

	// Set content encoding if gzip
	if s.config.GZip {
		uploadInput.ContentEncoding = aws.String("gzip")
		uploadInput.ContentType = aws.String("text/plain")
	}

	// Upload the file to S3.
	_, err = uploader.Upload(uploadInput)
	if err != nil {
		return 0, fmt.Errorf("issue uploading file: %s", err)
	}

	return 0, nil
}

// getMaxPartSize returns am acceptable part size for the supplied file size
//
// AWS limits the max file upload and the number of parts
// https://docs.aws.amazon.com/AmazonS3/latest/userguide/qfacts.html
func getMaxPartSize(fileSize int64) (int64, error) {
	maxParts := int64(10000)                // Max multipart upload parts is 10,000
	maxFileSize := int64(5 * 1099511627776) // Max file size in S3 is 5TB
	partSize1 := int64(5 * 1024 * 1024)     // 5 MB
	partSize2 := int64(32 * 1024 * 1024)    // 32 MB
	partSize3 := int64(64 * 1024 * 1024)    // 64 MB
	partSize4 := int64(128 * 1024 * 124)    // 128 MB
	partSize5 := int64(256 * 1024 * 124)    // 256 MB
	partSize6 := int64(512 * 1024 * 124)    // 512 MB

	if fileSize < (partSize1 * maxParts) {
		return partSize1, nil
	} else if fileSize < (partSize2 * maxParts) {
		return partSize2, nil
	} else if fileSize < (partSize3 * maxParts) {
		return partSize3, nil
	} else if fileSize < (partSize4 * maxParts) {
		return partSize4, nil
	} else if fileSize < (partSize5 * maxParts) {
		return partSize5, nil
	} else if fileSize < (partSize6*maxParts) && fileSize < maxFileSize {
		return partSize6, nil
	} else {
		return 0, fmt.Errorf("file is too large to upload; max supported file size is 5TB")
	}
}
