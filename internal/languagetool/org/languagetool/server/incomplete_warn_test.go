package server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApiV2_AllowIncompleteResultsWarning(t *testing.T) {
	api := NewApiV2(nil, nil)
	big := strings.Repeat("word ", 30_000) // > 100k chars
	r, err := api.Handle("check", map[string]string{
		"language":               "en",
		"text":                   big,
		"allowIncompleteResults": "true",
	})
	require.NoError(t, err)
	require.Contains(t, r.Body, "incomplete")
	require.Contains(t, r.Body, "warnings")
}
