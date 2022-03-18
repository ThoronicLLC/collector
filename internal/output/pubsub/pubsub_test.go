package pubsub

import (
	"encoding/json"
	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

var config1 = `{"project_id": "project-1", "topic_id": "sub-1", "credentials_path": "/tmp/file.txt"}`
var config2 = `{"project_id": "project-2", "topic_id": "sub-3", "credentials": {}}`
var badConfig1 = `{"project_id": "project-3", "topic_id": "", "credentials": {}}`
var badConfig2 = `{"project_id": "", "topic_id": "sub-3", "credentials": {}}`
var badConfig3 = `{"project_id": "project-1", "topic_id": "sub-1"}`

func TestValidate(t *testing.T) {
	arr := []string{config1, config2}
	for i, v := range arr {
		var testConfig Config
		err := json.Unmarshal([]byte(v), &testConfig)
		assert.Nilf(t, err, "test #%d - failed to unmarshal json: %s", i, err)
		err = core.ValidateStruct(&testConfig)
		assert.Nilf(t, err, "test #%d - validation error: %s", i, err)
	}
}

func TestValidateFailed(t *testing.T) {
	arr := []string{badConfig1, badConfig2}
	for i, v := range arr {
		var testConfig Config
		err := json.Unmarshal([]byte(v), &testConfig)
		assert.Nilf(t, err, "test #%d - failed to unmarshal json: %s", i, err)
		err = core.ValidateStruct(&testConfig)
		assert.NotNilf(t, err, "test #%d - validation should have returned an error: %s", i, err)
	}
}

func TestHandler(t *testing.T) {
	arr := []string{config1, config2}
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
	arr := []string{badConfig1, badConfig2, badConfig3}
	for i, v := range arr {
		var testConfig Config
		err := json.Unmarshal([]byte(v), &testConfig)
		assert.Nilf(t, err, "test #%d - failed to unmarshal json: %s", i, err)
		handleFunc := Handler()
		_, err = handleFunc([]byte(v))
		assert.NotNilf(t, err, "test #%d - validation should have returned an error: %s", i, err)
	}
}
