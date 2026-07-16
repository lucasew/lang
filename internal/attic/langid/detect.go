package langid

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/attic/data"
)

// Detect picks a language among candidates using a lightweight stopword score.
// This is not LT's full language detector; it is a honest auto default until
// the official detector data is ported. Low confidence falls back to en-US/en.
func Detect(text string, candidates []data.Language) (code string, ok bool) {
	if len(candidates) == 0 {
		return "", false
	}
	lower := strings.ToLower(text)
	words := tokenizeWords(lower)
	if len(words) == 0 {
		return prefer(candidates, "en-US", "en"), true
	}

	scores := map[string]int{}
	for fam, stops := range stopwords {
		for _, w := range words {
			if _, hit := stops[w]; hit {
				scores[fam]++
			}
		}
	}

	bestFam := ""
	bestScore := 0
	for fam, sc := range scores {
		if sc > bestScore {
			bestScore = sc
			bestFam = fam
		}
	}
	if bestScore == 0 || bestFam == "" {
		return prefer(candidates, "en-US", "en"), true
	}
	// Pick best matching candidate for family.
	for _, c := range candidates {
		if c.Family == bestFam || strings.HasPrefix(strings.ToLower(c.Code), bestFam) {
			// Prefer regional default when family matches.
			if bestFam == "en" {
				return prefer(candidates, "en-US", "en"), true
			}
			if bestFam == "pt" {
				return prefer(candidates, "pt-BR", "pt-PT", "pt"), true
			}
			if bestFam == "de" {
				return prefer(candidates, "de-DE", "de"), true
			}
			return c.Code, true
		}
	}
	return prefer(candidates, "en-US", "en"), true
}

func prefer(candidates []data.Language, codes ...string) string {
	for _, want := range codes {
		for _, c := range candidates {
			if strings.EqualFold(c.Code, want) {
				return c.Code
			}
		}
	}
	for _, want := range codes {
		for _, c := range candidates {
			if strings.EqualFold(c.Family, want) {
				return c.Code
			}
		}
	}
	return candidates[0].Code
}

func tokenizeWords(s string) []string {
	var out []string
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			out = append(out, b.String())
			b.Reset()
		}
	}
	for _, r := range s {
		if unicode.IsLetter(r) {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}

var stopwords = map[string]map[string]struct{}{
	"en": set("the", "and", "is", "of", "to", "in", "that", "it", "for", "you", "with", "on", "was", "are", "this"),
	"pt": set("de", "que", "o", "a", "e", "do", "da", "em", "um", "para", "é", "com", "não", "uma", "os"),
	"de": set("der", "die", "und", "das", "ist", "nicht", "ein", "eine", "zu", "den", "mit", "von", "auf", "für"),
	"fr": set("le", "de", "un", "et", "est", "la", "les", "des", "en", "du", "une", "que", "pour", "dans", "qui"),
	"es": set("de", "la", "que", "el", "en", "y", "los", "del", "se", "las", "por", "un", "con", "una", "para"),
	"it": set("di", "e", "il", "la", "che", "è", "per", "un", "in", "è", "del", "una", "sono", "con", "non"),
	"nl": set("de", "het", "een", "van", "en", "in", "is", "op", "te", "dat", "die", "voor", "met", "niet"),
	"pl": set("i", "w", "na", "z", "nie", "to", "się", "do", "że", "jak", "ale", "o", "po", "za"),
}

func set(words ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(words))
	for _, w := range words {
		m[w] = struct{}{}
	}
	return m
}
