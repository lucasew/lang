package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseLimitEnforcementMode(t *testing.T) {
	require.Equal(t, LimitEnforcementDisabled, ParseLimitEnforcementMode(nil))
	z := 0
	require.Equal(t, LimitEnforcementDisabled, ParseLimitEnforcementMode(&z))
	neg := -1
	require.Equal(t, LimitEnforcementDisabled, ParseLimitEnforcementMode(&neg))
	one := 1
	require.Equal(t, LimitEnforcementDisabled, ParseLimitEnforcementMode(&one))
	two := 2
	require.Equal(t, LimitEnforcementPerDay, ParseLimitEnforcementMode(&two))
	nine := 9
	require.Equal(t, LimitEnforcementDisabled, ParseLimitEnforcementMode(&nine))
	require.Equal(t, 1, LimitEnforcementDisabled.ID())
	require.Equal(t, 2, LimitEnforcementPerDay.ID())
}
