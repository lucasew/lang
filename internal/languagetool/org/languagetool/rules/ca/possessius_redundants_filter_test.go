package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func ptrPR(s string) *string { return &s }

func atrAt(token, pos, lemma string, start int) *languagetool.AnalyzedTokenReadings {
	r := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken(token, ptrPR(pos), ptrPR(lemma)), start)
	return r
}

func sentencePR(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	start := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", ptrPR(languagetool.SentenceStartTagName), nil))
	all := append([]*languagetool.AnalyzedTokenReadings{start}, toks...)
	return languagetool.NewAnalyzedSentence(all)
}

func TestPersonaNumberFromPX(t *testing.T) {
	p, n := PersonaNumberFromPX("PX1MS0S0")
	require.Equal(t, "1", p)
	require.Equal(t, "S", n)
}

func TestPossessiusRedundantsFilter_SuggestPronounFound(t *testing.T) {
	f := NewPossessiusRedundantsFilter()
	got := f.Suggest(PossessiveSuggestionInput{
		PronounFound:     true,
		ApostropheNeeded: true,
		NounToken:        "amic",
	})
	require.Equal(t, "l'amic", got)

	got = f.Suggest(PossessiveSuggestionInput{PronounFound: true, ApostropheNeeded: false})
	require.Equal(t, "", got)
}

func TestPossessiusRedundantsFilter_SuggestAddDative(t *testing.T) {
	f := NewPossessiusRedundantsFilter()
	got := f.Suggest(PossessiveSuggestionInput{
		Persona: "3", Number: "S",
		HasSomePronoun:   false,
		VerbToken:        "trenca",
		AroundPossessive: []string{"el", "braç"},
	})
	require.Contains(t, got, "trenca")
	require.Contains(t, got, "li")
}

func TestPossessiusRedundantsRegistered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.ca.PossessiusRedundantsFilter"))
}

// "trenca el seu braç" — no matching pronoun → suggest dative "li trenca el braç"
// tokens non-blank: [0]SENT [1]trenca [2]el [3]seu [4]braç
// patternTokenPos points at match start (verb index 1)
func TestPossessiusRedundantsFilter_AcceptAddDative(t *testing.T) {
	f := NewPossessiusRedundantsFilter()
	// positions: start offsets for SetOffsetPosition checks
	sent := sentencePR(
		atrAt("trenca", "VMIP3S00", "trencar", 0),
		atrAt("el", "DA0MS0", "el", 7),
		atrAt("seu", "PX3MS0S0", "seu", 10),
		atrAt("braç", "NCMS000", "braç", 14),
	)
	// mark whitespace before middle tokens
	for i, tok := range sent.GetTokensWithoutWhitespace() {
		if i > 1 {
			tok.SetWhitespaceBefore(true)
		}
	}
	m := rules.NewRuleMatch(nil, sent, 0, 18, "msg")
	// patternTokenPos = verb index in non-blank stream
	out := f.AcceptRuleMatch(m, nil, 1, nil, nil)
	require.NotNil(t, out)
	sugs := out.GetSuggestedReplacements()
	require.Len(t, sugs, 1)
	require.Contains(t, sugs[0], "li")
	require.Contains(t, sugs[0], "trenca")
	require.Contains(t, sugs[0], "braç")
	require.NotContains(t, sugs[0], "seu")
}

// Matching dative "li" before verb + redundant possessive → empty delete of possessive (no apostrophe)
func TestPossessiusRedundantsFilter_AcceptPronounFoundDelete(t *testing.T) {
	f := NewPossessiusRedundantsFilter()
	// li(PP) trenca el seu braç
	// PP3CSD00: person 3 at [2], number C at [4] — persona 3, number S on PX: number C matches any
	// Java: pronounPostag.substring(2,3).equals(persona) && (number.equals("C") || pronounPostag.substring(4,5).equals(number))
	// For pronoun: person at 2, number at 4. For PX: persona at 2, number at 6.
	// li: typically PP3CSD00 — person 3, number C
	sent := sentencePR(
		atrAt("li", "PP3CSD00", "li", 0),
		atrAt("trenca", "VMIP3S00", "trencar", 3),
		atrAt("el", "DA0MS0", "el", 10),
		atrAt("seu", "PX3MS0S0", "seu", 13),
		atrAt("braç", "NCMS000", "braç", 17),
	)
	// No GV chunks → posVerb = patternTokenPos (1) after walk back from patternTokenPos-1=0
	// Wait: patternTokenPos should be start of match. Java uses patternTokenPos then finds PX forward.
	// posVerb = patternTokenPos - 1, walk back while GV, then posVerb++.
	// If patternTokenPos is 2 (el/start of det-possessive-noun), posVerb starts at 1 (trenca).
	// Looking at Java: posPossessive starts at patternTokenPos and advances to PX.
	// So patternTokenPos can be at the verb or earlier.
	// posVerb = patternTokenPos - 1; while GV walk; posVerb++.
	// If patternTokenPos=2 (trenca is 1), posVerb starts 1... actually if patternTokenPos is verb index:
	// posVerb = verbIndex - 1 = li index 1? Let's set patternTokenPos to verb index 2 in this sentence
	// tokens: 0 SENT, 1 li, 2 trenca, 3 el, 4 seu, 5 braç
	// patternTokenPos = 2 (trenca): posVerb=1, has no GV, posVerb becomes 2 (trenca). Good.
	// pronoun before: posPronoun=1 (li) — PP3CSD00 matches person 3, number C with PX3... number S? number on PX is S; condition number.equals("C") || pronoun[4]==number → number S, pronoun number at 4 is C? PP3CSD00: indices 0P1P2=3 3C 4S?
	// PP3CSD00: [0]=P [1]=P [2]=3 [3]=C [4]=S [5]=D...
	// Wait PP3CSD00 length: P P 3 C S D 0 0
	// substring(2,3)=3, substring(4,5)=S
	// PX3MS0S0: persona=3, number=S (index 6)
	// match: person 3==3 && (number S==C? no || pronoun S==S) → true
	m := rules.NewRuleMatch(nil, sent, 13, 16, "msg")
	out := f.AcceptRuleMatch(m, nil, 2, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{""}, out.GetSuggestedReplacements())
	// offset on possessive only
	require.Equal(t, 13, out.GetFromPos())
	require.Equal(t, 16, out.GetToPos()) // start 13, "seu" len 3 → end 16
}

func TestPossessiusRedundantsFilter_NoMatch(t *testing.T) {
	f := NewPossessiusRedundantsFilter()
	require.Nil(t, f.AcceptRuleMatch(nil, nil, 0, nil, nil))
}
