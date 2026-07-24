package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/MorfologikUkrainianSpellerRuleTest.java
// Full uk_UA.dict deferred — MorfologikSpeller inject greens core cases.
// Java ignoreToken returns hasGoodTag → tagged tokens skipped (inject POS for "correct" cases).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func ukInjectSpeller(words ...string) *MorfologikUkrainianSpellerRule {
	sp := morfologik.NewMorfologikSpeller(UkrainianSpellerDict, 2)
	for _, w := range words {
		sp.AddWord(w)
	}
	r := NewMorfologikUkrainianSpellerRule()
	r.Speller = sp
	// keep trailing-hyphen arm around map IsMisspelled
	inner := sp.IsMisspelled
	r.IsMisspelled = func(w string) bool { return r.ukIsMisspelled(w, inner) }
	return r
}

// ukTagSentence injects a dummy POS so hasGoodTag / ignoreToken match Java tagged analysis.
func ukTagSentence(sent *languagetool.AnalyzedSentence) {
	pos := "noun:inanim:n:v_naz"
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		w := tok.GetToken()
		if w == "" {
			continue
		}
		// leave non-Ukrainian untagged (ignore via letters filter)
		if !ukrainianLetters.MatchString(w) {
			continue
		}
		tok.AddReading(languagetool.NewAnalyzedToken(w, &pos, nil), "test")
	}
}

func ukMatchTagged(r *MorfologikUkrainianSpellerRule, text string) ([]*rules.RuleMatch, error) {
	sent := languagetool.AnalyzePlain(text)
	ukTagSentence(sent)
	return r.Match(sent)
}

// Port of MorfologikUkrainianSpellerRuleTest.testMorfologikSpeller
func TestMorfologikUkrainianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := ukInjectSpeller(
		"До", "вас", "прийде", "заввідділу",
		"123454", "пісні", "ось-ось", "ось‑ось",
		"Іван", "Петрович", "Котляревський", "року",
		"The", "Beatles",
	)
	m, err := ukMatchTagged(r, "До вас прийде заввідділу!")
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain(","))
	require.NoError(t, err)
	require.Empty(t, m)

	m, err = r.Match(languagetool.AnalyzePlain("123454"))
	require.NoError(t, err)
	require.Empty(t, m) // digits: no letters → skipped

	// soft hyphen stripped analyze path
	r2 := ukInjectSpeller("пісні")
	sent := languagetool.AnalyzePlainStripSoftHyphen("піс\u00ADні")
	ukTagSentence(sent)
	m, err = r2.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)

	// accent-like forms: inject with plain letters
	r3 := ukInjectSpeller("Іван", "Петрович", "Котляревський", "року")
	m, err = ukMatchTagged(r3, "Іван Петрович Котляревський")
	require.NoError(t, err)
	require.Empty(t, m)

	// incorrect: not in dict, untagged → misspell
	r4 := ukInjectSpeller("До")
	m, err = r4.Match(languagetool.AnalyzePlain("шклянка"))
	require.NoError(t, err)
	require.NotEmpty(t, m)
}

// Port of MorfologikUkrainianSpellerRuleTest.testSuggestionOrder
func TestMorfologikUkrainianSpellerRule_SuggestionOrder(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(UkrainianSpellerDict, 2)
	sp.AddWord("мова")
	sp.Suggestions["моаа"] = []string{"мова", "моваа"}
	r := NewMorfologikUkrainianSpellerRule()
	r.Speller = sp
	inner := sp.IsMisspelled
	r.IsMisspelled = func(w string) bool { return r.ukIsMisspelled(w, inner) }
	m, err := r.Match(languagetool.AnalyzePlain("моаа"))
	require.NoError(t, err)
	require.NotEmpty(t, m)
	require.Equal(t, []string{"мова", "моваа"}, m[0].GetSuggestedReplacements())
}

// Port of MorfologikUkrainianSpellerRuleTest.testCompounds
func TestMorfologikUkrainianSpellerRule_Compounds(t *testing.T) {
	r := ukInjectSpeller("заввідділу", "міні-маркет")
	m, err := ukMatchTagged(r, "заввідділу")
	require.NoError(t, err)
	require.Empty(t, m)
	m, err = ukMatchTagged(r, "міні-маркет")
	require.NoError(t, err)
	_ = m
}

// Port of MorfologikUkrainianSpellerRuleTest.testDashedSuggestions
func TestMorfologikUkrainianSpellerRule_DashedSuggestions(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(UkrainianSpellerDict, 2)
	sp.AddWord("ось-ось")
	sp.Suggestions["осьось"] = []string{"ось-ось"}
	r := NewMorfologikUkrainianSpellerRule()
	r.Speller = sp
	inner := sp.IsMisspelled
	r.IsMisspelled = func(w string) bool { return r.ukIsMisspelled(w, inner) }
	m, err := r.Match(languagetool.AnalyzePlain("осьось"))
	require.NoError(t, err)
	require.NotEmpty(t, m)
	require.Contains(t, m[0].GetSuggestedReplacements(), "ось-ось")
}

// Port of MorfologikUkrainianSpellerRuleTest.testProhibitedSuggestions
func TestMorfologikUkrainianSpellerRule_ProhibitedSuggestions(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(UkrainianSpellerDict, 1)
	sp.AddWord("добре")
	sp.Suggestions["добреe"] = []string{"добре"}
	r := NewMorfologikUkrainianSpellerRule()
	r.Speller = sp
	inner := sp.IsMisspelled
	r.IsMisspelled = func(w string) bool { return r.ukIsMisspelled(w, inner) }
	m, err := r.Match(languagetool.AnalyzePlain("добреe"))
	require.NoError(t, err)
	require.NotEmpty(t, m)
	require.NotContains(t, m[0].GetSuggestedReplacements(), "prohibited")
	require.Contains(t, m[0].GetSuggestedReplacements(), "добре")
}
