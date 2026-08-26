package de

// Unit coverage for InsertCommaFilter branches beyond InsertCommaFilterTest twin.
// Uses real GermanTagger.INSTANCE (DiscoveredGermanTagger) — no invent POS maps.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func requireGermanTagger(t *testing.T) {
	t.Helper()
	WireGermanTaggerDefaults()
	require.NotNil(t, DiscoveredGermanTagger(), "GermanTagger.INSTANCE resources required")
}

func TestInsertCommaFilter_TwoToken(t *testing.T) {
	// Java parts.length == 2 — no POS required
	f := NewInsertCommaFilter()
	require.Equal(t, []string{"hoffe, es"}, f.Suggest("hoffe es"))
	require.Equal(t, []string{"steht, parkt"}, f.Suggest("steht parkt"))
}

func TestInsertCommaFilter_ThreeTokenVerProPer(t *testing.T) {
	// Java: VER: + PRO:PER: → "hoffe, es geht"
	requireGermanTagger(t)
	f := NewInsertCommaFilter()
	require.Equal(t, []string{"hoffe, es geht"}, f.Suggest("hoffe es geht"))
}

func TestInsertCommaFilter_ThreeTokenSagtMal(t *testing.T) {
	// Java: SAGT + mal + VER: → "Sag mal, hast"
	requireGermanTagger(t)
	f := NewInsertCommaFilter()
	require.Equal(t, []string{"Sag mal, hast"}, f.Suggest("Sag mal hast"))
}

func TestInsertCommaFilter_ThreeTokenVerAdvVer(t *testing.T) {
	// Java: VER: + ADV: + VER: → "denke, hier kann"
	requireGermanTagger(t)
	f := NewInsertCommaFilter()
	require.Equal(t, []string{"denke, hier kann"}, f.Suggest("denke hier kann"))
}

func TestInsertCommaFilter_FourTokenDenkeDemonstrative(t *testing.T) {
	// Java: DENKE_ETC + PRO:DEM: + SUB: → "schätze, diese Krawatte passt"
	requireGermanTagger(t)
	f := NewInsertCommaFilter()
	require.Equal(t, []string{"schätze, diese Krawatte passt"}, f.Suggest("schätze diese Krawatte passt"))
}

func TestInsertCommaFilter_FourTokenVerProPerAdvInr(t *testing.T) {
	// Java: VER + PRO:PER + ADV:INR → "Weißt du, warum diese"
	requireGermanTagger(t)
	f := NewInsertCommaFilter()
	require.Equal(t, []string{"Weißt du, warum diese"}, f.Suggest("Weißt du warum diese"))
}

func TestInsertCommaFilter_FourTokenBeiDir(t *testing.T) {
	// Java test case path: patternTokenPos==1 + BEI + DIR_ETC + VER
	requireGermanTagger(t)
	f := NewInsertCommaFilter()
	require.Equal(t, []string{"Hoffe, bei euch ist"}, f.Suggest("Hoffe bei euch ist"))
}

func TestInsertCommaFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter("org.languagetool.rules.de.InsertCommaFilter"))
	f := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.de.InsertCommaFilter")
	m := rules.NewRuleMatch(rules.NewFakeRule("I"), nil, 0, 8, "msg")
	m.SetSuggestedReplacements([]string{"sag mal"})
	out := f.AcceptRuleMatch(m, nil, 1, nil, nil)
	require.NotNil(t, out)
	// two-token path: no POS required
	require.Equal(t, []string{"sag, mal"}, out.GetSuggestedReplacements())
}

// TestInsertCommaFilter_PreservesType ports Java acceptRuleMatch:
// ruleMatch.setType(match.getType()) — RuleMatch.Type, not ITS IssueType.
func TestInsertCommaFilter_PreservesType(t *testing.T) {
	f := NewInsertCommaFilter()
	for _, typ := range []rules.RuleMatchType{
		rules.RuleMatchTypeHint,
		rules.RuleMatchTypeUnknownWord,
		rules.RuleMatchTypeOther,
	} {
		m := rules.NewRuleMatch(rules.NewFakeRule("I"), nil, 0, 8, "msg")
		m.SetType(typ)
		m.IssueType = "should-not-be-copied" // Java does not set IssueType here
		m.SetSuggestedReplacements([]string{"steht parkt"})
		out := f.AcceptRuleMatch(m, nil, 1, nil, nil)
		require.NotNil(t, out)
		require.Equal(t, typ, out.GetType(), "setType(match.getType())")
		require.Equal(t, []string{"steht, parkt"}, out.GetSuggestedReplacements())
		// NewRuleMatch default has empty IssueType; filter must not copy it from match.
		require.Equal(t, "", out.IssueType)
	}
}
