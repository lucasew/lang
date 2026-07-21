package detector

import (
	"strings"
	"unicode"
	"unicode/utf16"
)

// NGramDetector ports org.languagetool.language.identifier.detector.NGramDetector
// as a surface over CharNGramDetector plus Unicode script heuristics for CJK/etc.
// Full zip-model Kneser-Ney loading is deferred.
type NGramDetector struct {
	*CharNGramDetector
	MaxLength int
}

func NewNGramDetector(maxLength int) *NGramDetector {
	if maxLength <= 0 {
		maxLength = 1000
	}
	return &NGramDetector{
		CharNGramDetector: NewCharNGramDetector(3),
		MaxLength:         maxLength,
	}
}

// DetectLanguages scores text with char n-grams, then boosts by script heuristics.
func (d *NGramDetector) DetectLanguages(text string) map[string]float64 {
	if d == nil {
		return nil
	}
	// Java: if (text.length() > maxLength) text = text.substring(0, maxLength);
	if d.MaxLength > 0 && javaStringLenDet(text) > d.MaxLength {
		text = javaSubstringDet(text, 0, d.MaxLength)
	}
	scores := d.CharNGramDetector.DetectLanguages(text)
	if scores == nil {
		scores = map[string]float64{}
	}
	// digits-only → noop language signal
	if isDigitsOnlyText(text) {
		return map[string]float64{"zz": 1}
	}
	// script boosts
	var hasKo, hasJa, hasZh, hasKm, hasTl, hasHy, hasEl, hasTa bool
	for _, r := range text {
		switch {
		case r >= 0xac00 && r <= 0xd7a3:
			hasKo = true
		case r >= 0x3040 && r <= 0x30ff:
			hasJa = true
		case unicode.Is(unicode.Han, r):
			hasZh = true
		case r >= 0x1780 && r <= 0x17ff:
			hasKm = true
		case r >= 0x1700 && r <= 0x171f:
			hasTl = true
		case r >= 0x0530 && r <= 0x058f:
			hasHy = true
		case r >= 0x0370 && r <= 0x03ff:
			hasEl = true
		case r >= 0x0b80 && r <= 0x0bff:
			hasTa = true
		}
	}
	if hasKo {
		scores["ko"] += 2
	}
	if hasJa {
		scores["ja"] += 2
	}
	if hasZh {
		scores["zh"] += 2
	}
	if hasKm {
		scores["km"] += 2
	}
	if hasTl {
		scores["tl"] += 1.5
	}
	if hasHy {
		scores["hy"] += 1.5
	}
	if hasEl {
		scores["el"] += 1.5
	}
	if hasTa {
		scores["ta"] += 1.5
	}
	return scores
}

// TopLanguage returns the best code or "" if empty.
func (d *NGramDetector) TopLanguage(text string) string {
	scores := d.DetectLanguages(text)
	var best string
	var bestV float64 = -1
	for k, v := range scores {
		if v > bestV {
			bestV = v
			best = k
		}
	}
	return best
}

func isDigitsOnlyText(text string) bool {
	s := strings.TrimSpace(text)
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func javaStringLenDet(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func javaSubstringDet(s string, from, to int) string {
	u := utf16.Encode([]rune(s))
	if from < 0 {
		from = 0
	}
	if to > len(u) {
		to = len(u)
	}
	if from >= to {
		return ""
	}
	return string(utf16.Decode(u[from:to]))
}
