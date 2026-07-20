package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func sampleSentenceTokenLemma() *languagetool.AnalyzedSentence {
	// Include SENT_START-like padding so MinTokenCount (often ≥2) is satisfied.
	pos := "pos"
	lem := "lemma"
	ss := languagetool.SentenceStartTagName
	start := languagetool.NewAnalyzedToken("", &ss, nil)
	tok := languagetool.NewAnalyzedToken("token", &pos, &lem)
	return testSentence(languagetool.NewAnalyzedTokenReadingsAt(start, 0),
		languagetool.NewAnalyzedTokenReadingsAt(tok, 0),)
}

func TestRuleSet_TextHintsAreHonored(t *testing.T) {
	// rule requiring "token" surface — present in sample
	suitable := NewAbstractTokenBasedRule("S", "d", "en", []*PatternToken{Token("token")})
	// rule requiring "unsuitable" — absent
	unsuitable := NewAbstractTokenBasedRule("U", "d", "en", []*PatternToken{Token("unsuitable")})
	// unclassified POS-like rule without surface token hints: empty token list
	unrelated := NewAbstractTokenBasedRule("P", "d", "en", nil)

	set := TextHintedRuleSet([]RuleIDGetter{suitable, unsuitable, unrelated})
	got := set.RulesForSentence(sampleSentenceTokenLemma())
	ids := map[string]bool{}
	for _, r := range got {
		ids[r.GetID()] = true
	}
	require.True(t, ids["S"])
	require.False(t, ids["U"])
	require.True(t, ids["P"]) // no text hints → always included
}

func TestRuleSet_LemmaHintsAreHonored(t *testing.T) {
	// inflected form matching lemma "lemma"
	pt := Token("lemma")
	pt.MatchInflected = true
	suitable := NewAbstractTokenBasedRule("L", "d", "en", []*PatternToken{pt})
	// unsuitable lemma
	bad := Token("unsuitable")
	bad.MatchInflected = true
	unsuitable := NewAbstractTokenBasedRule("U", "d", "en", []*PatternToken{bad})
	// pos-only (no token surface) stays
	unrelated := NewAbstractTokenBasedRule("P", "d", "en", nil)

	set := TextLemmaHintedRuleSet([]RuleIDGetter{suitable, unsuitable, unrelated})
	got := set.RulesForSentence(sampleSentenceTokenLemma())
	ids := map[string]bool{}
	for _, r := range got {
		ids[r.GetID()] = true
	}
	require.True(t, ids["L"])
	require.False(t, ids["U"])
	require.True(t, ids["P"])
}
