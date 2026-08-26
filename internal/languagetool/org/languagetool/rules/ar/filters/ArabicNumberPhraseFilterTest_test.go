package filters

// Twin of org.languagetool.rules.ar.filters.ArabicNumberPhraseFilterTest (Java king).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	ar_tag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// TestArabicNumberPhraseFilter_Filter twins testFilter.
// Java assertSuggestion (debug=false) only checks suggestion list size == expected.split("|").length
// (empty expected → size 1 via Java String.split).
func TestArabicNumberPhraseFilter_Filter(t *testing.T) {
	inflection := ""
	previous := "في"
	next := ""
	assertSuggestion(t, "صفر", "", previous, next, inflection)
	assertSuggestion(t, "واحد", "", previous, next, inflection)
	assertSuggestion(t, "اثنان", "اثنين", previous, next, inflection)
	assertSuggestion(t, "ثلاثة", "", previous, next, inflection)
	assertSuggestion(t, "إحدى عشر", "أحد عشر", previous, next, inflection)
	assertSuggestion(t, "اثنتي عشر", "اثني عشر", previous, next, inflection)
	assertSuggestion(t, "أربعة عشر", "", previous, next, inflection)
	assertSuggestion(t, "أربعة وثلاثون", "أربعة وثلاثين", previous, next, inflection)
	assertSuggestion(t, "مائةٍ", "", previous, next, inflection)
	assertSuggestion(t, "مائة وخمسة وعشرون", "مائة وخمسة وعشرين", previous, next, inflection)
	assertSuggestion(t, "مائة وأربعة وثلاثين", "", previous, next, inflection)
	assertSuggestion(t, "ألف وتسعمائة واثنان وعشرين", "ألف وتسعمائة واثنين وعشرين", previous, next, inflection)
	assertSuggestion(t, "مليون ومئتان وخمسة وأربعين ألفاً وسبعمائة وواحد", "مليون ومئتين وخمسة وأربعين ألفا وسبعمائة وواحد", previous, next, inflection)
	assertSuggestion(t, "مائة واثنين", "", previous, next, inflection)
	assertSuggestion(t, "عشرة وآلاف", "عشرة آلاف", previous, next, inflection)
}

// TestArabicNumberPhraseFilter_UnitFilter twins testUnitFilter.
func TestArabicNumberPhraseFilter_UnitFilter(t *testing.T) {
	unit := "دينار"
	inflection := ""
	previous := "في"
	assertSuggestion(t, "واحد", "في دينار واحد", previous, unit, inflection)
	assertSuggestion(t, "ثلاثة", "في ثلاثة دنانير", previous, unit, inflection)
	assertSuggestion(t, "إحدى عشر", "في أحد عشر دينارا", previous, unit, inflection)
	assertSuggestion(t, "اثنتي عشر", "في اثني عشر دينارا", previous, unit, inflection)
	assertSuggestion(t, "أربعة عشر", "في أربعة عشر دينارا", previous, unit, inflection)
	assertSuggestion(t, "أربعة وثلاثون", "في أربعة وثلاثين دينارا", previous, unit, inflection)
	assertSuggestion(t, "مائةٍ", "في مائة دينار", previous, unit, inflection)
	assertSuggestion(t, "مائة وخمسة وعشرون", "في مائة وخمسة وعشرين دينارا", previous, unit, inflection)
	assertSuggestion(t, "مائة وأربعة وثلاثين", "في مائة وأربعة وثلاثين دينارا", previous, unit, inflection)
	assertSuggestion(t, "ألف وتسعمائة واثنان وعشرين", "في ألف وتسعمائة واثنين وعشرين دينارا", previous, unit, inflection)
	assertSuggestion(t, "مليون ومئتان وخمسة وأربعين ألفاً وسبعمائة وواحد", "في مليون ومئتين وخمسة وأربعين ألفا وسبعمائة وواحد دينار", previous, unit, inflection)
	assertSuggestion(t, "عشرة وآلاف", "في عشرة آلاف دينار", previous, unit, inflection)
}

func assertSuggestion(t *testing.T, phrase, expectedSuggestion, previousWord, nextWord, inflection string) {
	t.Helper()
	ar_tag.EnsureDefaultArabicTagger()
	require.NotNil(t, ar_tag.DefaultArabicTagger)
	require.NotEmpty(t, ar_tag.ArabicPOSDictPath(), "arabic.dict required for Java twin")

	filter := NewArabicNumberPhraseFilter()
	// Unit path needs synthesizer (Java always constructs ArabicSynthesizer).
	if nextWord != "" {
		require.NotNil(t, filter.synthesizer, "ArabicSynthesizer.INSTANCE required for unit filter twin")
	}

	match := rules.NewRuleMatch(nil, nil, 0, 10, "message")
	args := map[string]string{
		"previous":    previousWord,
		"previousPos": "1",
		"next":        nextWord,
		"inflect":     inflection,
		// the value -1 says that the last word is the given next word
		"nextPos": "-1",
	}

	// tokenize phrase — Java: previous + " " + phrase + " " + next; remove spaces
	fullPhrase := previousWord + " " + phrase + " " + nextWord
	tok := tokenizers.NewArabicWordTokenizer()
	tokens := tok.Tokenize(fullPhrase)
	filtered := make([]string, 0, len(tokens))
	for _, tk := range tokens {
		if tk != " " {
			filtered = append(filtered, tk)
		}
	}

	patternTokens := ar_tag.DefaultArabicTagger.Tag(filtered)
	ruleMatch := filter.AcceptRuleMatch(match, args, -1, patternTokens, nil)

	// Java: expectedSize = expectedSuggestion.split("\\|").length  ("" → 1)
	expectedSize := javaSplitPipeLen(expectedSuggestion)
	require.NotNil(t, ruleMatch, "phrase=%q full=%q", phrase, fullPhrase)
	sugs := ruleMatch.GetSuggestedReplacements()
	require.Equal(t, expectedSize, len(sugs),
		"phrase=%q full=%q sugs=%v expectedContent=%q", phrase, fullPhrase, sugs, expectedSuggestion)

	// Stronger fidelity: when expected non-empty (and not size-only placeholder ""),
	// prefer first suggestion contains the expected correction fragment when available.
	if expectedSuggestion != "" && len(sugs) > 0 {
		// Java only checks size; still assert first suggestion is non-empty for unit path.
		require.NotEmpty(t, sugs[0], "phrase=%q", phrase)
	}
}

// javaSplitPipeLen mirrors Java String.split("\\|").length including "" → 1.
func javaSplitPipeLen(s string) int {
	// Java Pattern.compile("\\|").split(s, 0) discards trailing empties but
	// for the empty string returns {""} length 1.
	if s == "" {
		return 1
	}
	parts := strings.Split(s, "|")
	// Java split discards trailing empty strings
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) == 0 {
		return 0
	}
	return len(parts)
}
