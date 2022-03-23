package msgraph

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (client *Client) SecurityAlerts(params url.Values) (*GraphListResponse, error) {
	headers := make(map[string]string, 0)

	// Build URI
	securityAlertsUri, err := client.buildGraphUri("/v1.0/security/alerts")
	if err != nil {
		return nil, fmt.Errorf("issue building security alerts URI: %s", err)
	}

	// Get alerts
	res, body, err := client.makeCall(securityAlertsUri, nil, params, http.MethodGet, headers)
	// Conduct request
	if err != nil {
		return nil, fmt.Errorf("issue conducting request: %s", err)
	}

	// Log if partial responses were returned
	if res.StatusCode() == 206 {
		client.logger.Warnf("partial response returned: %s", strings.Join(res.Header().Values("Warning"), "; "))
	}

	// Unmarshal body
	var alertsResponse GraphListResponse
	err = json.Unmarshal(body, &alertsResponse)
	if err != nil {
		return nil, fmt.Errorf("issue unmarshalling request body into struct: %s", err)
	}

	return &alertsResponse, nil
}
