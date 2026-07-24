package uk

// Twin of MissingHyphenRuleTest — POS noun + WordTagged gate (Java WordTagger).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestMissingHyphenRule_Rule(t *testing.T) {
	rule := NewMissingHyphenRule(nil)
	// Accept all suggested compounds (simulates dictionary hits).
	rule.WordTagged = func(word string) bool { return true }

	check := func(text, want string) {
		t.Helper()
		matches := rule.Match(analyzeMissingHyphen(text))
		require.Equal(t, 1, len(matches), "text %q", text)
		require.Equal(t, want, matches[0].GetSuggestedReplacements()[0], "text %q", text)
	}

	check("Поїхали у штаб квартиру.", "штаб-квартиру")
	check("Роблю тайм аут", "тайм-аут")
	check("Такий компакт диск.", "компакт-диск")
	// :alt → join without hyphen
	check("Такий міні автомобіль.", "мініавтомобіль")
	check("Арт проект вийшов провальним.", "Артпроект")
	check("Шоу бізнес - не моє.", "Шоубізнес")
	check("екс віце-президент", "ексвіце-президент")

	// exceptions / correct
	require.Equal(t, 0, len(rule.Match(analyzeMissingHyphen("всі медіа півострова."))))
	require.Equal(t, 0, len(rule.Match(analyzeMissingHyphen("на шоу поп-діви"))))
	require.Equal(t, 0, len(rule.Match(analyzeMissingHyphen("Тут все гаразд."))))
	// блок removed from prefixes
	require.Equal(t, 0, len(rule.Match(analyzeMissingHyphen("Такий блок схемі не потрібен."))))
}

func TestMissingHyphenRule_FailClosedWithoutPOS(t *testing.T) {
	rule := NewMissingHyphenRule(nil)
	rule.WordTagged = func(word string) bool { return true }
	// AnalyzePlain: no noun POS on second token
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("Поїхали у штаб квартиру.")))
}

func TestMissingHyphenRule_FailClosedWithoutWordTagged(t *testing.T) {
	rule := NewMissingHyphenRule(nil)
	// POS present but no WordTagged and not :alt-with-next-alt
	require.Empty(t, rule.Match(analyzeMissingHyphen("Поїхали у штаб квартиру.")))
}

func TestMissingHyphenRule_AltWithNextAltPOS(t *testing.T) {
	rule := NewMissingHyphenRule(nil)
	// :alt path without WordTagged when next has :alt
	sent := languagetool.AnalyzeWithTagger("міні автомобіль", func(tok string) []languagetool.TokenTag {
		switch strings.ToLower(tok) {
		case "автомобіль":
			return []languagetool.TokenTag{{POS: "noun:m:v_naz:alt", Lemma: "автомобіль"}}
		default:
			return nil
		}
	})
	matches := rule.Match(sent)
	require.Equal(t, 1, len(matches))
	require.Equal(t, "мініавтомобіль", matches[0].GetSuggestedReplacements()[0])
}

// analyzeMissingHyphen tags real words as noun (Java FreeLing noun on second token).
// Does not invent POS for punctuation-only tokens (e.g. "-").
func analyzeMissingHyphen(text string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTagger(text, func(tok string) []languagetool.TokenTag {
		low := strings.ToLower(tok)
		if !allLowerUK.MatchString(low) {
			return nil
		}
		// require at least one letter (exclude "-" alone which matches dash in ALL_LOWER)
		hasLetter := false
		for _, r := range low {
			if r >= 'а' && r <= 'я' || r == 'і' || r == 'ї' || r == 'є' || r == 'ґ' {
				hasLetter = true
				break
			}
		}
		if !hasLetter {
			return nil
		}
		if low == "аут" {
			return []languagetool.TokenTag{{POS: "noun:m:v_naz", Lemma: "аут"}}
		}
		return []languagetool.TokenTag{{POS: "noun:m:v_naz", Lemma: low}}
	})
}
