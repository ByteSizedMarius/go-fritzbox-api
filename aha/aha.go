// Package aha is a Go-Wrapper for the AHA HTTP Interface (AVM Home Automation HTTP Interface)
package aha

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/clbanning/mxj"
)

func unixStringToTime(unixString string) time.Time {
	i, _ := strconv.ParseInt(unixString, 10, 64)
	return time.Unix(i, 0)
}

func unmarshalKey(b []byte, key string, dest interface{}) error {
	mv, err := mxj.NewMapXml(b)
	if err != nil {
		return fmt.Errorf("parse XML: %w", err)
	}

	j, err := mv.Json(true)
	if err != nil {
		return fmt.Errorf("convert to JSON: %w", err)
	}

	tmp := map[string]json.RawMessage{}
	if err = json.Unmarshal(j, &tmp); err != nil {
		return fmt.Errorf("unmarshal JSON map: %w", err)
	}

	if err = json.Unmarshal(tmp[key], dest); err != nil {
		return fmt.Errorf("unmarshal key %q: %w", key, err)
	}
	return nil
}
