package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDateCheckFilter(t *testing.T) {
	f := NewDateCheckFilter()
	require.NotNil(t, f)
	_, err := f.AcceptRuleMatch(map[string]string{"year": "2014"})
	require.Error(t, err)
}
