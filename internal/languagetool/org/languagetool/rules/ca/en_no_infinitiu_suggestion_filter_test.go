package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrEN(s string) *string { return &s }

func atrEN(token, pos, lemma string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, ptrEN(pos), ptrEN(lemma)), start)
}

func sentenceEN(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	start := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", ptrEN(languagetool.SentenceStartTagName), nil))
	return languagetool.NewAnalyzedSentence(append([]*languagetool.AnalyzedTokenReadings{start}, toks...))
}

func TestEnNoInfinitiuSuggestionFilter_Suggest(t *testing.T) {
	f := NewEnNoInfinitiuSuggestionFilter()
	f.Synth = func(lemma, postag string) string {
		if postag == "VMIP3S00" {
			return "veu"
		}
		if postag == "VMIP1S00" {
			return "veig"
		}
		return ""
	}
	got := f.Suggest(EnNoInfinitiuInput{
		TempsVerbal: "VMIP1S00",
		Lemma:       "veure",
		VerbBefore:  false,
	})
	require.Contains(t, got, "com que no veu")
	require.Contains(t, got, "com que no veig")
}

func TestEnNoInfinitiuRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.EnNoInfinitiuSuggestionFilter"))
}

// Pattern: "en no veure" with finite verb after "veu" → com que no veu / veig context from after
// tokens non-blank: [0]SENT [1]en [2]no [3]veure [4]veu
// Match span "en no veure" FromPos=en, ToPos=end of veure
// startPos = firstVerb(veure)-2 = en
func TestEnNoInfinitiuSuggestionFilter_AcceptAfter(t *testing.T) {
	f := NewEnNoInfinitiuSuggestionFilter()
	f.Synth = func(lemma, postag string) string {
		switch {
		case lemma == "veure" && postag == "VMIP3S00":
			return "veu"
		case lemma == "veure" && postag == "VMIP1S00":
			return "veig"
		default:
			return ""
		}
	}
	en := atrEN("en", "SPS00", "en", 0)
	no := atrEN("no", "RN", "no", 3)
	no.SetWhitespaceBefore(true)
	veure := atrEN("veure", "VMN00000", "veure", 6)
	veure.SetWhitespaceBefore(true)
	veu := atrEN("veu", "VMIP3S00", "veure", 12)
	veu.SetWhitespaceBefore(true)

	sent := sentenceEN(en, no, veure, veu)
	// match "en no veure" end at veure end pos
	m := rules.NewRuleMatch(nil, sent, 0, veure.GetEndPos(), "msg")
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	// postag from after is 3S → only one synth (no extra 3S branch)
	require.NotEmpty(t, sugs)
	require.Contains(t, sugs[0], "com que no")
	require.Contains(t, sugs[0], "veu")
}

// Finite verb before: digué en no venir → perquè no …
func TestEnNoInfinitiuSuggestionFilter_AcceptBefore(t *testing.T) {
	f := NewEnNoInfinitiuSuggestionFilter()
	f.Synth = func(lemma, postag string) string {
		if lemma == "venir" && postag == "VMII3S00" {
			return "venia"
		}
		// II for past: digué is VMIS3S00 → not IP/IF → VMII + 3S00
		if lemma == "venir" && len(postag) >= 4 && postag[:4] == "VMII" {
			return "venia"
		}
		return ""
	}
	digue := atrEN("digué", "VMIS3S00", "dir", 0)
	en := atrEN("en", "SPS00", "en", 6)
	en.SetWhitespaceBefore(true)
	no := atrEN("no", "RN", "no", 9)
	no.SetWhitespaceBefore(true)
	venir := atrEN("venir", "VMN00000", "venir", 12)
	venir.SetWhitespaceBefore(true)

	sent := sentenceEN(digue, en, no, venir)
	// match starts at "en"
	m := rules.NewRuleMatch(nil, sent, en.GetStartPos(), venir.GetEndPos(), "msg")
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	require.Contains(t, sugs[0], "perquè no")
	require.Contains(t, sugs[0], "venia")
}

func TestEnNoInfinitiuSuggestionFilter_NoSynth(t *testing.T) {
	f := NewEnNoInfinitiuSuggestionFilter()
	require.Nil(t, f.AcceptRuleMatch(nil, nil, 0, nil, nil))
}

func TestEnNoInfinitiuSuggestionFilter_NoNeighborVerb(t *testing.T) {
	f := NewEnNoInfinitiuSuggestionFilter()
	f.Synth = func(lemma, postag string) string { return "x" }
	// only en no veure — no finite neighbor
	en := atrEN("en", "SPS00", "en", 0)
	no := atrEN("no", "RN", "no", 3)
	no.SetWhitespaceBefore(true)
	veure := atrEN("veure", "VMN00000", "veure", 6)
	veure.SetWhitespaceBefore(true)
	sent := sentenceEN(en, no, veure)
	m := rules.NewRuleMatch(nil, sent, 0, veure.GetEndPos(), "msg")
	require.Nil(t, f.AcceptRuleMatch(m, nil, 0, nil, nil))
}
