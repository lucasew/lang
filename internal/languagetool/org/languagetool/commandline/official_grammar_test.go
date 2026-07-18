package commandline

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestDiscoverAndLoadOfficialENGrammar(t *testing.T) {
	p := DiscoverLanguageGrammarXML(nil, "en")
	require.NotEmpty(t, p, "official en grammar.xml")
	require.NotContains(t, p, "soft.xml")
	lt := languagetool.NewJLanguageTool("en")
	n, err := patterns.RegisterGrammarFile(lt, p, "en")
	require.NoError(t, err)
	t.Logf("registered %d rules from %s", n, p)
	require.Greater(t, n, 100, "should load many surface-simple rules from official grammar")
	// COULD_OF family may be in grammar under various ids — at least a/an style or modal of
	ms := lt.Check("I could of done better.")
	// Not required to fire if rule needs POS/unify not yet supported — log only
	t.Logf("could-of matches=%d", len(ms))
	for _, m := range ms {
		t.Logf("  %s", m.RuleID)
	}
}

func TestDiscoverAndLoadOfficialENStyle(t *testing.T) {
	p := DiscoverLanguageStyleXML(nil, "en")
	require.NotEmpty(t, p, "official en style.xml")
	require.NotContains(t, p, "soft.xml")
	require.Contains(t, p, "style.xml")
	lt := languagetool.NewJLanguageTool("en")
	n, err := patterns.RegisterGrammarFile(lt, p, "en")
	require.NoError(t, err)
	t.Logf("registered %d rules from %s", n, p)
	require.Greater(t, n, 10, "style.xml should register surface-simple pattern rules")
}

func TestConfigureCoreLT_LoadsOfficialGrammarWhenEnabled(t *testing.T) {
	t.Setenv("LANG_USE_UPSTREAM_GRAMMAR", "1")
	lt, err := configureCoreLT("en", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	ids := lt.GetAllRegisteredRuleIDs()
	// grammar + style both load under the same gate (Java getRuleFileNames).
	require.Greater(t, len(ids), 100, "core + grammar + style when enabled")
	// Filter-class rules still skipped when unsupported (e.g. incomplete filters).
	ms := lt.Check("This is a simple sentence.")
	for _, m := range ms {
		require.NotContains(t, m.RuleID, "MULTITOKEN_SPELLING", "%+v", m)
		// Surface-only REPEATED_VERBS without POS should not fire on plain English prose.
		require.NotEqual(t, "REPEATED_VERBS", m.RuleID, "%+v", m)
	}
}
