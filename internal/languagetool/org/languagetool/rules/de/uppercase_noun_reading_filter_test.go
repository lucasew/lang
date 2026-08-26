package de

// Behavior matrix for UppercaseNounReadingFilter (Java has no dedicated *Test.java).
// Production path: GermanTagger.INSTANCE.tag via DiscoveredGermanTagger — no invent POS maps.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func requireUppercaseNounGermanTagger(t *testing.T) {
	t.Helper()
	WireGermanTaggerDefaults()
	require.NotNil(t, DiscoveredGermanTagger(), "GermanTagger.INSTANCE resources required")
}

// TestUppercaseNounReadingFilter_BehaviorMatrix_RealTagger proves Java-visible outcomes
// with real GermanTagger.INSTANCE (stand→Stand SUB accept; pure ADJ / unknown reject).
func TestUppercaseNounReadingFilter_BehaviorMatrix_RealTagger(t *testing.T) {
	requireUppercaseNounGermanTagger(t)
	f := NewUppercaseNounReadingFilter()
	// Tag field nil → resolveTagger uses process-wide GermanTagger.INSTANCE twin.

	// Java javadoc example: "stand" → "Stand" noun → accept
	require.True(t, f.Accept("stand"), "stand→Stand has SUB: and no ADJ")
	// Pure noun
	require.True(t, f.Accept("Haus"), "Haus is SUB:")
	require.True(t, f.Accept("haus"), "haus→Haus is SUB:")
	// Substantivized infinitive: Laufen has SUB:*:INF readings
	require.True(t, f.Accept("laufen"), "laufen→Laufen has SUB:INF noun reading")
	// bank → Bank noun
	require.True(t, f.Accept("bank"), "bank→Bank is SUB:")

	// ADJ-only (or SUB+ADJ together): Java requires !hasPartialPosTag("ADJ")
	// "Gut" carries both SUB and ADJ → reject
	require.False(t, f.Accept("gut"), "Gut has ADJ → rejected even with SUB")
	// pure adjective uppercased
	require.False(t, f.Accept("schön"), "Schön is ADJ only")
	// Unknown / no noun reading
	require.False(t, f.Accept("xyzzy"), "unknown → no SUB:")
}

func TestUppercaseNounReadingFilter_AcceptRuleMatch_API(t *testing.T) {
	requireUppercaseNounGermanTagger(t)
	f := NewUppercaseNounReadingFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("U_NOUN"), nil, 0, 5, "msg")

	// accept path: stand → Stand noun (Java javadoc example)
	require.Same(t, m, f.AcceptRuleMatch(m, map[string]string{"token": "stand"}, 0, nil, nil))
	// reject path: unknown
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"token": "xyzzy"}, 0, nil, nil))
	// reject path: SUB+ADJ (Gut)
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"token": "gut"}, 0, nil, nil))
	// empty token is present (not null) → Java does not throw; no SUB: → null
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"token": ""}, 0, nil, nil))
	// registered creator instance also uses real tagger
	reg := patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.de.UppercaseNounReadingFilter")
	require.NotNil(t, reg)
	require.Same(t, m, reg.AcceptRuleMatch(m, map[string]string{"token": "stand"}, 0, nil, nil))
	require.Nil(t, reg.AcceptRuleMatch(m, map[string]string{"token": "xyzzy"}, 0, nil, nil))
}

func TestUppercaseNounReadingFilter_MissingTokenPanics(t *testing.T) {
	f := NewUppercaseNounReadingFilter()
	m := rules.NewRuleMatch(rules.NewFakeRule("U_NOUN"), nil, 0, 1, "msg")
	// Java: RuntimeException("Set 'token' for filter " + className + " in rule " + id)
	require.PanicsWithValue(t,
		"Set 'token' for filter org.languagetool.rules.de.UppercaseNounReadingFilter in rule U_NOUN",
		func() {
			f.AcceptRuleMatch(m, map[string]string{}, 0, nil, nil)
		})
	// nil arguments map also missing token
	require.Panics(t, func() {
		f.AcceptRuleMatch(m, nil, 0, nil, nil)
	})
}

func TestUppercaseNounReadingFilter_UsesTagOverride(t *testing.T) {
	// Optional Tag inject mirrors InsertCommaFilter (tests / explicit wire).
	requireUppercaseNounGermanTagger(t)
	f := NewUppercaseNounReadingFilter()
	f.Tag = DiscoveredGermanTagger()
	require.True(t, f.Accept("stand"))
	require.False(t, f.Accept("xyzzy"))
	m := rules.NewRuleMatch(rules.NewFakeRule("U"), nil, 0, 5, "msg")
	require.Same(t, m, f.AcceptRuleMatch(m, map[string]string{"token": "stand"}, 0, nil, nil))
}

func TestUppercaseNounReadingFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.de.UppercaseNounReadingFilter"))
	require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.de.UppercaseNounReadingFilter"))
}
