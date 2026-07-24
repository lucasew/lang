package detector

import (
	"unicode"
	"unicode/utf16"
)

// UnicodeBasedDetector ports org.languagetool.language.identifier.detector.UnicodeBasedDetector.
type UnicodeBasedDetector struct {
	maxCheckLength int
}

const (
	defaultMaxCheckLength = 50
	unicodeThreshold      = 0.5
)

func NewUnicodeBasedDetector() *UnicodeBasedDetector {
	return NewUnicodeBasedDetectorMax(defaultMaxCheckLength)
}

func NewUnicodeBasedDetectorMax(maxCheckLength int) *UnicodeBasedDetector {
	return &UnicodeBasedDetector{maxCheckLength: maxCheckLength}
}

// GetDominantLangCodes ports UnicodeBasedDetector.getDominantLangCodes.
// Java iterates charAt(i) for i < min(str.length(), maxCheckLength) — UTF-16 units.
func (d *UnicodeBasedDetector) GetDominantLangCodes(str string) []string {
	if d == nil {
		return nil
	}
	max := d.maxCheckLength
	if max <= 0 {
		max = defaultMaxCheckLength
	}
	u := utf16.Encode([]rune(str))
	limit := len(u)
	if limit > max {
		limit = max
	}
	arabic, cyrillic, cjk, khmer, tamil := 0, 0, 0, 0, 0
	greek, devanagari, thai, hebrew, hangul := 0, 0, 0, 0, 0
	significant := 0
	for i := 0; i < limit; i++ {
		val := int(u[i])
		// Character.isWhitespace / isDigit on the UTF-16 code unit
		if !characterIsWhitespace(val) && !characterIsDigit(val) && val != '.' {
			significant++
		}
		if val >= 0x0600 && val <= 0x06FF {
			arabic++
		}
		if val >= 0x0400 && val <= 0x04FF {
			cyrillic++
		}
		if (val >= 0x4E00 && val <= 0x9FFF) || (val >= 0x3040 && val <= 0x309F) || (val >= 0x30A0 && val <= 0x30FF) {
			cjk++
		}
		if val >= 0x1780 && val <= 0x17FF {
			khmer++
		}
		if val >= 0xB82 && val <= 0xBFA {
			tamil++
		}
		if (val >= 0x0370 && val <= 0x03FF) || (val >= 0x1F00 && val <= 0x1FFF) {
			greek++
		}
		if val >= 0x0900 && val <= 0x097F {
			devanagari++
		}
		if val >= 0x0E00 && val <= 0x0E7F {
			thai++
		}
		if (val >= 0x0590 && val <= 0x05FF) || (val >= 0xFB1D && val <= 0xFB40) {
			hebrew++
		}
		if (val >= 0xAC00 && val <= 0xD7AF) || (val >= 0x1100 && val <= 0x11FF) ||
			(val >= 0x3130 && val <= 0x318F) || (val >= 0xA960 && val <= 0xA97F) ||
			(val >= 0xD7B0 && val <= 0xD7FF) {
			hangul++
		}
	}
	if significant == 0 {
		return nil
	}
	var langCodes []string
	sig := float32(significant)
	if float32(arabic)/sig >= unicodeThreshold {
		langCodes = append(langCodes, "ar", "fa")
	}
	if float32(cyrillic)/sig >= unicodeThreshold {
		langCodes = append(langCodes, "ru", "uk", "be")
	}
	if float32(cjk)/sig >= unicodeThreshold {
		langCodes = append(langCodes, "zh", "ja")
	}
	if float32(khmer)/sig >= unicodeThreshold {
		langCodes = append(langCodes, "km")
	}
	if float32(tamil)/sig >= unicodeThreshold {
		langCodes = append(langCodes, "ta")
	}
	if float32(greek)/sig >= unicodeThreshold {
		langCodes = append(langCodes, "el")
	}
	if float32(devanagari)/sig >= unicodeThreshold {
		langCodes = append(langCodes, "hi", "mr")
	}
	if float32(thai)/sig >= unicodeThreshold {
		langCodes = append(langCodes, "th")
	}
	if float32(hebrew)/sig >= unicodeThreshold {
		langCodes = append(langCodes, "he")
	}
	if float32(hangul)/sig >= unicodeThreshold {
		langCodes = append(langCodes, "ko")
	}
	return langCodes
}

// characterIsWhitespace ports java.lang.Character.isWhitespace(int) for BMP code units.
// Used by UnicodeBasedDetector (not isSpaceChar).
func characterIsWhitespace(val int) bool {
	if val < 0 || val > 0xFFFF {
		return false
	}
	r := rune(val)
	switch r {
	case '\t', '\n', '\u000B', '\f', '\r':
		return true
	case 0x1C, 0x1D, 0x1E, 0x1F:
		return true
	}
	// Java excludes non-breaking Zs from isWhitespace
	if r == '\u00A0' || r == '\u2007' || r == '\u202F' {
		return false
	}
	return unicode.Is(unicode.Zs, r) || unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r)
}

func characterIsDigit(val int) bool {
	return val >= '0' && val <= '9'
}
