package syslog

import (
	"encoding/json"
	"github.com/ThoronicLLC/collector/pkg/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

var config1 = `{"address": "0.0.0.0", "port": 8433, "protocol": "tcp", "format": "raw", "flush_frequency": 10}`
var config2 = `{"address": "127.0.0.1", "port": 1514, "protocol": "udp", "format": "automatic", "flush_frequency": 100}`
var config3 = `{"address": "172.55.0.1", "port": 1514, "protocol": "both", "format": "RFC3164", "flush_frequency": 1000}`
var config4 = `{"address": "192.168.1.1", "port": 514, "protocol": "tcp", "format": "RFC5424", "flush_frequency": 10000}`
var config5 = `{"address": "10.120.0.1", "port": 1514, "protocol": "both", "format": "RFC6587", "flush_frequency": 100000}`
var badConfig1 = `{"address": "0.0.0.0", "port": 8433, "protocol": "tcp", "format": "raw", "flush_frequency": -1}`
var badConfig2 = `{"address": "0.0.0.0", "port": 8433, "protocol": "tcp", "format": "something", "flush_frequency": 10}`
var badConfig3 = `{"address": "0.0.0.0", "port": 8433, "protocol": "icmp", "format": "raw", "flush_frequency": 10}`
var badConfig4 = `{"address": "0.0.0.0", "port": 9999999, "protocol": "tcp", "format": "raw", "flush_frequency": 10}`
var badConfig5 = `{"address": "localhost", "port": 8443, "protocol": "tcp", "format": "raw", "flush_frequency": 10}`

func TestValidate(t *testing.T) {
	arr := []string{config1, config2, config3, config4, config5}
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
	arr := []string{config1, config2, config3, config4, config5}
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
