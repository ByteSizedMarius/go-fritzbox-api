package fritzbox

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
	"unicode/utf16"
)

const (
	defaultSID    = "0000000000000000"
	sessionExpiry = 10 * time.Minute
)

var ErrInvalidCredentials = errors.New("invalid credentials")

// session represents an authenticated FRITZ!Box session.
type session struct {
	client *Client

	sid       string
	challenge string
	blockTime time.Duration
	expires   time.Time

	rightsName   []string
	rightsAccess []int8
}

// sessionResponse is the XML structure returned by login_sid.lua
type sessionResponse struct {
	XMLName   xml.Name `xml:"SessionInfo"`
	SID       string   `xml:"SID"`
	Challenge string   `xml:"Challenge"`
	BlockTime int      `xml:"BlockTime"`

	RightsName   []string `xml:"Rights>Name"`
	RightsAccess []int8   `xml:"Rights>Access"`
}

func newSession(c *Client) *session {
	return &session{
		sid:    defaultSID,
		client: c,
	}
}

// open retrieves the challenge from FRITZ!Box.
func (s *session) open() error {
	var resp sessionResponse
	err := s.client.AhaRequestXML(http.MethodGet, "login_sid.lua", nil, &resp)
	if err != nil {
		return fmt.Errorf("get challenge: %w", err)
	}

	s.challenge = resp.Challenge
	s.blockTime = time.Duration(resp.BlockTime) * time.Second
	return nil
}

// auth sends the challenge response and completes authentication.
func (s *session) auth() error {
	response, err := computeChallengeResponse(s.challenge, s.client.Password)
	if err != nil {
		return fmt.Errorf("compute response: %w", err)
	}

	var resp sessionResponse
	err = s.client.AhaRequestXML(
		http.MethodPost,
		"login_sid.lua",
		Values{"username": s.client.Username, "response": response},
		&resp,
	)
	if err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	if resp.SID == defaultSID {
		return ErrInvalidCredentials
	}

	s.sid = resp.SID
	s.expires = time.Now().Add(sessionExpiry)
	s.rightsName = resp.RightsName
	s.rightsAccess = resp.RightsAccess
	return nil
}

func (s *session) close() {
	s.sid = defaultSID
}

func (s *session) isExpired() bool {
	return time.Now().After(s.expires)
}

func (s *session) String() string {
	return fmt.Sprintf("Session{sid: %s, expires: %s}", s.sid, s.expires.Format(time.RFC3339))
}

// computeChallengeResponse generates the MD5-based challenge response.
// FRITZ!Box uses UTF-16LE encoding with codepoints >255 replaced by '.'.
func computeChallengeResponse(challenge, password string) (string, error) {
	buf := new(bytes.Buffer)
	h := md5.New()

	chars := utf16.Encode([]rune(challenge + "-" + password))
	for _, char := range chars {
		if char > 255 {
			char = 0x2e
		}
		if err := binary.Write(buf, binary.LittleEndian, char); err != nil {
			return "", err
		}
	}

	if _, err := io.Copy(h, buf); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%x", challenge, h.Sum(nil)), nil
}
