package msgraph

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

type authResponse struct {
	TokenType    string      `json:"token_type"`
	ExpiresIn    json.Number `json:"expires_in"`
	ExtExpiresIn json.Number `json:"ext_expires_in"`
	AccessToken  string      `json:"access_token"`
}

type Client struct {
	TenantId           string `json:"tenant_id"`
	ClientId           string `json:"client_id"`
	ClientSecret       string `json:"client_secret"`
	AccessToken        string `json:"access_token"`
	restyClient        *resty.Client
	scope              string
	accessTokenExpires time.Time
	logger             *log.Entry
	graphEndpoint      string
	loginEndpoint      string
}

type GraphListResponse struct {
	Context  string        `json:"@odata.context"`
	NextLink string        `json:"@odata.nextLink"`
	Value    []interface{} `json:"value"`
}

func NewClient(tenantId, clientId, clientSecret, scope string) *Client {
	baseLogger := log.New()
	return &Client{
		TenantId:           tenantId,
		ClientId:           clientId,
		ClientSecret:       clientSecret,
		AccessToken:        "",
		restyClient:        resty.New().SetRetryCount(3).SetRetryWaitTime(2 * time.Second).SetRetryMaxWaitTime(10 * time.Second),
		scope:              scope,
		accessTokenExpires: time.Now(),
		logger:             log.NewEntry(baseLogger),
		graphEndpoint:      "https://graph.microsoft.com",
		loginEndpoint:      "https://login.microsoftonline.com",
	}
}

func (client *Client) SetLogger(logger *log.Entry) error {
	// Handle nil logger
	if logger == nil {
		return fmt.Errorf("cannot set a null logger")
	}
	client.logger = logger
	return nil
}

func (client *Client) SetGraphEndpoint(graphEndpoint string) error {
	_, err := url.Parse(graphEndpoint)
	if err != nil {
		return fmt.Errorf("issue parsing endpoint: %s", err)
	}
	client.graphEndpoint = graphEndpoint
	return nil
}

func (client *Client) SetLoginEndpoint(loginEndpoint string) error {
	_, err := url.Parse(loginEndpoint)
	if err != nil {
		return fmt.Errorf("issue parsing endpoint: %s", err)
	}
	client.loginEndpoint = loginEndpoint
	return nil
}

func (client *Client) Ping() bool {
	err := client.login()
	return err == nil
}

func (client *Client) login() error {
	params := url.Values{}
	params.Set("scope", client.scope)
	params.Set("client_id", client.ClientId)
	params.Set("client_secret", client.ClientSecret)
	params.Set("grant_type", "client_credentials")
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	// Build login URI
	uri, err := client.buildLoginUri(fmt.Sprintf("/%s/oauth2/v2.0/token", client.TenantId))
	if err != nil {
		return fmt.Errorf("issue building login URI: %s", err)
	}

	// Conduct request
	_, body, err := client.makeCallSafe(uri, nil, params, http.MethodPost, headers)
	if err != nil {
		return fmt.Errorf("error in login request: %v", err)
	}

	var res authResponse
	err = json.Unmarshal(body, &res)

	// Handle error
	if err != nil {
		return fmt.Errorf("issue unmarshalling response body: %s", err)
	}

	client.AccessToken = res.AccessToken
	i, err := res.ExpiresIn.Int64()
	if err != nil {
		client.accessTokenExpires = time.Now().Add(time.Minute * time.Duration(29))
	} else {
		client.accessTokenExpires = time.Now().Add(time.Second * time.Duration(i))
	}

	return nil
}

func (client *Client) makeCall(uri string, body interface{}, params url.Values, method string, headers map[string]string) (*resty.Response, []byte, error) {
	if time.Now().After(client.accessTokenExpires) || client.AccessToken == "" {
		err := client.login()
		if err != nil {
			return nil, nil, err
		}
	}

	headers["Authorization"] = fmt.Sprintf("Bearer %s", client.AccessToken)

	return client.makeCallSafe(uri, body, params, method, headers)
}

func (client *Client) makeCallStream(uri string, body interface{}, params url.Values, method string, headers map[string]string) (*resty.Response, io.ReadCloser, error) {
	if time.Now().After(client.accessTokenExpires) || client.AccessToken == "" {
		err := client.login()
		if err != nil {
			return nil, nil, err
		}
	}

	headers["Authorization"] = fmt.Sprintf("Bearer %s", client.AccessToken)

	return client.makeCallStreamSafe(uri, body, params, method, headers)
}

func (client *Client) makeCallSafe(uri string, body interface{}, params url.Values, method string, headers map[string]string) (*resty.Response, []byte, error) {
	// Setup client
	tmpClient := client.restyClient.R().SetHeaders(headers)

	// Set json if no content type set
	if _, exists := headers["Content-Type"]; !exists {
		tmpClient.SetHeader("Content-Type", "application/json")
	}

	// Conduct request
	doRequest := func() (*resty.Response, error) {
		graphUrl, err := url.Parse(uri)
		if err != nil {
			return nil, err
		}
		switch {
		case method == http.MethodGet:
			// Conduct request
			graphUrl.RawQuery = params.Encode()
			return tmpClient.Get(graphUrl.String())
		case method == http.MethodPost:
			if tmpClient.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
				tmpClient.SetBody(params.Encode())
			} else {
				tmpClient.SetBody(body)
			}
			return tmpClient.Post(uri)
		case method == http.MethodPut:
			if tmpClient.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
				tmpClient.SetBody(params.Encode())
			} else {
				tmpClient.SetBody(body)
			}
			return tmpClient.Put(uri)
		case method == http.MethodDelete:
			return tmpClient.Delete(uri)
		}
		return nil, fmt.Errorf("invalid method")
	}

	// Get response
	response, err := doRequest()
	if err != nil {
		return response, nil, err
	}

	// Handle invalid responses
	if response.IsError() {
		return response, nil, fmt.Errorf("invalid response: %s", response.Status())
	}

	return response, response.Body(), nil
}

func (client *Client) makeCallStreamSafe(uri string, body interface{}, params url.Values, method string, headers map[string]string) (*resty.Response, io.ReadCloser, error) {
	// Setup client
	tmpClient := client.restyClient.R().SetHeaders(headers).SetDoNotParseResponse(true)

	// Set json if no content type set
	if _, exists := headers["Content-Type"]; !exists {
		tmpClient.SetHeader("Content-Type", "application/json")
	}

	// Conduct request
	doRequest := func() (*resty.Response, error) {
		graphUrl, err := url.Parse(uri)
		if err != nil {
			return nil, err
		}
		switch {
		case method == http.MethodGet:
			// Conduct request
			graphUrl.RawQuery = params.Encode()
			return tmpClient.Get(graphUrl.String())
		case method == http.MethodPost:
			if tmpClient.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
				tmpClient.SetBody(params.Encode())
			} else {
				tmpClient.SetBody(body)
			}
			return tmpClient.Post(uri)
		case method == http.MethodPut:
			if tmpClient.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
				tmpClient.SetBody(params.Encode())
			} else {
				tmpClient.SetBody(body)
			}
			return tmpClient.Put(uri)
		case method == http.MethodDelete:
			return tmpClient.Delete(uri)
		}
		return nil, fmt.Errorf("invalid method")
	}

	// Get response
	response, err := doRequest()
	if err != nil {
		return response, nil, err
	}

	// Handle invalid responses
	if response.IsError() {
		return response, nil, fmt.Errorf("invalid response: %s", response.Status())
	}

	return response, response.RawBody(), nil
}

func (client *Client) buildGraphUri(uriPath string) (string, error) {
	u, err := url.Parse(client.graphEndpoint)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, uriPath)
	return u.String(), nil
}

func (client *Client) buildLoginUri(uriPath string) (string, error) {
	u, err := url.Parse(client.loginEndpoint)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, uriPath)
	return u.String(), nil
}
