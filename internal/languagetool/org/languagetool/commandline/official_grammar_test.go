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

func TestConfigureCoreLT_LoadsOfficialGrammar(t *testing.T) {
	lt, err := configureCoreLT("en", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	// With official grammar present, many rules registered
	ids := lt.GetAllRegisteredRuleIDs()
	require.Greater(t, len(ids), 50, "core + grammar should register many rules")
}
