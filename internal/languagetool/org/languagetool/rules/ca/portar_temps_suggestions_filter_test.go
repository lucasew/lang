package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrPT(s string) *string { return &s }

func atrPT(token, pos, lemma string, start int) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, ptrPT(pos), ptrPT(lemma)), start)
}

func sentencePT(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	start := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", ptrPT(languagetool.SentenceStartTagName), nil))
	return languagetool.NewAnalyzedSentence(append([]*languagetool.AnalyzedTokenReadings{start}, toks...))
}

func TestPortarTempsSuggestionsFilter_Suggest(t *testing.T) {
	f := NewPortarTempsSuggestionsFilter()
	f.SynthFer = func(p string) string { return "fa" }
	got := f.Suggest(PortarTempsInput{
		PortarPostag: "VMIP3S00",
		TimeTokens:   []string{"una", "hora"},
		Kind:         PortarTempsQue,
		CasingModel:  "porta",
	})
	require.Equal(t, "fa una hora que", got)

	f.SynthInfinitiveToFinite = func(lemma, tag string) string { return "treballa" }
	got = f.Suggest(PortarTempsInput{
		PortarPostag:  "VMIP3S00",
		TimeTokens:    []string{"una", "hora"},
		Kind:          PortarTempsGerund,
		NextLemma:     "treballar",
		PronounsAfter: "ho",
	})
	require.Contains(t, got, "fa una hora que")
	require.Contains(t, got, "treballa")
}

func TestPortarTempsRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.PortarTempsSuggestionsFilter"))
}

// porta una hora que … → fa una hora que
// tokens: [0]SENT [1]porta [2]una(PTime) [3]hora(PTime) [4]que [5].
func TestPortarTempsSuggestionsFilter_AcceptQue(t *testing.T) {
	f := NewPortarTempsSuggestionsFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postag string, re bool) []string {
		lem := ""
		if tok.GetLemma() != nil {
			lem = *tok.GetLemma()
		}
		if lem == "fer" && re {
			return []string{"fa"}
		}
		return nil
	}
	una := atrPT("una", "DI0FS0", "un", 6)
	hora := atrPT("hora", "NCFS000", "hora", 10)
	una.SetChunkTags([]string{"PTime"})
	hora.SetChunkTags([]string{"PTime"})
	una.SetWhitespaceBefore(true)
	hora.SetWhitespaceBefore(true)
	que := atrPT("que", "CS", "que", 15)
	que.SetWhitespaceBefore(true)
	dot := atrPT(".", "_PUNCT", ".", 18)

	sent := sentencePT(
		atrPT("porta", "VMIP3S00", "portar", 0),
		una, hora, que, dot,
	)
	m := rules.NewRuleMatch(nil, sent, 0, 14, "msg")
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"fa una hora que"}, out.GetSuggestedReplacements())
}

// porta una hora treballant → fa una hora que treballa
func TestPortarTempsSuggestionsFilter_AcceptGerund(t *testing.T) {
	f := NewPortarTempsSuggestionsFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postag string, re bool) []string {
		lem := ""
		if tok.GetLemma() != nil {
			lem = *tok.GetLemma()
		}
		switch {
		case lem == "fer":
			return []string{"fa"}
		case lem == "treballar":
			return []string{"treballa"}
		default:
			return nil
		}
	}
	una := atrPT("una", "DI0FS0", "un", 6)
	hora := atrPT("hora", "NCFS000", "hora", 10)
	una.SetChunkTags([]string{"PTime"})
	hora.SetChunkTags([]string{"PTime"})
	una.SetWhitespaceBefore(true)
	hora.SetWhitespaceBefore(true)
	ger := atrPT("treballant", "VMG0000", "treballar", 15)
	ger.SetWhitespaceBefore(true)
	dot := atrPT(".", "_PUNCT", ".", 26)
	dot.SetWhitespaceBefore(true)

	sent := sentencePT(
		atrPT("porta", "VMIP3S00", "portar", 0),
		una, hora, ger, dot,
	)
	m := rules.NewRuleMatch(nil, sent, 0, 14, "msg")
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"fa una hora que treballa"}, out.GetSuggestedReplacements())
}

func TestPortarTempsSuggestionsFilter_NoSynth(t *testing.T) {
	f := NewPortarTempsSuggestionsFilter()
	sent := sentencePT(atrPT("porta", "VMIP3S00", "portar", 0))
	m := rules.NewRuleMatch(nil, sent, 0, 5, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, nil, 0, nil, nil))
}

func TestPortarTempsSuggestionsFilter_AcceptEstarPred(t *testing.T) {
	f := NewPortarTempsSuggestionsFilter()
	f.Synthesize = func(tok *languagetool.AnalyzedToken, postag string, re bool) []string {
		lem := ""
		if tok.GetLemma() != nil {
			lem = *tok.GetLemma()
		}
		switch {
		case lem == "fer":
			return []string{"fa"}
		case lem == "estar":
			return []string{"està"}
		default:
			return nil
		}
	}
	una := atrPT("una", "DI0FS0", "un", 6)
	hora := atrPT("hora", "NCFS000", "hora", 10)
	una.SetChunkTags([]string{"PTime"})
	hora.SetChunkTags([]string{"PTime"})
	una.SetWhitespaceBefore(true)
	hora.SetWhitespaceBefore(true)
	aqui := atrPT("aquí", "RG", "aquí", 15)
	aqui.SetWhitespaceBefore(true)
	// need lastTokenPos+1 in bounds: Java checks lastTokenPos+1 >= length at start of time end
	// after PTime, last is "aquí", need one more token for the early check lastTokenPos+1 >= len
	// Actually check is: if (lastTokenPos + 1 >= tokens.length) return null;
	// so we need at least one token after lastToken (the continuation token itself is lastTokenPos,
	// so we need lastTokenPos+1 to exist — e.g. period after aquí)
	dot := atrPT(".", "_PUNCT", ".", 20)
	dot.SetWhitespaceBefore(true)

	sent := sentencePT(
		atrPT("porta", "VMIP3S00", "portar", 0),
		una, hora, aqui, dot,
	)
	// Wait: "aquí" is not in the estar list as alone... list has aquí. Yes.
	// But lastToken is aquí which matches equals("aquí"). adjustEndPos-- so end is lastTokenPos-1 = hora
	m := rules.NewRuleMatch(nil, sent, 0, 14, "msg")
	out := f.AcceptRuleMatch(m, nil, 0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"fa una hora que està"}, out.GetSuggestedReplacements())
}
