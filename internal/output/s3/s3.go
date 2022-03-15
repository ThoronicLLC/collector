package s3

import (
	"encoding/json"
	"fmt"
	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
	"time"
)

type s3Output struct {
	config []byte
}

type s3Config struct {
	Bucket          string `json:"bucket"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	FilePrefix      string `json:"file_prefix"`
	FileExtension   string `json:"file_extension"`
	FileTimestamp   bool   `json:"file_timestamp"`
	MaxRetries      int    `json:"max_retries"`
}

func Handler() core.OutputHandler {
	return func(config []byte) core.Output {
		return &s3Output{
			config: config,
		}
	}
}

func (s *s3Output) Write(inputFile string) (int, error) {
	var conf s3Config
	err := json.Unmarshal(s.config, &conf)
	if err != nil {
		return 0, fmt.Errorf("issue unmarshalling config: %s", err)
	}

	// Open file
	fs, err := os.Open(inputFile)
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
	creds := credentials.NewStaticCredentials(conf.AccessKeyID, conf.SecretAccessKey, "")
	_, err = creds.Get()
	if err != nil {
		return 0, fmt.Errorf("invalid credentials: %s", err)
	}

	// Initialize AWS config
	awsConf := aws.NewConfig().WithRegion(conf.Region).WithCredentials(creds)
	if conf.MaxRetries > 0 {
		awsConf = awsConf.WithMaxRetries(conf.MaxRetries)
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

	// Build file name
	fileName := fmt.Sprintf("%s", conf.FilePrefix)
	if conf.FileTimestamp {
		fileName = fmt.Sprintf("%s.%d", fileName, time.Now().Unix())
	}
	fileName = fmt.Sprintf("%s.%s", fileName, conf.FileExtension)

	// Create upload input
	uploadInput := &s3manager.UploadInput{
		Bucket: aws.String(conf.Bucket),
		Key:    aws.String(fileName),
		Body:   fs,
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
