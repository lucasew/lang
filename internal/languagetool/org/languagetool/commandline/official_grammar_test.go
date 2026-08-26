package commandline

import (
	"strings"
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

func TestDiscoverAndLoadEnglishL2Grammar(t *testing.T) {
	de := DiscoverEnglishL2GrammarXML(nil, "de")
	require.NotEmpty(t, de, "grammar-l2-de.xml")
	require.Contains(t, de, "grammar-l2-de.xml")
	require.NotContains(t, de, "soft")
	fr := DiscoverEnglishL2GrammarXML(nil, "fr")
	require.NotEmpty(t, fr, "grammar-l2-fr.xml")
	require.Contains(t, fr, "grammar-l2-fr.xml")
	require.Empty(t, DiscoverEnglishL2GrammarXML(nil, "es"), "no invent for other mother tongues")

	lt := languagetool.NewJLanguageTool("en")
	n, err := patterns.RegisterGrammarFile(lt, de, "en")
	require.NoError(t, err)
	t.Logf("L2-de registered %d rules from %s", n, de)
	require.Greater(t, n, 0)
	// Java idprefix="L2_" → L2_THAN_AS
	ids := lt.GetAllRegisteredRuleIDs()
	var hasL2 bool
	for _, id := range ids {
		if strings.HasPrefix(id, "L2_") {
			hasL2 = true
			break
		}
	}
	require.True(t, hasL2, "ids=%v", ids)
}

func TestConfigureCoreLT_LoadsL2WhenMotherTongueDE(t *testing.T) {
	lt, err := configureCoreLT("en", &CommandLineOptions{Language: "en", MotherTongue: "de"})
	require.NoError(t, err)
	ids := lt.GetAllRegisteredRuleIDs()
	var hasL2 bool
	for _, id := range ids {
		if strings.HasPrefix(id, "L2_") {
			hasL2 = true
			t.Logf("found L2 id %s", id)
			break
		}
	}
	require.True(t, hasL2, "motherTongue=de should register L2_* ids; ids sample=%v", ids[:min(20, len(ids))])
	ltCore, err := configureCoreLT("en", &CommandLineOptions{Language: "en"})
	require.NoError(t, err)
	require.Greater(t, len(ids), len(ltCore.GetAllRegisteredRuleIDs()),
		"motherTongue=de should load more rules than core alone")
}

func TestDiscoverLanguagePatternRuleFiles_EN(t *testing.T) {
	files := DiscoverLanguagePatternRuleFiles(nil, "en")
	require.GreaterOrEqual(t, len(files), 2, "grammar + style")
	require.Contains(t, files[0], "grammar.xml")
	require.Contains(t, files[1], "style.xml")
	for _, f := range files {
		require.NotContains(t, f, "soft")
	}
}

func TestDiscoverLanguagePatternRuleFiles_ENUS_Variant(t *testing.T) {
	files := DiscoverLanguagePatternRuleFiles(nil, "en-US")
	require.GreaterOrEqual(t, len(files), 2, "base grammar + style at least")
	// Java adds en/en-US/grammar.xml when present.
	var hasVariantGrammar bool
	for _, f := range files {
		require.NotContains(t, f, "soft")
		if strings.Contains(f, "en-US") && strings.Contains(f, "grammar.xml") {
			hasVariantGrammar = true
		}
	}
	// inspiration or testdata should provide en-US/grammar.xml
	require.True(t, hasVariantGrammar, "files=%v", files)
}

func TestDiscoverLanguagePatternRuleFiles_UK_ExtraRuleFiles(t *testing.T) {
	// Java Ukrainian.RULE_FILES after super.getRuleFileNames() (grammar.xml first).
	files := DiscoverLanguagePatternRuleFiles(nil, "uk")
	require.NotEmpty(t, files)
	require.Contains(t, files[0], "grammar.xml")
	require.NotContains(t, files[0], "grammar-grammar.xml")
	want := []string{
		"grammar-spelling.xml",
		"grammar-grammar.xml",
		"grammar-barbarism.xml",
		"grammar-style.xml",
		"grammar-punctuation.xml",
	}
	for _, name := range want {
		found := false
		for _, f := range files {
			require.NotContains(t, f, "soft")
			if strings.HasSuffix(f, name) || strings.Contains(f, "/"+name) {
				found = true
				break
			}
		}
		require.True(t, found, "missing %s in %v", name, files)
	}
	// order: extras after grammar.xml (and optional style/custom)
	idx := map[string]int{}
	for i, f := range files {
		for _, name := range want {
			if strings.Contains(f, name) {
				if _, ok := idx[name]; !ok {
					idx[name] = i
				}
			}
		}
	}
	require.Less(t, idx["grammar-spelling.xml"], idx["grammar-grammar.xml"])
	require.Less(t, idx["grammar-grammar.xml"], idx["grammar-barbarism.xml"])
	require.Less(t, idx["grammar-barbarism.xml"], idx["grammar-style.xml"])
	require.Less(t, idx["grammar-style.xml"], idx["grammar-punctuation.xml"])
}

func TestDiscoverLanguagePatternRuleFiles_SK_Typography(t *testing.T) {
	files := DiscoverLanguagePatternRuleFiles(nil, "sk")
	require.NotEmpty(t, files)
	require.Contains(t, files[0], "grammar.xml")
	var hasTypo bool
	for _, f := range files {
		if strings.Contains(f, "grammar-typography.xml") {
			hasTypo = true
		}
	}
	require.True(t, hasTypo, "files=%v", files)
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
