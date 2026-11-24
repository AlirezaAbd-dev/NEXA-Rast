package assets

import (
	"strings"
	"testing"
)

func TestPayloadJS(t *testing.T) {
	tests := []struct {
		name     string
		contains string
		absent   string
	}{
		{
			name:     "inlines css and removes placeholder",
			contains: "unicode-bidi: plaintext;",
			absent:   "__NEXA_RAST_CSS__",
		},
		{
			name:     "includes idempotent guard",
			contains: "__nexaRastInstalled",
		},
		{
			name:     "keeps code blocks ltr",
			contains: "direction: ltr !important;",
		},
		{
			name:     "batches mutations with requestAnimationFrame",
			contains: "requestAnimationFrame",
		},
	}

	got, err := PayloadJS()
	if err != nil {
		t.Fatalf("PayloadJS() unexpected error: %v", err)
	}
	if got == "" {
		t.Fatal("PayloadJS() returned empty string")
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.contains != "" && !strings.Contains(got, tc.contains) {
				t.Errorf("PayloadJS() missing %q", tc.contains)
			}
			if tc.absent != "" && strings.Contains(got, tc.absent) {
				t.Errorf("PayloadJS() unexpectedly contains %q", tc.absent)
			}
		})
	}
}
