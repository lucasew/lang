package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPServerConfigDefaults(t *testing.T) {
	c := NewHTTPServerConfig()
	require.Equal(t, DefaultPort, c.Port)
	require.True(t, c.SuggestionsEnabled)
	require.True(t, c.AnonymousAccessAllowed)

	require.NoError(t, c.ApplyArgs([]string{"--port", "9090", "--public", "--verbose"}))
	require.Equal(t, 9090, c.Port)
	require.True(t, c.PublicAccess)
	require.True(t, c.Verbose)

	c.ApplyProperties(map[string]string{
		"maxTextLength":               "1000",
		"requestLimit":                "50",
		"pipelineCaching":             "true",
		"maxPipelinePoolSize":         "5",
		"disabledRuleIds":             "FOO, BAR",
	})
	require.Equal(t, 1000, c.MaxTextLengthAnonymous)
	require.Equal(t, 50, c.RequestLimit)
	require.True(t, c.IsPipelineCachingEnabled())
	require.Equal(t, []string{"FOO", "BAR"}, c.DisabledRuleIDs)
}

func TestRequestCounter(t *testing.T) {
	c := NewRequestCounter()
	require.Equal(t, 1, c.IncrementRequestCount())
	c.IncrementHandleCount("1.1.1.1", 10)
	c.IncrementHandleCount("2.2.2.2", 11)
	require.Equal(t, 2, c.HandleCount())
	require.Equal(t, 2, c.DistinctIPs())
	c.DecrementHandleCount(10)
	require.Equal(t, 1, c.HandleCount())
	require.Equal(t, 1, c.DistinctIPs())
}

func TestErrorRequestLimiter(t *testing.T) {
	e := NewErrorRequestLimiter(2, 60)
	require.True(t, e.WouldAccessBeOkay("1.1.1.1"))
	e.LogAccess("1.1.1.1")
	e.LogAccess("1.1.1.1")
	require.False(t, e.WouldAccessBeOkay("1.1.1.1"))
	require.Error(t, e.CheckLimit("1.1.1.1"))
}

func TestPipelineFreezeAndPool(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.PipelineCaching = true
	cfg.MaxPipelinePoolSize = 4
	cfg.DisabledRuleIDs = []string{"GLOBAL_OFF"}

	pool := NewPipelinePool(cfg)
	settings := NewPipelineSettings("en", "anon")

	pl, err := pool.Borrow(settings)
	require.NoError(t, err)
	require.True(t, pl.IsFrozen())
	require.True(t, pl.IsRuleDisabled("GLOBAL_OFF"))
	require.Error(t, pl.DisableRule("X")) // frozen

	pool.Return(settings, pl)
	require.Equal(t, 1, pool.IdleCount(settings))

	pl2, err := pool.Borrow(settings)
	require.NoError(t, err)
	require.Same(t, pl, pl2)
	pool.Return(settings, pl2)
}

func TestPipelinePoolExhausted(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.PipelineCaching = true
	cfg.MaxPipelinePoolSize = 1
	pool := NewPipelinePool(cfg)
	s := NewPipelineSettings("de", "u")
	_, err := pool.Borrow(s)
	require.NoError(t, err)
	_, err = pool.Borrow(s)
	require.Error(t, err)
	var ue *UnavailableError
	require.ErrorAs(t, err, &ue)
}

func TestUserLimitsAndJwt(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.MaxTextLengthAnonymous = 100
	cfg.PremiumAlways = true
	cfg.MaxTextLengthPremium = 5000
	lim := DefaultUserLimits(cfg)
	require.True(t, lim.HasPremium)
	require.Equal(t, 5000, lim.MaxTextLength)
	require.False(t, JwtNone.IsValid)
	j := NewJwtContent(true, true, map[string]any{"sub": "u1"})
	require.True(t, j.IsPremium)
}

func TestLimitEnforcementMode(t *testing.T) {
	require.Equal(t, LimitEnforcementDisabled, ParseLimitEnforcementMode(nil))
	two := 2
	require.Equal(t, LimitEnforcementPerDay, ParseLimitEnforcementMode(&two))
	nine := 9
	require.Equal(t, LimitEnforcementDisabled, ParseLimitEnforcementMode(&nine))
}

func TestActiveRules(t *testing.T) {
	a := NewActiveRules()
	a.EnterPattern("RULE_A")
	a.EnterPattern("RULE_A")
	a.EnterSpellCheck("teh")
	require.Equal(t, 2, a.GetActivePatternRules()["RULE_A"])
	require.Equal(t, []string{"teh"}, a.GetActiveSpellChecks())
	a.LeavePattern("RULE_A")
	require.Equal(t, 1, a.GetActivePatternRules()["RULE_A"])
	a.LeaveSpellCheck("teh")
	require.Empty(t, a.GetActiveSpellChecks())
}

func TestConfidenceMapLoader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "conf-en.csv")
	require.NoError(t, os.WriteFile(path, []byte("# comment\nMORFOLOGIK_RULE_EN,0.9\nUPPERCASE,0.5,extra\n"), 0o644))
	pattern := filepath.Join(dir, "conf-{lang}.csv")
	m, err := NewConfidenceMapLoader().Load(pattern, []string{"en", "xx"})
	require.NoError(t, err)
	require.Len(t, m, 2)

	_, err = NewConfidenceMapLoader().Load("/no/{lang}/file.csv", []string{"en"})
	require.Error(t, err)
	_, err = NewConfidenceMapLoader().Load("/no/placeholder.csv", []string{"en"})
	require.Error(t, err)
}

func TestExceptions(t *testing.T) {
	require.Contains(t, NewTooManyRequestsError("x").Error(), "x")
	require.Contains(t, NewTextTooLongError("long").Error(), "long")
	require.Contains(t, NewBadRequestError("b").Error(), "b")
	require.Contains(t, NewAuthError("a").Error(), "a")
	require.Contains(t, NewPortBindingError("p").Error(), "p")
	require.Contains(t, NewPathNotFoundError("n").Error(), "n")
	require.Contains(t, NewIllegalConfigurationError("c").Error(), "c")
	inner := NewBadRequestError("inner")
	u := NewUnavailableError("busy", inner)
	require.Contains(t, u.Error(), "busy")
	require.ErrorIs(t, u, inner)
	require.Contains(t, u.Error(), "inner")
}
