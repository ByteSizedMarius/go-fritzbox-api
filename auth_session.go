package go_fritzbox_api

/*
*
* The MIT License (MIT)
* Copyright (c) 2015 Philipp Franke
*
* Modified by ByteSizedMarius (2022)
* Code remains under MIT license (can be found in the LICENSE file.)
*
 */

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	"unicode/utf16"
)

const (
	// DefaultSid is an invalid session in order to perform and
	// identify logouts.
	DefaultSid = "0000000000000000"
	// DefaultExpires is the amount of time of inactivity before
	// the FRITZ!Box automatically closes a session.
	DefaultExpires = 10 * time.Minute
)

var (
	// ErrInvalidCred is the error returned by auth when
	// login attempt is not successful.
	ErrInvalidCred = errors.New("fritzbox: invalid credentials")

	// ErrExpiredSess means that Client was too long inactive.
	ErrExpiredSess = errors.New("fritzbox: session expired")
)

// session represents a FRITZ!Box session
type session struct {
	client *Client

	XMLName   xml.Name      `xml:"SessionInfo"`
	Sid       string        `xml:"SID"`
	Challenge string        `xml:"Challenge"`
	BlockTime time.Duration `xml:"BlockTime"`

	// Rights' representation is a little tricky
	RightsName   []string `xml:"Rights>Name"`
	RightsAccess []int8   `xml:"Rights>Access"`

	// Session expires after 10 minutes
	Expires time.Time `xml:"-"`
}

// newSession returns a new FRITZ!Box session.
func newSession(c *Client) *session {
	return &session{
		Sid:    DefaultSid,
		client: c,
	}
}

// open retrieves the challenge from FRITZ!Box.
func (s *session) open() error {
	return s.client.requestAndDecode(http.MethodGet, "login_sid.lua", nil, s)
}

// auth sends the Response (Challenge-Response) to the FRITZ!Box and
// returns an error, if any.
func (s *session) auth(username, password string) error {
	cr, err := computeResponse(s.Challenge, password)
	if err != nil {
		return err
	}

	err = s.client.requestAndDecode(http.MethodPost, "login_sid.lua", url.Values{"username": {username}, "response": {cr}}, s)
	if err != nil {
		return err
	}

	// Is login attempt successful?
	if s.Sid == DefaultSid {
		return ErrInvalidCred
	}

	return s.refresh()
}

// close closes a session
func (s *session) close() {
	s.Sid = DefaultSid
}

// isExpired returns true if session is expired
func (s *session) isExpired() bool {
	return s.Expires.Before(time.Now())
}

// refresh updates expires
func (s *session) refresh() error {
	if s.isExpired() && (s.Expires != time.Time{}) {
		s.close()
		return ErrExpiredSess
	}
	s.Expires = time.Now().Add(DefaultExpires)
	return nil
}

// ComputeResponse generates a response for challenge-response auth
// with the given challenge and secret. It returns the reponse and
// an error, if any.
func computeResponse(challenge, secret string) (string, error) {
	buf := new(bytes.Buffer)
	h := md5.New()

	chars := utf16.Encode([]rune(fmt.Sprintf("%s-%s", challenge, secret)))

	for _, char := range chars {
		// According to AVM's technical notes: unicode codepoints
		// above 255 needs to be converted to "." (0x2e 0x00 in UTF-16LE)
		if char > 255 {
			char = 0x2e
		}

		err := binary.Write(buf, binary.LittleEndian, char)
		if err != nil {
			return "", err
		}
	}

	_, _ = io.Copy(h, buf)
	r := fmt.Sprintf("%s-%x", challenge, h.Sum(nil))

	return r, nil
}

func (s *session) String() string {
	return s.Sid
}
