package sqs

import (
	"encoding/json"
	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

var config1 = `{"queue_url": "https://example.com", "region": "us-east-1", "access_key_id": "1234567890", "secret_access_key": "1234567890", "poll_frequency": 30, "flush_frequency": 100}`
var config2 = `{"queue_url": "https://example.com", "region": "us-east-1", "access_key_id": "1234567890", "secret_access_key": "1234567890"}`
var badConfig1 = `{"queue_url": "", "region": "us-east-1", "access_key_id": "1234567890", "secret_access_key": "1234567890", "poll_frequency": 30, "flush_frequency": 100}`
var badConfig2 = `{"queue_url": "https://example.com", "region": "", "access_key_id": "1234567890", "secret_access_key": "1234567890", "poll_frequency": 30, "flush_frequency": 100}`
var badConfig3 = `{"queue_url": "https://example.com", "region": "us-east-1", "access_key_id": "", "secret_access_key": "1234567890", "poll_frequency": 30, "flush_frequency": 100}`
var badConfig4 = `{"queue_url": "https://example.com", "region": "us-east-1", "access_key_id": "1234567890", "secret_access_key": "", "poll_frequency": 30, "flush_frequency": 100}`
var badConfig5 = `{"queue_url": "https://example.com", "region": "us-east-1", "access_key_id": "1234567890", "secret_access_key": "1234567890", "poll_frequency": 0, "flush_frequency": 100}`
var badConfig6 = `{"queue_url": "https://example.com", "region": "us-east-1", "access_key_id": "1234567890", "secret_access_key": "1234567890", "poll_frequency": 30, "flush_frequency": 0}`

func TestValidate(t *testing.T) {
	arr := []string{config1, config2}
	for i, v := range arr {
		testConfig := defaultConfig()
		err := json.Unmarshal([]byte(v), &testConfig)
		assert.Nilf(t, err, "test #%d - failed to unmarshal json: %s", i, err)
		err = core.ValidateStruct(&testConfig)
		assert.Nilf(t, err, "test #%d - validation error: %s", i, err)
	}
}

func TestValidateFailed(t *testing.T) {
	arr := []string{badConfig1, badConfig2, badConfig3, badConfig4, badConfig5, badConfig6}
	for i, v := range arr {
		testConfig := defaultConfig()
		err := json.Unmarshal([]byte(v), &testConfig)
		assert.Nilf(t, err, "test #%d - failed to unmarshal json: %s", i, err)
		err = core.ValidateStruct(&testConfig)
		assert.NotNilf(t, err, "test #%d - validation should have returned an error: %s", i, err)
	}
}

func TestHandler(t *testing.T) {
	arr := []string{config1, config2}
	for i, v := range arr {
		testConfig := defaultConfig()
		err := json.Unmarshal([]byte(v), &testConfig)
		assert.Nilf(t, err, "test #%d - failed to unmarshal json: %s", i, err)
		handleFunc := Handler()
		_, err = handleFunc([]byte(v))
		assert.Nilf(t, err, "test #%d - validation error: %s", i, err)
	}
}

func TestHandlerFailed(t *testing.T) {
	arr := []string{badConfig1, badConfig2, badConfig3, badConfig4, badConfig5, badConfig6}
	for i, v := range arr {
		testConfig := defaultConfig()
		err := json.Unmarshal([]byte(v), &testConfig)
		assert.Nilf(t, err, "test #%d - failed to unmarshal json: %s", i, err)
		handleFunc := Handler()
		_, err = handleFunc([]byte(v))
		assert.NotNilf(t, err, "test #%d - validation should have returned an error: %s", i, err)
	}
}
