package fritzbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
)

// ValueFromJsonPath extracts a nested value from JSON using a path of keys.
func ValueFromJsonPath(body string, pathKeys []string) (v map[string]interface{}, err error) {
	err = json.Unmarshal([]byte(body), &v)
	if err != nil {
		var e *json.SyntaxError
		if errors.As(err, &e) {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		return
	}

	r, ok := v[pathKeys[0]].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("key %q not found or not a map", pathKeys[0])
	}
	for i := 1; i < len(pathKeys); i++ {
		r, ok = r[pathKeys[i]].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("key %q not found or not a map", pathKeys[i])
		}
	}

	return r, nil
}

// Pow returns n^m as int.
func Pow(n, m int) int {
	return int(math.Pow(float64(n), float64(m)))
}
