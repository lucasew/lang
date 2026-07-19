package uk

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// ALT_DASHES_IN_WORD: letter/digit — letter or letter — digit (en dash U+2013).
var altDashesInWordRE = regexp.MustCompile(
	`(?i)[а-яіїєґ0-9a-z]\x{2013}[а-яіїєґ]|[а-яіїєґ]\x{2013}[0-9]`,
)

// WordsWithBrackets approximates UkrainianWordTokenizer.WORDS_WITH_BRACKETS_PATTERN.
// Java tags adjusted word after stripping [] when this matches.
var wordsWithBracketsRE = regexp.MustCompile(`\[[^\]]+\]`)

// leftOAdjInvalidSolidRE: ^(prefix)(rest) for solid LEFT_O_ADJ_INVALID compounds (len>=9).
var leftOAdjInvalidSolidRE *regexp.Regexp

func init() {
	// Build from leftOAdjInvalid map keys (order not important for |).
	keys := make([]string, 0, len(leftOAdjInvalid))
	for k := range leftOAdjInvalid {
		keys = append(keys, regexp.QuoteMeta(k))
	}
	// Sort longer first so "зовнішньо" wins over shorter if any overlap
	// (simple length sort)
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if len(keys[j]) > len(keys[i]) {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	leftOAdjInvalidSolidRE = regexp.MustCompile(`(?i)^(` + strings.Join(keys, "|") + `)(.+)$`)
}

// AltDashReadings ports getAnalyzedTokens en-dash → hyphen re-tag.
// Returns hyphen-tagged tokens with surface kept as original (hyphen form POS/lemma).
// Java also appends a null reading for the original surface; caller may add it.
func AltDashReadings(word string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || !strings.Contains(word, "\u2013") {
		return nil
	}
	if !altDashesInWordRE.MatchString(word) {
		return nil
	}
	hyphenated := strings.ReplaceAll(word, "\u2013", "-")
	// Tag hyphenated form (dict surface + lower)
	var words []tagging.TaggedWord
	words = append(words, tagWord(hyphenated)...)
	if low := strings.ToLower(hyphenated); low != hyphenated {
		words = append(words, tagWord(low)...)
	}
	if len(words) == 0 {
		return nil
	}
	var out []*languagetool.AnalyzedToken
	seen := map[string]struct{}{}
	for _, tw := range words {
		if tw.PosTag == "" {
			continue
		}
		key := tw.Lemma + "|" + tw.PosTag
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		p, l := tw.PosTag, tw.Lemma
		out = append(out, languagetool.NewAnalyzedToken(hyphenated, &p, &l))
	}
	if len(out) == 0 {
		return nil
	}
	// Java: newTokens.add(new AnalyzedToken(origWord, null, null));
	out = append(out, languagetool.NewAnalyzedToken(word, nil, nil))
	return out
}

// BracketAltReadings ports getAnalyzedTokens strip [] → :alt.
func BracketAltReadings(word string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || !strings.Contains(word, "[") || !strings.Contains(word, "]") {
		return nil
	}
	if !wordsWithBracketsRE.MatchString(word) {
		return nil
	}
	adjusted := strings.ReplaceAll(strings.ReplaceAll(word, "[", ""), "]", "")
	if adjusted == word || adjusted == "" {
		return nil
	}
	return taggedWithExtra(word, tagWord(adjusted), ":alt", nil)
}

// SolidLeftOAdjInvalidReadings ports getAnalyzedTokens LEFT_O_ADJ_INVALID_PATTERN solid form.
// length >= 9; retag rest as adj; lemma = prefix + adj lemma.
func SolidLeftOAdjInvalidReadings(word string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if tagWord == nil || utf8.RuneCountInString(word) < 9 {
		return nil
	}
	// only when no hyphen (solid) — hyphen path is oAdjMatch
	if strings.Contains(word, "-") {
		return nil
	}
	m := leftOAdjInvalidSolidRE.FindStringSubmatch(strings.ToLower(word))
	if m == nil {
		return nil
	}
	// Preserve original surface case for token; use groups from lower match lengths
	// Re-match on original for case of rest if needed — Java uses matcher on word with CASE_INSENSITIVE.
	mOrig := leftOAdjInvalidSolidRE.FindStringSubmatch(word)
	if mOrig == nil {
		return nil
	}
	prefix, rest := mOrig[1], mOrig[2]
	wdList := tagWord(rest)
	if low := strings.ToLower(rest); low != rest && len(wdList) == 0 {
		wdList = tagWord(low)
	}
	var out []*languagetool.AnalyzedToken
	for _, tw := range wdList {
		if !strings.HasPrefix(tw.PosTag, "adj") {
			continue
		}
		lemma := prefix + tw.Lemma
		p, l := tw.PosTag, lemma
		out = append(out, languagetool.NewAnalyzedToken(word, &p, &l))
	}
	return out
}
