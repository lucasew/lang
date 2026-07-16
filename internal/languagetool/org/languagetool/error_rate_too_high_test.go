package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorRateTooHighException(t *testing.T) {
	err := NewErrorRateTooHighException("too many errors")
	require.EqualError(t, err, "too many errors")
}
