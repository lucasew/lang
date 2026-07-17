package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApiV2_BinarySpeller(t *testing.T) {
	if softEnglishUSDictPath() == "" {
		t.Skip("en_US.dict not available")
	}
	api := NewApiV2(nil, nil)
	r, err := api.Handle("check", map[string]string{
		"language": "en",
		"text":     "I recieve the book.",
	})
	require.NoError(t, err)
	require.Contains(t, r.Body, "MORFOLOGIK_RULE_EN_US")
	require.Contains(t, r.Body, "receive")
}
