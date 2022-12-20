package kafka

import (
  "encoding/json"
  "github.com/ThoronicLLC/collector/pkg/core"
  "github.com/stretchr/testify/assert"
  "testing"
)

var config1 = `{"brokers": ["uri-1", "uri-2"], "topic": "topic-1", "group_id": "security", "flush_frequency": 300}`
var config2 = `{"brokers": ["uri-1"], "topic": "topic-1", "group_id": "security", "auth_config": {"scram_sha_256": {"enabled": true, "username": "user", "password": "pass"}}, "flush_frequency": 300}`
var config3 = `{"brokers": ["uri-1"], "topic": "topic-1", "group_id": "security", "auth_config": {"scram_sha_512": {"enabled": true, "username": "user", "password": "pass"}}, "flush_frequency": 300}`
var config4 = `{"brokers": ["uri-1"], "topic": "topic-1", "group_id": "security", "auth_config": {"gssapi_password": {"enabled": true, "username": "user", "password": "pass"}}, "flush_frequency": 300}`
var badConfig1 = `{"brokers": [], "topic": "topic-1", "group_id": "security", "flush_frequency": 300}`
var badConfig2 = `{"brokers": [], "topic": "", "group_id": "", "flush_frequency": 0}`
var badConfig3 = `{"brokers": ["uri-1"], "topic": "", "group_id": "security", "flush_frequency": 300}`
var badConfig4 = `{"brokers": ["uri-1"], "topic": "topic-1", "group_id": "", "flush_frequency": 300}`
var badConfig5 = `{"brokers": ["uri-1"], "topic": "topic-1", "group_id": "security", "flush_frequency": 0}`

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
  arr := []string{badConfig1, badConfig2, badConfig3, badConfig4, badConfig5}
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
