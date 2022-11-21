package go_fritzbox_api

import (
	"encoding/json"
	"fmt"
	"time"
)

type HFInterface interface {
	String() string
	fromJSON(m map[string]json.RawMessage) (HFInterface, error)
}

// -------------------------------------------
//
// KeepAlive
//
// -------------------------------------------

type HFKeepAlive struct {
}

func (hfka HFKeepAlive) String() string {
	return ""
}

func (hfka HFKeepAlive) fromJSON(m map[string]json.RawMessage) (HFInterface, error) {
	return HFAlert{}, nil
}

// -------------------------------------------
//
// Alert
//
// -------------------------------------------

type HFAlert struct {
	State          string `json:"state"`
	LastChangeTime string `json:"lastalertchgtimestamp"`
}

func (hfa HFAlert) String() string {
	return fmt.Sprintf("{State: %s, Last Alert: %s (%s)}", hfa.State, hfa.GetLastAlertTimestamp().Format("02.01.2006 15:04:05"), hfa.LastChangeTime)
}

// GetLastAlertTimestamp converts the last change timestamp into go time struct
func (hfa HFAlert) GetLastAlertTimestamp() time.Time {
	return unixStringToTime(hfa.LastChangeTime)
}

// IsAlertActive returns true is alert is active
func (hfa HFAlert) IsAlertActive() bool {
	return hfa.State == "1"
}

func (hfa HFAlert) fromJSON(m map[string]json.RawMessage) (HFInterface, error) {
	err := json.Unmarshal(m["alert"], &hfa)
	if err != nil {
		return hfa, err
	}

	return hfa, nil
}

// -------------------------------------------
//
// OnOff
//
// -------------------------------------------

type HFOnOff struct {
}

func (hfoo HFOnOff) String() string {
	return ""
}

func (hfoo HFOnOff) fromJSON(m map[string]json.RawMessage) (HFInterface, error) {
	return HFOnOff{}, nil
}

// -------------------------------------------
//
// LevelControl
//
// -------------------------------------------

type HFLevelControl struct {
}

func (hflc HFLevelControl) String() string {
	return ""
}

func (hflc HFLevelControl) fromJSON(m map[string]json.RawMessage) (HFInterface, error) {
	return HFLevelControl{}, nil
}

// -------------------------------------------
//
// ColorControl
//
// -------------------------------------------

type HFColorControl struct {
}

func (hfcc HFColorControl) String() string {
	return ""
}

func (hfcc HFColorControl) fromJSON(m map[string]json.RawMessage) (HFInterface, error) {
	return HFColorControl{}, nil
}

// -------------------------------------------
//
// OpenClose
//
// -------------------------------------------

type HFOpenClose struct {
}

func (hfoc HFOpenClose) String() string {
	return ""
}

func (hfoc HFOpenClose) fromJSON(m map[string]json.RawMessage) (HFInterface, error) {
	return HFOpenClose{}, nil
}

// -------------------------------------------
//
// OpenCloseConfig
//
// -------------------------------------------

type HFOpenCloseConfig struct {
}

func (hfocc HFOpenCloseConfig) String() string {
	return ""
}

func (hfocc HFOpenCloseConfig) fromJSON(m map[string]json.RawMessage) (HFInterface, error) {
	return HFOpenCloseConfig{}, nil
}

// -------------------------------------------
//
// SimpleButton
//
// -------------------------------------------

type HFSimpleButton struct {
}

func (hfsb HFSimpleButton) String() string {
	return ""
}

func (hfsb HFSimpleButton) fromJSON(m map[string]json.RawMessage) (HFInterface, error) {
	return HFSimpleButton{}, nil
}

// -------------------------------------------
//
// SuotaUpdate
//
// -------------------------------------------

type HFSuotaUpdate struct {
}

func (hfsu HFSuotaUpdate) String() string {
	return ""
}

func (hfsu HFSuotaUpdate) fromJSON(m map[string]json.RawMessage) (HFInterface, error) {
	return HFSuotaUpdate{}, nil
}
