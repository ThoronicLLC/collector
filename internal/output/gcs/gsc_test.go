package gcs

import (
	"encoding/json"
	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

var config1 = `{"bucket": "sample-bucket", "path": "this/is/the/folder", "credentials_path": "/tmp/file.txt", "composite": true}`
var config2 = `{"bucket": "sample-bucket", "path": "this/is/the/folder", "credentials_path": "/tmp/file.txt"}`
var config3 = `{"bucket": "sample-bucket", "path": "this/is/the/folder", "credentials": {}}`
var config4 = `{"bucket": "sample-bucket", "path": "this/is/the/folder", "credentials": {}, "composite": false}`
var badConfig1 = `{"bucket": "", "path": "this/is/the/folder", "credentials": {}, "composite": false}`
var badConfig2 = `{"bucket": "sample-bucket", "path": "", "credentials": {}, "composite": false}`
var badConfig3 = `{"bucket": "sample-bucket", "path": "this/is/the/folder", "composite": false}`

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
