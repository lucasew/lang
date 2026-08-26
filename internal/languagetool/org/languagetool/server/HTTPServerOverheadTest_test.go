package server

// Twin of HTTPServerOverheadTest — lightweight config + detect overhead smoke.
import (
	"testing"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of HTTPServerOverheadTest (no @Test)
func TestHTTPServerOverhead_NoTests(t *testing.T) {
	cfg := NewHTTPServerConfig()
	require.NotNil(t, cfg)
	require.Greater(t, cfg.Port, 0)

	lt := languagetool.NewJLanguageTool("en")
	start := time.Now()
	for i := 0; i < 50; i++ {
		require.NotEmpty(t, lt.Analyze("overhead probe"))
		_ = DetectLanguageOfString("overhead probe text here", nil, nil)
	}
	require.Less(t, time.Since(start), 5*time.Second)

	// request limit defaults are sane
	require.GreaterOrEqual(t, cfg.MaxTextLengthAnonymous, 0)
}
