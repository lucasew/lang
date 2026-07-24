package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckErrorRate(t *testing.T) {
	// disabled rate
	require.NoError(t, CheckErrorRate(100, 30, 0, "English", 100))
	// wordCounter ≤ 25: never trips
	require.NoError(t, CheckErrorRate(100, 25, 0.1, "English", 100))
	// under rate
	require.NoError(t, CheckErrorRate(1, 100, 0.5, "English", 100))
	// over rate after >25 words
	err := CheckErrorRate(50, 30, 0.1, "English", 200)
	require.Error(t, err)
	var e *ErrorRateTooHighException
	require.ErrorAs(t, err, &e)
	require.Contains(t, e.Error(), "10%")
	require.Contains(t, e.Error(), "English")
	require.Contains(t, e.Error(), "200")
}

func TestErrorRateTooHighException(t *testing.T) {
	err := NewErrorRateTooHighException("too many errors")
	require.EqualError(t, err, "too many errors")
}
