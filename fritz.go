package fritzbox

/*
*
* The MIT License (MIT)
* Copyright (c) 2015 Philipp Franke
*
* Modified by ByteSizedMarius (2022)
* Code remains under MIT license (can be found in the LICENSE file)
*
 */

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Client struct {
	BaseUrl  string
	Username string
	Password string

	session *session
	urlO    *url.URL
	client  *http.Client
}

func (c *Client) SetCustomHTTPClient(client *http.Client) {
	c.client = client
}

func (c *Client) Initialize() error {
	if c.isInitialized() {
		fmt.Println("Already initialized. Closing current session.")
		c.Close()
	}

	// set default base url
	if c.BaseUrl == "" {
		c.BaseUrl = "http://192.168.178.1/"
	}

	// check url validity
	if !strings.HasPrefix(c.BaseUrl, "http://") {
		return fmt.Errorf("base url has to start with http(s)://")
	}

	// convert url
	fu, err := url.Parse(c.BaseUrl)
	if err != nil {
		return err
	}
	c.urlO = fu

	// set http client
	if c.client == nil {
		c.client = http.DefaultClient
	}

	// authenticate
	return c.auth(c.Username, c.Password)
}

func (c *Client) SID() string {
	return c.session.Sid
}

func (c *Client) Close() {
	if c.isInitialized() {
		if c.session != nil {
			c.session.close()
		}

		c.client.CloseIdleConnections()
		c.client = nil
	}
}

// CustomRequest allows one to send a custom request. The method can be http.MethodPost or http.MethodGet.
// The urlPath should always only be the path (for example "data.lua"), get-queries are set via the data-field. See examples.
func (c *Client) CustomRequest(method string, urlPath string, data url.Values) (statusCode int, body string, err error) {
	if err = c.checkExpiry(); err != nil {
		return
	}

	resp, err := c.doRequest(method, urlPath, data)
	if err != nil {
		return
	}

	body, err = getBody(resp)
	statusCode = resp.StatusCode
	return
}

//
//
// client
//
//

func (c *Client) requestAndDecode(method string, urlStr string, data url.Values, mystruct interface{}) (err error) {
	resp, err := c.doRequest(method, urlStr, data)
	if err != nil {
		return err
	}

	return decode(resp, mystruct)
}

func (c *Client) doRequest(method string, urlStr string, data url.Values) (resp *http.Response, err error) {

	// refresh session potentially
	if c.session != nil {
		if err = c.session.refresh(); err != nil {
			return nil, err
		}
	}

	// build request url
	u, err := c.resolveUrl(urlStr)
	if err != nil {
		return
	}

	// if POST request, encode body accordingly
	var buf io.Reader = nil
	if method == http.MethodPost {
		buf = strings.NewReader(data.Encode())
	}

	// create request
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	// if GET, add query to url
	if method == http.MethodGet {
		q := req.URL.Query()
		for k, e := range data {
			q.Add(k, e[0])
		}
		req.URL.RawQuery = q.Encode()
	}

	// send request
	resp, err = c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if sc := resp.StatusCode; sc > 200 || sc > 299 {
		return nil, fmt.Errorf("invalid statuscode (%v)", resp.StatusCode)
	}

	return resp, nil
}

func (c *Client) resolveUrl(urlStr string) (u *url.URL, err error) {
	// build reqest
	rel, err := url.Parse(urlStr)
	if err != nil {
		return
	}
	return c.urlO.ResolveReference(rel), nil
}

func (c *Client) get(urlStr string) (req *http.Request, err error) {
	u, err := c.resolveUrl(urlStr)
	if err != nil {
		return
	}

	if c.session != nil {
		if c.session.Sid != DefaultSid {
			values := u.Query()
			values.Set("sid", c.session.Sid)
			u.RawQuery = values.Encode()
		}
	}

	req, err = http.NewRequest("GET", u.String(), nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	if err != nil {
		return nil, err
	}
	return req, nil
}

func (c Client) isInitialized() bool {
	return c.client != nil
}

func (c *Client) checkExpiry() (err error) {
	if c.isInitialized() && c.session.isExpired() || !c.isInitialized() {
		err = c.Initialize()
	}

	return
}

func (c *Client) String() string {
	return c.session.String()
}

// auth sends a auth request and returns an error, if any. session is stored
// in Client in order to perform requests with authentification.
func (c *Client) auth(username, password string) error {
	var s *session
	if c.session == nil {
		s = newSession(c)
		c.session = s
	} else {
		s = c.session
	}

	if err := s.open(); err != nil {
		return err
	}

	if err := s.auth(username, password); err != nil {
		return err
	}

	return nil
}

//
//
// Helper
//
//

func decode(resp *http.Response, mystruct interface{}) (err error) {
	if strings.Contains(resp.Header.Get("Content-Type"), "text/xml") {
		err = xml.NewDecoder(resp.Body).Decode(mystruct)
	} else if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		err = json.NewDecoder(resp.Body).Decode(mystruct)
	}

	return
}

func getBody(resp *http.Response) (body string, err error) {
	b, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return
	}

	body = string(b)
	return
}

func printRequestBody(req *http.Request) {
	if req.Body == nil {
		return
	}

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", b)
	fmt.Println()
}

func getUntil(main string, split string) string {
	return main[:strings.Index(main, split)]
}

func getFrom(main string, split string) string {
	return main[strings.Index(main, split):]
}

func getFromOffset(main string, split string, offset int) string {
	return main[strings.Index(main, split)+offset:]
}
