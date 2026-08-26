package server

// Twin of languagetool-server/src/test/java/org/languagetool/server/JwtTest.java
import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func checkDefaultLimits(t *testing.T, lim *UserLimits, expectJWT JwtContent) {
	t.Helper()
	require.NotNil(t, lim)
	require.Equal(t, math.MaxInt32, lim.MaxTextLength)
	require.Equal(t, int64(-1), lim.MaxCheckTimeMillis)
	require.False(t, lim.HasPremium)
	require.Nil(t, lim.DictionaryCacheSize)
	require.Nil(t, lim.PremiumUID)
	require.False(t, lim.SkipLimits)
	require.Nil(t, lim.RequestsPerDay)
	require.Equal(t, LimitEnforcementDisabled, lim.LimitEnforcement)
	require.Equal(t, expectJWT.IsValid, lim.JWT.IsValid)
	require.Equal(t, expectJWT.IsPremium, lim.JWT.IsPremium)
}

// Port of JwtTest.getLimitsWithJwtTokenTest
func TestJwt_GetLimitsWithJwtTokenTest(t *testing.T) {
	cfg := NewHTTPServerConfig()
	// defaults: MaxTextLengthAnonymous is MaxInt32, MaxCheckTime -1
	lim := GetLimitsWithJwtToken(cfg, "", "", "")
	checkDefaultLimits(t, lim, JwtNone)
}

// Port of JwtTest.getUserLimitsTest
func TestJwt_GetUserLimitsTest(t *testing.T) {
	cfg := NewHTTPServerConfig()
	params := map[string]string{}
	lim := GetUserLimits(params, cfg, "")
	// Java checkDefaults(userLimits, null) — JWT may be zero/none
	require.False(t, lim.HasPremium)
	require.Equal(t, math.MaxInt32, lim.MaxTextLength)

	auth := GetAuthHeader("Bearer: kdsajgtfoi43hjrt9i342htfg0eqhj0-49jrtfg9o0jnm32-0er34jghg908hn")
	require.NotEmpty(t, auth)
	withAuth := GetUserLimits(params, cfg, auth)
	checkDefaultLimits(t, withAuth, JwtNone)

	params["username"] = "user"
	params["tokenV2"] = "0815-token"
	withUser := GetUserLimits(params, cfg, auth)
	checkDefaultLimits(t, withUser, JwtNone)
}
