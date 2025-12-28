// Package fritzbox provides a client for interacting with AVM FRITZ!Box routers.
package fritzbox

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var defaultHeaders = http.Header{
	"Content-Type":    {"application/x-www-form-urlencoded"},
	"Accept-Encoding": {"gzip, deflate"},
}

// Client handles authentication and communication with the FRITZ!Box.
type Client struct {
	BaseUrl  string
	Username string
	Password string

	session *session
	baseURL *url.URL
	http    *http.Client
}

// New creates a new client with the given credentials.
// Call Connect() to authenticate with the FRITZ!Box.
func New(username, password string) *Client {
	return &Client{
		Username: username,
		Password: password,
	}
}

// Connect initializes and authenticates the client.
// If already connected, the existing session is closed first.
func (c *Client) Connect() error {
	if c.IsConnected() {
		c.Close()
	}

	if c.BaseUrl == "" {
		c.BaseUrl = "http://192.168.178.1/"
	} else if !strings.HasPrefix(c.BaseUrl, "http") {
		return fmt.Errorf("base url must start with http(s)://")
	}

	var err error
	c.baseURL, err = url.Parse(c.BaseUrl)
	if err != nil {
		return fmt.Errorf("parse base url: %w", err)
	}

	if c.http == nil {
		c.http = http.DefaultClient
	}

	return c.authenticate()
}

// Close terminates the session and releases resources.
func (c *Client) Close() {
	if c.session != nil {
		c.session.close()
		c.session = nil
	}
	if c.http != nil {
		c.http.CloseIdleConnections()
		c.http = nil
	}
}

// IsConnected returns true if the client has an active session.
func (c *Client) IsConnected() bool {
	return c.http != nil
}

// SID returns the current session ID.
func (c *Client) SID() string {
	if c.session == nil {
		return ""
	}
	return c.session.sid
}

func (c *Client) String() string {
	if c.session == nil {
		return "Client{not connected}"
	}
	return c.session.String()
}

// SetHTTPClient sets a custom HTTP client. Must be called before Connect().
func (c *Client) SetHTTPClient(client *http.Client) {
	c.http = client
}

// IsExpired returns true if the session has expired due to inactivity.
func (c *Client) IsExpired() bool {
	return c.session == nil || c.session.isExpired()
}

// CheckExpiry reconnects if the session is expired or not connected.
func (c *Client) CheckExpiry() error {
	if !c.IsConnected() || c.IsExpired() {
		return c.Connect()
	}
	return nil
}

func (c *Client) authenticate() error {
	if c.session == nil {
		c.session = newSession(c)
	}

	if err := c.session.open(); err != nil {
		return err
	}
	return c.session.auth()
}

// AHA HTTP Interface methods - used by aha package.
// These use form-encoded bodies and SID as query/form parameter.

// AhaRequest sends a form-encoded request to the FRITZ!Box.
// GET: data as query parameters. POST: data as form body.
func (c *Client) AhaRequest(method, path string, data Values) (*http.Response, error) {
	u, err := c.resolveURL(path)
	if err != nil {
		return nil, err
	}

	var body io.Reader
	if method == http.MethodPost && data != nil {
		body = strings.NewReader(data.Encode())
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header = defaultHeaders.Clone()

	if method == http.MethodGet && data != nil {
		q := req.URL.Query()
		for k, v := range data {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		resp.Body.Close()
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}

	return resp, nil
}

// AhaRequestString sends a request and returns the response body as a string.
func (c *Client) AhaRequestString(method, path string, data Values) (int, string, error) {
	resp, err := c.AhaRequest(method, path, data)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("read response: %w", err)
	}

	return resp.StatusCode, strings.TrimSpace(string(body)), nil
}

// AhaRequestXML sends a request and decodes the XML response into target.
func (c *Client) AhaRequestXML(method, path string, data Values, target any) error {
	resp, err := c.AhaRequest(method, path, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return xml.NewDecoder(resp.Body).Decode(target)
}

func (c *Client) resolveURL(path string) (*url.URL, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("parse path: %w", err)
	}
	return c.baseURL.ResolveReference(rel), nil
}

// REST API methods - used by smarthome and unsafe packages.
// These use JSON bodies and Authorization header instead of form-encoded data.

// RestRequest sends a JSON request to the FRITZ!Box REST API.
// Body is JSON-marshaled for PUT/POST requests. Pass nil for GET/DELETE.
func (c *Client) RestRequest(method, path string, body any) ([]byte, int, error) {
	u, err := c.resolveURL(path)
	if err != nil {
		return nil, 0, err
	}

	var bodyReader io.Reader
	if body != nil && (method == http.MethodPut || method == http.MethodPost) {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = strings.NewReader(string(jsonBytes))
	}

	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "AVM-SID "+c.SID())
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// RestGet sends a GET request to the REST API.
func (c *Client) RestGet(path string) ([]byte, int, error) {
	return c.RestRequest(http.MethodGet, path, nil)
}

// RestPut sends a PUT request with a JSON body to the REST API.
func (c *Client) RestPut(path string, body any) ([]byte, int, error) {
	return c.RestRequest(http.MethodPut, path, body)
}

// RestPost sends a POST request with a JSON body to the REST API.
func (c *Client) RestPost(path string, body any) ([]byte, int, error) {
	return c.RestRequest(http.MethodPost, path, body)
}

// RestDelete sends a DELETE request to the REST API.
func (c *Client) RestDelete(path string) ([]byte, int, error) {
	return c.RestRequest(http.MethodDelete, path, nil)
}
