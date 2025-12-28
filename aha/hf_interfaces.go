package aha

import (
	"encoding/json"
	"fmt"
	"time"
)

type Interface interface {
	String() string
	fromJSON(m map[string]json.RawMessage) (Interface, error)
}

type KeepAlive struct {
}

func (hfka KeepAlive) String() string {
	return ""
}

func (hfka KeepAlive) fromJSON(m map[string]json.RawMessage) (Interface, error) {
	return KeepAlive{}, nil
}

type Alert struct {
	State          string `json:"state"`
	LastChangeTime string `json:"lastalertchgtimestamp"`
}

// String returns a string representation of the alert
func (hfa Alert) String() string {
	return fmt.Sprintf(
		"{State: %s, Last Alert: %s (%s)}",
		hfa.State, hfa.GetLastAlertTimestamp().Format("02.01.2006 15:04:05"), hfa.LastChangeTime,
	)
}

// GetLastAlertTimestamp converts the last change timestamp into go time struct
func (hfa Alert) GetLastAlertTimestamp() time.Time {
	return unixStringToTime(hfa.LastChangeTime)
}

// IsAlertActive returns true if alert is active
func (hfa Alert) IsAlertActive() bool {
	return hfa.State == "1"
}

// fromJSON parses the json into an Alert
func (hfa Alert) fromJSON(m map[string]json.RawMessage) (Interface, error) {
	err := json.Unmarshal(m["alert"], &hfa)
	return hfa, err
}

type OnOff struct {
}

func (hfoo OnOff) String() string {
	return ""
}

func (hfoo OnOff) fromJSON(m map[string]json.RawMessage) (Interface, error) {
	return OnOff{}, nil
}

type LevelControl struct {
}

func (hflc LevelControl) String() string {
	return ""
}

func (hflc LevelControl) fromJSON(m map[string]json.RawMessage) (Interface, error) {
	return LevelControl{}, nil
}

type ColorControl struct {
}

func (hfcc ColorControl) String() string {
	return ""
}

func (hfcc ColorControl) fromJSON(m map[string]json.RawMessage) (Interface, error) {
	return ColorControl{}, nil
}

type OpenClose struct {
}

func (hfoc OpenClose) String() string {
	return ""
}

func (hfoc OpenClose) fromJSON(m map[string]json.RawMessage) (Interface, error) {
	return OpenClose{}, nil
}

type OpenCloseConfig struct {
}

func (hfocc OpenCloseConfig) String() string {
	return ""
}

func (hfocc OpenCloseConfig) fromJSON(m map[string]json.RawMessage) (Interface, error) {
	return OpenCloseConfig{}, nil
}

type SimpleButton struct {
}

func (hfsb SimpleButton) String() string {
	return ""
}

func (hfsb SimpleButton) fromJSON(m map[string]json.RawMessage) (Interface, error) {
	return SimpleButton{}, nil
}

type HFSuotaUpdate struct {
}

func (hfsu HFSuotaUpdate) String() string {
	return ""
}

func (hfsu HFSuotaUpdate) fromJSON(m map[string]json.RawMessage) (Interface, error) {
	return HFSuotaUpdate{}, nil
}
