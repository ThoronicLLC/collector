package pubsub

import (
	"encoding/json"
	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

var config1 = `{"project_id": "project-1", "subscription_id": "sub-1", "credentials_path": "/tmp/file.txt", "flush_frequency": 10}`
var config2 = `{"project_id": "project-2", "subscription_id": "sub-2", "credentials_path": "/tmp/file.txt", "flush_frequency": 100}`
var config3 = `{"project_id": "project-3", "subscription_id": "sub-3", "credentials": {}, "flush_frequency": 1000}`
var config4 = `{"project_id": "project-4", "subscription_id": "sub-4", "credentials": {}, "flush_frequency": 10000}`
var badConfig1 = `{"project_id": "project-1", "subscription_id": "sub-1", "credentials": {}, "flush_frequency": -1}`
var badConfig2 = `{"project_id": "project-2", "subscription_id": "sub-2", "credentials": {}, "flush_frequency": 0}`
var badConfig3 = `{"project_id": "project-3", "subscription_id": "", "credentials": {}, "flush_frequency": 10}`
var badConfig4 = `{"project_id": "", "subscription_id": "sub-3", "credentials": {}, "flush_frequency": 10}`
var badConfig5 = `{"project_id": "project-1", "subscription_id": "sub-1", "flush_frequency": 10}`

func TestValidate(t *testing.T) {
	arr := []string{config1, config2, config3, config4}
	for i, v := range arr {
		var testConfig Config
		err := json.Unmarshal([]byte(v), &testConfig)
		assert.Nilf(t, err, "test #%d - failed to unmarshal json: %s", i, err)
		err = core.ValidateStruct(&testConfig)
		assert.Nilf(t, err, "test #%d - validation error: %s", i, err)
	}
}

func TestValidateFailed(t *testing.T) {
	arr := []string{badConfig1, badConfig2, badConfig3, badConfig4}
	for i, v := range arr {
		var testConfig Config
		err := json.Unmarshal([]byte(v), &testConfig)
		assert.Nilf(t, err, "test #%d - failed to unmarshal json: %s", i, err)
		err = core.ValidateStruct(&testConfig)
		assert.NotNilf(t, err, "test #%d - validation should have returned an error: %s", i, err)
	}
}

func TestHandler(t *testing.T) {
	arr := []string{config1, config2, config3, config4}
	for i, v := range arr {
		var testConfig Config
		err := json.Unmarshal([]byte(v), &testConfig)
		assert.Nilf(t, err, "test #%d - failed to unmarshal json: %s", i, err)
		handleFunc := Handler()
		_, err = handleFunc([]byte(v))
		assert.Nilf(t, err, "test #%d - validation error: %s", i, err)
	}
}

func TestHandlerFailed(t *testing.T) {
	arr := []string{badConfig1, badConfig2, badConfig3, badConfig4, badConfig5}
	for i, v := range arr {
		var testConfig Config
		err := json.Unmarshal([]byte(v), &testConfig)
		assert.Nilf(t, err, "test #%d - failed to unmarshal json: %s", i, err)
		handleFunc := Handler()
		_, err = handleFunc([]byte(v))
		assert.NotNilf(t, err, "test #%d - validation should have returned an error: %s", i, err)
	}
}
