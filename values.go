package fritzbox

import "net/url"

// Values is a convenience type for building form-encoded request parameters.
// Used by AhaRequest methods; REST methods use JSON bodies instead.
type Values map[string]string

// Encode returns the values as a URL-encoded string.
func (v Values) Encode() string {
	return v.URLValues().Encode()
}

// URLValues converts to standard library url.Values.
func (v Values) URLValues() url.Values {
	uv := make(url.Values, len(v))
	for key, value := range v {
		uv.Set(key, value)
	}
	return uv
}
