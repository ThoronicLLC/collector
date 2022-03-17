package variable_replacer

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

var currentTime = time.Now()
var currentYear = fmt.Sprintf("%s", currentTime.Format("2006"))
var currentMonth = fmt.Sprintf("%s", currentTime.Format("01"))
var currentMonthDay = fmt.Sprintf("%s", currentTime.Format("January"))
var currentDay = fmt.Sprintf("%s", currentTime.Format("02"))
var currentHour = fmt.Sprintf("%s", currentTime.Format("15"))
var currentMinute = fmt.Sprintf("%s", currentTime.Format("04"))
var currentSecond = fmt.Sprintf("%s", currentTime.Format("05"))
var currentTimezone = fmt.Sprintf("%s", currentTime.Format("Z07:00:00"))

func TestVariableReplacer(t *testing.T) {
	// Test year
	val := VariableReplacer(currentTime, "hello_%year%")
	expected := fmt.Sprintf("hello_%s", currentYear)
	assert.Truef(t, val == expected, fmt.Sprintf("expected: %s; got: %s", expected, val))

	// Test month
	val = VariableReplacer(currentTime, "hello_%month%")
	expected = fmt.Sprintf("hello_%s", currentMonth)
	assert.Truef(t, val == expected, fmt.Sprintf("expected: %s; got: %s", expected, val))

	// Test month name
	val = VariableReplacer(currentTime, "hello_%month_name%")
	expected = fmt.Sprintf("hello_%s", currentMonthDay)
	assert.Truef(t, val == expected, fmt.Sprintf("expected: %s; got: %s", expected, val))

	// Test day
	val = VariableReplacer(currentTime, "hello_%day%")
	expected = fmt.Sprintf("hello_%s", currentDay)
	assert.Truef(t, val == expected, fmt.Sprintf("expected: %s; got: %s", expected, val))

	// Test hour
	val = VariableReplacer(currentTime, "hello_%hour%")
	expected = fmt.Sprintf("hello_%s", currentHour)
	assert.Truef(t, val == expected, fmt.Sprintf("expected: %s; got: %s", expected, val))

	// Test minute
	val = VariableReplacer(currentTime, "hello_%minute%")
	expected = fmt.Sprintf("hello_%s", currentMinute)
	assert.Truef(t, val == expected, fmt.Sprintf("expected: %s; got: %s", expected, val))

	// Test second
	val = VariableReplacer(currentTime, "hello_%second%")
	expected = fmt.Sprintf("hello_%s", currentSecond)
	assert.Truef(t, val == expected, fmt.Sprintf("expected: %s; got: %s", expected, val))

	// Test timezone
	val = VariableReplacer(currentTime, "hello_%timezone%")
	expected = fmt.Sprintf("hello_%s", currentTimezone)
	assert.Truef(t, val == expected, fmt.Sprintf("expected: %s; got: %s", expected, val))

	// Test unix time
	beforeUnix := currentTime.Add(-1 * time.Hour).Unix()
	afterUnix := currentTime.Add(1 * time.Hour).Unix()
	currentUnixString := VariableReplacer(currentTime, "%unix%")
	currentUnixInt, err := strconv.Atoi(currentUnixString)
	assert.Nilf(t, err, "did not expect an error")
	currentUnix := int64(currentUnixInt)
	assert.Truef(t, (beforeUnix < currentUnix) && (currentUnix < afterUnix), fmt.Sprintf("expected: %s; got: %s", expected, val))
}
