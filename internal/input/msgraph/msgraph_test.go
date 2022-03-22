package msgraph

import (
	"encoding/json"
	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

var config1 = `{"tenant_id": "tenant-1", "client_id": "client-1", "client_secret": "secret-1", "schedule": 10}`
var config2 = `{"tenant_id": "tenant-1", "client_id": "client-1", "client_secret": "secret-1", "schedule": 100}`
var badConfig1 = `{"tenant_id": "tenant-1", "client_id": "client-1", "client_secret": "secret-1", "schedule": -1}`
var badConfig2 = `{"tenant_id": "tenant-1", "client_id": "client-1", "client_secret": "secret-1", "schedule": 0}`
var badConfig3 = `{"tenant_id": "", "client_id": "client-1", "client_secret": "secret-1", "schedule": 10}`
var badConfig4 = `{"tenant_id": "tenant-1", "client_id": "", "client_secret": "secret-1", "schedule": 10}`
var badConfig5 = `{"tenant_id": "tenant-1", "client_id": "client-1", "client_secret": "", "schedule": 10}`

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
