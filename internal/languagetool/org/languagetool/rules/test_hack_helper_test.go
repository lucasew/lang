package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsTest(t *testing.T) {
	// under go test this should be true
	require.True(t, IsTest())
}
