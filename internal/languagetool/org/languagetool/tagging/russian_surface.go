package tagging

import "strings"

// Combining acute / grave as used in RussianTagger.java string literals.
const (
	ruAcute = "\u0301"
	ruGrave = "\u0300"
)

// NormalizeRussianSurface ports RussianTagger.tag preprocessing before getAnalyzedTokens:
// strip combining acute/grave on vowels, map ʼ → ъ.
// mayMissingYo is the first-stage Java flag (before dictionary confirmation).
func NormalizeRussianSurface(word string) (normalized string, mayMissingYo bool) {
	if len([]rune(word)) <= 1 {
		return word, false
	}
	// Java mayMissingYo conditions (before replace):
	if !(strings.Contains(word, "ё") || strings.Contains(word, "Ё")) &&
		(strings.Contains(word, "е") || strings.Contains(word, "Е")) &&
		!strings.Contains(word, "е"+ruAcute) &&
		!strings.Contains(word, "о"+ruAcute) &&
		!strings.Contains(word, "а"+ruAcute) &&
		!strings.Contains(word, "у"+ruAcute) &&
		!strings.Contains(word, "и"+ruAcute) &&
		!strings.Contains(word, "ю"+ruAcute) &&
		!strings.Contains(word, "ы"+ruAcute) &&
		!strings.Contains(word, "э"+ruAcute) &&
		!strings.Contains(word, "я"+ruAcute) {
		mayMissingYo = true
	}
	w := word
	// Acute (vowel + U+0301)
	for _, v := range []string{"о", "а", "е", "у", "и", "ы", "э", "ю", "я"} {
		w = strings.ReplaceAll(w, v+ruAcute, v)
	}
	// Grave (vowel + U+0300)
	for _, v := range []string{"о", "а", "е", "у", "ы", "э", "ю", "я"} {
		w = strings.ReplaceAll(w, v+ruGrave, v)
	}
	// Precomposed ѝ → и
	w = strings.ReplaceAll(w, "ѝ", "и")
	// Modifier letter apostrophe → hard sign
	w = strings.ReplaceAll(w, "\u02BC", "ъ")
	return w, mayMissingYo
}

// RussianMayMissingYoConfirmed ports the second half of RussianTagger MayMissingYO:
// keep the flag only if lowercased surface with е→ё is known to the word tagger.
func RussianMayMissingYoConfirmed(normalized string, mayMissingYo bool, wt WordTagger) bool {
	if !mayMissingYo || wt == nil {
		return false
	}
	wordLc := strings.ToLower(normalized)
	wordLc = strings.ReplaceAll(wordLc, "е", "ё")
	return len(wt.Tag(wordLc)) > 0
}
