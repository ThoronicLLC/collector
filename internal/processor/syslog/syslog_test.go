package syslog

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseRaw(t *testing.T) {
	msg1 := `<134>Apr 13 10:23:46 demo-host CEF:0|archer|archer|1.1.15.20|archer:heartbeat|Heartbeat|0|dvc=127.0.0.1 rt=1649820106246 cat=archer:SYS`
	msg2 := `<134>Apr 13 10:52:19 demo-host CEF:0|archer|archer|1.1.15.20|archer:access|Access event|10|msg=Source Port\=58326 Ports count\=1 cs1Label=Source URL rt=1649820149781 cs1=https://archer.local/demo src=192.168.1.11 destinationServiceName=PORT_SCAN externalId=2034604 Name dvc=192.168.1.50 suser=admin cat=archer:alerts shost=someone.local dhost=archer.local`
	msg3 := `<191>1 2022-04-13T11:21:57.586018+07:00 demo-host demo-app 666 12543 [555] {"source": "192.168.1.11", "destination": "192.168.1.15", "message": "Port scan was detected"}`

	expectedRaw1 := `Apr 13 10:23:46 demo-host CEF:0|archer|archer|1.1.15.20|archer:heartbeat|Heartbeat|0|dvc=127.0.0.1 rt=1649820106246 cat=archer:SYS`
	result1, err := parseRaw(msg1)
	assert.Nil(t, err, "should have received a successful result from parseRaw(msg1)")
	assert.Equal(t, expectedRaw1, result1)

	expectedRaw2 := `Apr 13 10:52:19 demo-host CEF:0|archer|archer|1.1.15.20|archer:access|Access event|10|msg=Source Port\=58326 Ports count\=1 cs1Label=Source URL rt=1649820149781 cs1=https://archer.local/demo src=192.168.1.11 destinationServiceName=PORT_SCAN externalId=2034604 Name dvc=192.168.1.50 suser=admin cat=archer:alerts shost=someone.local dhost=archer.local`
	result2, err := parseRaw(msg2)
	assert.Nil(t, err, "should have received a successful result from parseRaw(msg2)")
	assert.Equal(t, expectedRaw2, result2)

	expectedRaw3 := `1 2022-04-13T11:21:57.586018+07:00 demo-host demo-app 666 12543 [555] {"source": "192.168.1.11", "destination": "192.168.1.15", "message": "Port scan was detected"}`
	result3, err := parseRaw(msg3)
	assert.Nil(t, err, "should have received a successful result from parseRaw(msg2)")
	assert.Equal(t, expectedRaw3, result3)
}

func TestParseRfc3164(t *testing.T) {
	msg1 := `<134>Apr 13 10:23:46 demo-host CEF:0|archer|archer|1.1.15.20|archer:heartbeat|Heartbeat|0|dvc=127.0.0.1 rt=1649820106246 cat=archer:SYS`
	msg2 := `<134>Apr 13 10:52:19 demo-host CEF:0|archer|archer|1.1.15.20|archer:access|Access event|10|msg=Source Port\=58326 Ports count\=1 cs1Label=Source URL rt=1649820149781 cs1=https://archer.local/demo src=192.168.1.11 destinationServiceName=PORT_SCAN externalId=2034604 Name dvc=192.168.1.50 suser=admin cat=archer:alerts shost=someone.local dhost=archer.local`

	expectedRfc1 := `CEF:0|archer|archer|1.1.15.20|archer:heartbeat|Heartbeat|0|dvc=127.0.0.1 rt=1649820106246 cat=archer:SYS`
	result1, err := parseRfc3164(msg1)
	assert.Nil(t, err, "should have received a successful result from parseRfc5424(msg1)")
	assert.Equal(t, expectedRfc1, result1)

	expectedRfc2 := `CEF:0|archer|archer|1.1.15.20|archer:access|Access event|10|msg=Source Port\=58326 Ports count\=1 cs1Label=Source URL rt=1649820149781 cs1=https://archer.local/demo src=192.168.1.11 destinationServiceName=PORT_SCAN externalId=2034604 Name dvc=192.168.1.50 suser=admin cat=archer:alerts shost=someone.local dhost=archer.local`
	result2, err := parseRfc3164(msg2)
	assert.Nil(t, err, "should have received a successful result from parseRfc5424(msg2)")
	assert.Equal(t, expectedRfc2, result2)
}

func TestParseRfc5424(t *testing.T) {
	msg1 := `<191>1 2022-04-13T11:21:57.586018+07:00 demo-host demo-app 666 12543 [555] {"source": "192.168.1.11", "destination": "192.168.1.15", "message": "Port scan was detected"}`

	expectedRfc1 := `{"source": "192.168.1.11", "destination": "192.168.1.15", "message": "Port scan was detected"}`
	result1, err := parseRfc5424(msg1)
	assert.Nil(t, err, "should have received a successful result from parseRfc5424(msg1)")
	assert.Equal(t, expectedRfc1, result1)
}
