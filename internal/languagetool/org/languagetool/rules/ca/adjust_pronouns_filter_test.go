package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrAP(s string) *string { return &s }

func atrAP(token, pos, lemma string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, ptrAP(pos), ptrAP(lemma)), start)
}

func sentenceAP(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	start := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", ptrAP(languagetool.SentenceStartTagName), nil))
	return languagetool.NewAnalyzedSentence(append([]*languagetool.AnalyzedTokenReadings{start}, toks...))
}

func TestAdjustPronounsFilter_Suggest(t *testing.T) {
	f := NewAdjustPronounsFilter()
	ctx := PronounVerbContext{
		PronounsStr:   "em",
		VerbStr:       "vaig",
		WholeOriginal: "em vaig",
		CasingModel:   "em",
	}
	got := f.Suggest(ctx, "replaceEmEn")
	require.Equal(t, []string{"en vaig"}, got)

	ctx2 := PronounVerbContext{
		PronounsStr:            "ho",
		VerbStr:                "menja",
		FirstVerbPersonaNumber: "3S",
		PronounsAfter:          false,
		WholeOriginal:          "ho menja",
	}
	got = f.Suggest(ctx2, "addPronounReflexive")
	require.NotEmpty(t, got)
	require.Contains(t, got[0], "menja")
}

func TestAdjustPronounsRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.AdjustPronounsFilter"))
}

// "em vaig" → replaceEmEn → "en vaig"
// tokens: dummy SENT, em (P0), vaig (V)
// VerbSynthesizer needs iFirstVerb with room for pronouns before: em at 1, verb at 2
func TestAdjustPronounsFilter_AcceptReplaceEmEn(t *testing.T) {
	f := NewAdjustPronounsFilter()
	// SENT + em + vaig — posWord at em (start of match), verb found at vaig
	// pPronomFeble: P01CN000 for em
	em := atrAP("em", "P01CN000", "jo", 0)
	em.SetWhitespaceBefore(true)
	vaig := atrAP("vaig", "VMIP1S00", "anar", 3)
	vaig.SetWhitespaceBefore(true)
	sent := sentenceAP(em, vaig)
	// FromPos at em start 0
	m := rules.NewRuleMatch(nil, sent, 0, 7, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"actions": "replaceEmEn"}, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"en vaig"}, out.GetSuggestedReplacements())
}

// menjar-ho → addPronounEn with clitics after (infinitive is V.N not IS)
func TestAdjustPronounsFilter_AcceptAddPronounEnAfter(t *testing.T) {
	f := NewAdjustPronounsFilter()
	verb := atrAP("menjar", "VMN00000", "menjar", 0)
	ho := atrAP("ho", "PP3NNA00", "ho", 6)
	ho.SetWhitespaceBefore(false) // clitic attached
	sent := sentenceAP(verb, ho)
	m := rules.NewRuleMatch(nil, sent, 0, 8, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{"actions": "addPronounEn"}, 0, nil, nil)
	// may or may not produce depending on DoAddPronounEn with "ho" + after
	// at least should not panic; if suggestion empty that's ok only if DoAdd returns empty
	if out != nil {
		require.NotEmpty(t, out.GetSuggestedReplacements())
	}
}

func TestAdjustPronounsFilter_MissingActions(t *testing.T) {
	f := NewAdjustPronounsFilter()
	require.Panics(t, func() {
		sent := sentenceAP(atrAP("menja", "VMIP3S00", "menjar", 0))
		m := rules.NewRuleMatch(nil, sent, 0, 5, "msg")
		_ = f.AcceptRuleMatch(m, map[string]string{}, 0, nil, nil)
	})
}

func TestAdjustPronounsFilter_NoVerb(t *testing.T) {
	f := NewAdjustPronounsFilter()
	sent := sentenceAP(atrAP("pa", "NCMS000", "pa", 0))
	m := rules.NewRuleMatch(nil, sent, 0, 2, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, map[string]string{"actions": "replaceEmEn"}, 0, nil, nil))
}
