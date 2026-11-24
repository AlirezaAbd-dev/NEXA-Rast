// Package assets embeds and builds the CSS/JS RTL payload injected into Antigravity 2.0.
package assets

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed rtl.css
var rtlCSS string

//go:embed rtl.js
var rtlJS string

var (
	payloadOnce sync.Once
	payloadJS   string
	payloadErr  error
)

// PayloadJS returns the self-contained JavaScript payload with the RTL stylesheet inlined.
// The result is constructed once and cached for subsequent CDP injections.
func PayloadJS() (string, error) {
	payloadOnce.Do(func() {
		encodedCSS, err := json.Marshal(rtlCSS)
		if err != nil {
			payloadErr = fmt.Errorf("encode rtl css: %w", err)
			return
		}
		payloadJS = strings.Replace(rtlJS, "__NEXA_RAST_CSS__", string(encodedCSS), 1)
	})
	if payloadErr != nil {
		return "", payloadErr
	}
	return payloadJS, nil
}
