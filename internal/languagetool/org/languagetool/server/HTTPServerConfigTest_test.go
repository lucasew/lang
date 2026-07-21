package server

// Twin of HTTPServerConfigTest
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of HTTPServerConfigTest.testArgumentParsing
func TestHTTPServerConfig_ArgumentParsing(t *testing.T) {
	c1, err := NewHTTPServerConfigFromArgs([]string{})
	require.NoError(t, err)
	require.Equal(t, DefaultPort, c1.Port)
	require.False(t, c1.PublicAccess)
	require.False(t, c1.Verbose)

	c2, err := NewHTTPServerConfigFromArgs([]string{"--public"})
	require.NoError(t, err)
	require.Equal(t, DefaultPort, c2.Port)
	require.True(t, c2.PublicAccess)
	require.False(t, c2.Verbose)

	c3, err := NewHTTPServerConfigFromArgs([]string{"--port", "80"})
	require.NoError(t, err)
	require.Equal(t, 80, c3.Port)
	require.False(t, c3.PublicAccess)

	c4, err := NewHTTPServerConfigFromArgs([]string{"--port", "80", "--public"})
	require.NoError(t, err)
	require.Equal(t, 80, c4.Port)
	require.True(t, c4.PublicAccess)
}

// Port of HTTPServerConfigTest.shouldLoadLanguageModelDirectoryFromCommandLineArguments
func TestHTTPServerConfig_ShouldLoadLanguageModelDirectoryFromCommandLineArguments(t *testing.T) {
	dir := t.TempDir()
	lm := filepath.Join(dir, "languageModelDirectory")
	require.NoError(t, os.MkdirAll(lm, 0o755))

	c, err := NewHTTPServerConfigFromArgs([]string{LanguageModelOption, lm})
	require.NoError(t, err)
	require.NotEmpty(t, c.LanguageModelDir)
	require.Equal(t, lm, c.LanguageModelDir)
	st, err := os.Stat(c.LanguageModelDir)
	require.NoError(t, err)
	require.True(t, st.IsDir())
	require.True(t, filepath.Base(c.LanguageModelDir) == "languageModelDirectory" ||
		filepath.Base(c.LanguageModelDir) == filepath.Base(lm))
}

func TestParseJavaProperties(t *testing.T) {
	p := ParseJavaProperties(`
# comment
! also comment
port=8082
maxTextLength = 500
requestLimit: 10
pipelineCaching true
emptyKey=
`)
	require.Equal(t, "8082", p["port"])
	require.Equal(t, "500", p["maxTextLength"])
	require.Equal(t, "10", p["requestLimit"])
	require.Equal(t, "true", p["pipelineCaching"])
	require.Equal(t, "", p["emptyKey"])
}

// Ports HTTPServerConfig.parseConfigFile open-source subset + getOptionalProperty defaults.
func TestHTTPServerConfig_LoadFromPropertyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "server.properties")
	body := `
# LanguageTool server config
maxTextLength=2000
maxTextLengthAnonymous=1000
maxCheckTimeMillis=5000
maxCheckTimeMillisPremium=10000
requestLimit=100
requestLimitInBytes=10000
pipelineCaching=true
pipelinePrewarming=true
maxPipelinePoolSize=7
pipelineExpireTimeInSeconds=30
maxCheckThreads=4
textCheckerQueueSize=16
cacheSize=100
cacheTTLSeconds=60
maxErrorsPerWordRate=0.5
suggestionsEnabled=false
maxSpellingSuggestions=3
blockedReferrers=spam.example,  bad.example
premiumOnly=false
anonymousAccessAllowed=false
prometheusMonitoring=true
prometheusPort=9400
disabledRuleIds=RULE_A, RULE_B
localApiMode=true
motherTongue=de-DE
preferredLanguages=en, de, fr
trustXForwardForHeader=true
maxWorkQueueSize=50
ipFingerprintFactor=2
timeoutRequestLimit=5
`
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	c := NewHTTPServerConfig()
	require.NoError(t, c.LoadFromPropertyFile(path))
	require.Equal(t, 1000, c.MaxTextLengthAnonymous) // tier override
	require.Equal(t, 2000, c.MaxTextLengthLoggedIn)
	require.Equal(t, 2000, c.MaxTextLengthPremium)
	require.Equal(t, int64(5000), c.MaxCheckTimeMillisAnonymous)
	require.Equal(t, int64(10000), c.MaxCheckTimeMillisPremium)
	require.Equal(t, 100, c.RequestLimit)
	require.Equal(t, 10000, c.RequestLimitInBytes)
	require.True(t, c.PipelineCaching)
	require.True(t, c.PipelinePrewarming)
	require.Equal(t, 7, c.MaxPipelinePoolSize)
	require.Equal(t, 30, c.PipelineExpireTime)
	require.Equal(t, 4, c.MaxCheckThreads)
	require.Equal(t, 16, c.TextCheckerQueueSize)
	require.Equal(t, 100, c.CacheSize)
	require.Equal(t, int64(60), c.CacheTTLSeconds)
	require.InDelta(t, 0.5, c.MaxErrorsPerWordRate, 1e-9)
	require.False(t, c.SuggestionsEnabled)
	require.Equal(t, 3, c.MaxSpellingSuggestions)
	require.Equal(t, []string{"spam.example", "bad.example"}, c.BlockedReferrers)
	require.False(t, c.AnonymousAccessAllowed)
	require.True(t, c.PrometheusMonitoring)
	require.Equal(t, 9400, c.PrometheusPort)
	require.Equal(t, []string{"RULE_A", "RULE_B"}, c.DisabledRuleIDs)
	require.True(t, c.LocalAPIMode)
	require.Equal(t, "de-DE", c.MotherTongue)
	require.Equal(t, []string{"en", "de", "fr"}, c.PreferredLanguages)
	require.True(t, c.TrustXForwardedForHeader)
	require.Equal(t, 50, c.MaxWorkQueueSize)
	require.Equal(t, 2, c.IPFingerprintFactor)
	require.Equal(t, 5, c.TimeoutRequestLimit)

	// missing file
	require.Error(t, c.LoadFromPropertyFile(filepath.Join(dir, "nope.properties")))
}

func TestHTTPServerConfig_ApplyProperties_AfterTheDeadlineRejected(t *testing.T) {
	c := NewHTTPServerConfig()
	require.Panics(t, func() {
		c.ApplyProperties(map[string]string{"mode": "AfterTheDeadline"})
	})
}
