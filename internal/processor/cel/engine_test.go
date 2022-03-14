package cel

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

var logEntry *log.Entry

var event1 = `{"code": 400, "status": "invalid request", "data": {"errors": ["invalid page", "invalid scope"]}}`
var event2 = `{"code": 200, "status": "success", "data": {"message": "hello world"}}`
var event3 = `{"code": 500, "error": "server error"}`
var event4 = `{"code": 400, "status": "invalid request", "data": {"errors": ["invalid header"]}}`

var rule1 = `event.code == 200`
var rule2 = `event.code == 200 || event.code == 400`
var rule3 = `event.code == 200 && has(event.data) && has(event.data.message) && event.data.message == "hello world"`
var rule4 = `event.code == 400 && has(event.data) && has(event.data.errors) && event.data.errors.exists(x, x == "invalid page")`

func init() {
	log.SetLevel(log.DebugLevel)
	logEntry = log.WithField("processor", "cel")
}

func TestRuleDetection(t *testing.T) {
	value := ruleDetection(event1, []string{rule1}, logEntry)
	assert.False(t, value, "event1 code 400 should not trigger rule for code 200")
	value2 := ruleDetection(event2, []string{rule1}, logEntry)
	assert.Truef(t, value2, "event2 code 200 should trigger rule for code 200")
	value3 := ruleDetection(event3, []string{rule1}, logEntry)
	assert.False(t, value3, "event3 code 500 should not trigger rule for code 200")
}

func TestRuleDetection2(t *testing.T) {
	value := ruleDetection(event1, []string{rule2}, logEntry)
	assert.Truef(t, value, "event1 code 400 should trigger rule for code 200 or 400")
	value2 := ruleDetection(event2, []string{rule2}, logEntry)
	assert.Truef(t, value2, "event2 code 200 should trigger rule for code 200 or 400")
	value3 := ruleDetection(event3, []string{rule2}, logEntry)
	assert.Falsef(t, value3, "event3 code 500 should not trigger rule for code 200 or 400")
}

func TestRuleDetection3(t *testing.T) {
	value := ruleDetection(event1, []string{rule3}, logEntry)
	assert.Falsef(t, value, "event1 should not trigger rule")
	value2 := ruleDetection(event2, []string{rule3}, logEntry)
	assert.Truef(t, value2, "event2 should trigger rule")
	value3 := ruleDetection(event3, []string{rule3}, logEntry)
	assert.Falsef(t, value3, "event3 should not trigger rule")
}

func TestRuleDetection4(t *testing.T) {
	value := ruleDetection(event1, []string{rule4}, logEntry)
	assert.Truef(t, value, "event1 should trigger rule")
	value2 := ruleDetection(event2, []string{rule4}, logEntry)
	assert.Falsef(t, value2, "event2 should not trigger rule")
	value3 := ruleDetection(event3, []string{rule4}, logEntry)
	assert.Falsef(t, value3, "event3 should not trigger rule")
	value4 := ruleDetection(event4, []string{rule4}, logEntry)
	assert.Falsef(t, value4, "event4 should not trigger rule")
}

func TestRuleDetection5(t *testing.T) {
	value := ruleDetection(event1, []string{rule3, rule4}, logEntry)
	assert.Truef(t, value, "event1 should trigger rule")
	value2 := ruleDetection(event2, []string{rule3, rule4}, logEntry)
	assert.Truef(t, value2, "event2 should trigger rule")
	value3 := ruleDetection(event3, []string{rule3, rule4}, logEntry)
	assert.Falsef(t, value3, "event3 should not trigger rule")
}

func TestValidateRule(t *testing.T) {
	value := validateRule(rule1)
	assert.Nilf(t, value, "rule1 should be a valid rule and return no errors")
	value2 := validateRule(rule2)
	assert.Nilf(t, value2, "rule2 should be a valid rule and return no errors")
	value3 := validateRule(rule3)
	assert.Nilf(t, value3, "rule3 should be a valid rule and return no errors")
	value4 := validateRule(rule4)
	assert.Nilf(t, value4, "rule4 should be a valid rule and return no errors")
}

func TestValidateRule2(t *testing.T) {
	value := validateRule(`event === "hi"`)
	assert.Errorf(t, value, "should return an error for an invalid rule")
	value2 := validateRule(`event ||| "hi"`)
	assert.Errorf(t, value2, "should return an error for an invalid rule")
	value3 := validateRule(`event "hi"`)
	assert.Errorf(t, value3, "should return an error for an invalid rule")
}
