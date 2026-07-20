package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

type urlRule struct {
	id, url string
}

func (r urlRule) GetID() string          { return r.id }
func (r urlRule) GetDescription() string { return "d" }
func (r urlRule) GetURL() string         { return r.url }

func TestToLocalMatches_MatchURL(t *testing.T) {
	r := urlRule{id: "R", url: "https://rule.example/"}
	sent := languagetool.AnalyzePlain("ab")
	m := NewRuleMatch(r, sent, 0, 2, "msg")
	m.SetURL("https://match.example/lemma")
	out := ToLocalMatches([]*RuleMatch{m})
	require.Len(t, out, 1)
	require.Equal(t, "https://match.example/lemma", out[0].URL)
	require.Equal(t, "R", out[0].RuleID)
}

func TestToLocalMatches_RuleURLFallback(t *testing.T) {
	r := urlRule{id: "DE_CASE", url: "https://dict.leo.org/case"}
	sent := languagetool.AnalyzePlain("ab")
	m := NewRuleMatch(r, sent, 0, 2, "msg")
	out := ToLocalMatches([]*RuleMatch{m})
	require.Len(t, out, 1)
	require.Equal(t, "https://dict.leo.org/case", out[0].URL)
}

// bareIDRule exposes only GetID — RuleMeta fills category/ITS when getters absent.
type bareIDRule struct{ id string }

func (r bareIDRule) GetID() string { return r.id }

func TestToLocalMatches_RuleMetaFallback(t *testing.T) {
	// Java RuleMeta: DE_AGREEMENT → GRAMMAR when rule has no GetCategory.
	sent := languagetool.AnalyzePlain("ab")
	m := NewRuleMatch(bareIDRule{id: "DE_AGREEMENT"}, sent, 0, 2, "msg")
	out := ToLocalMatches([]*RuleMatch{m})
	require.Len(t, out, 1)
	require.Equal(t, "DE_AGREEMENT", out[0].RuleID)
	require.Equal(t, "GRAMMAR", out[0].CategoryID)
	require.Equal(t, "grammar", out[0].IssueType)
	require.NotEmpty(t, out[0].Description)
	require.NotEqual(t, "DE_AGREEMENT", out[0].Description)

	// Unknown id: do not invent category (uncategorized RuleMeta skipped).
	m2 := NewRuleMatch(bareIDRule{id: "TOTALLY_UNKNOWN_XYZ"}, sent, 0, 2, "msg")
	out2 := ToLocalMatches([]*RuleMatch{m2})
	require.Len(t, out2, 1)
	require.Empty(t, out2[0].CategoryID)
	require.Empty(t, out2[0].IssueType)
}

func TestToLocalMatches_OriginalErrorStrFromSentence(t *testing.T) {
	// Java SwissGerman uses sentence.substring(from,to); ToLocalMatches fills OriginalErrorStr.
	r := urlRule{id: "AI_DE_GGEC_REPLACEMENT_ORTHOGRAPHY_SPELL", url: ""}
	sent := languagetool.AnalyzePlain("gross Haus")
	m := NewRuleMatch(r, sent, 0, 5, "msg")
	m.SetSuggestedReplacements([]string{"groß"})
	out := ToLocalMatches([]*RuleMatch{m})
	require.Len(t, out, 1)
	require.Equal(t, "gross", out[0].OriginalErrorStr)

	// Prefer sentence positions when set (Java setOriginalErrorStr).
	m2 := NewRuleMatch(r, sent, 100, 105, "msg") // document pos out of sentence range
	m2.SetSentencePosition(0, 5)
	out2 := ToLocalMatches([]*RuleMatch{m2})
	require.Len(t, out2, 1)
	require.Equal(t, "gross", out2[0].OriginalErrorStr)

	// Do not overwrite explicit surface.
	m3 := NewRuleMatch(r, sent, 0, 5, "msg")
	m3.OriginalErrorStr = "kept"
	out3 := ToLocalMatches([]*RuleMatch{m3})
	require.Equal(t, "kept", out3[0].OriginalErrorStr)
}

// premiumByID is a test Premium registry (Java Premium.get().isPremiumRule).
type premiumByID map[string]bool

func (p premiumByID) IsPremiumRule(ruleID string) bool { return p[ruleID] }

// Java RuleMatchesAsJsonSerializer writes match.getSpecificRuleId() as "id".
func TestToLocalMatches_SpecificRuleIdPreferred(t *testing.T) {
	sent := languagetool.AnalyzePlain("außerdem außerdem")
	m := NewRuleMatch(bareIDRule{id: "DE_REPEATEDWORDS"}, sent, 0, 9, "msg")
	m.SetSpecificRuleId("DE_REPEATEDWORDS_AUSSERDEM")
	out := ToLocalMatches([]*RuleMatch{m})
	require.Len(t, out, 1)
	require.Equal(t, "DE_REPEATEDWORDS_AUSSERDEM", out[0].RuleID)
	// RuleMeta still categorizes specific lemma suffix IDs
	require.Equal(t, "REPETITIONS_STYLE", out[0].CategoryID)
	require.Equal(t, "style", out[0].IssueType)
}

func TestToLocalMatches_DefaultPremiumRegistry(t *testing.T) {
	prev := languagetool.DefaultPremium
	languagetool.DefaultPremium = premiumByID{"SECRET_RULE": true}
	t.Cleanup(func() { languagetool.DefaultPremium = prev })

	sent := languagetool.AnalyzePlain("ab")
	m := NewRuleMatch(bareIDRule{id: "SECRET_RULE"}, sent, 0, 2, "msg")
	out := ToLocalMatches([]*RuleMatch{m})
	require.Len(t, out, 1)
	require.True(t, out[0].IsPremium)

	// Open-source PremiumOff: PREMIUM in id still marks premium for LocalMatch paths.
	languagetool.DefaultPremium = languagetool.PremiumOff{}
	m2 := NewRuleMatch(bareIDRule{id: "FOO_PREMIUM_BAR"}, sent, 0, 2, "msg")
	out2 := ToLocalMatches([]*RuleMatch{m2})
	require.True(t, out2[0].IsPremium)

	m3 := NewRuleMatch(bareIDRule{id: "OPEN_RULE"}, sent, 0, 2, "msg")
	out3 := ToLocalMatches([]*RuleMatch{m3})
	require.False(t, out3[0].IsPremium)
}


// Java CleanOverlappingFilter uses Tag.picky demotion and isIncludedInErrorsCorrectedAllAtOnce.
// ToLocalMatches must copy both flags from the rule onto LocalMatch.
func TestToLocalMatches_PickyAndAllAtOnce(t *testing.T) {
	sent := languagetool.AnalyzePlain("Hallo,")
	r := NewFakeRuleWithTag("P2_RULE", TagPicky)
	r.SetIncludedInErrorsCorrectedAllAtOnce(true)
	m := NewRuleMatch(r, sent, 0, 6, "msg")
	m.SetSuggestedReplacement("Hallo")
	m.SetSentencePosition(0, 6)
	out := ToLocalMatches([]*RuleMatch{m})
	require.Len(t, out, 1)
	require.True(t, out[0].IsPicky)
	require.True(t, out[0].IncludedInErrorsCorrectedAllAtOnce)
	require.Equal(t, 0, out[0].FromPosSentence)
	require.Equal(t, 6, out[0].ToPosSentence)
	require.Equal(t, "Hallo,", out[0].OriginalErrorStr)
}

func TestToLocalMatches_NotPickyByDefault(t *testing.T) {
	sent := languagetool.AnalyzePlain("ab")
	m := NewRuleMatch(NewFakeRule("PLAIN"), sent, 0, 2, "msg")
	out := ToLocalMatches([]*RuleMatch{m})
	require.Len(t, out, 1)
	require.False(t, out[0].IsPicky)
	require.False(t, out[0].IncludedInErrorsCorrectedAllAtOnce)
}

func TestToLocalMatches_ToneTagsAndGoalSpecific(t *testing.T) {
	r := NewFakeRule("TONE_RULE")
	r.SetToneTags(languagetool.ToneFormal, languagetool.ToneClarity)
	r.SetGoalSpecific(true)
	m := NewRuleMatch(r, nil, 0, 2, "msg")
	out := ToLocalMatches([]*RuleMatch{m})
	require.Len(t, out, 1)
	require.Equal(t, []languagetool.ToneTag{languagetool.ToneFormal, languagetool.ToneClarity}, out[0].ToneTags)
	require.True(t, out[0].GoalSpecific)
	// level/tone filter: goal-specific + formal tone inactive under default empty tone set
	meta := languagetool.LocalMatchLevelToneMeta(out[0])
	require.False(t, languagetool.IsRuleActiveForLevelAndToneTags(meta, languagetool.LevelDefault, nil))
	require.True(t, languagetool.IsRuleActiveForLevelAndToneTags(meta, languagetool.LevelDefault, map[languagetool.ToneTag]struct{}{languagetool.ToneFormal: {}}))
}
