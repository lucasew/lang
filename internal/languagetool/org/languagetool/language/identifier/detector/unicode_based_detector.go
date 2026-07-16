package detector

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

// GetDominantLangCodes returns language codes whose script dominates the sample.
func (d *UnicodeBasedDetector) GetDominantLangCodes(str string) []string {
	// Match Java: iterate UTF-16 code units via range over string as runes for BMP scripts
	// (all script ranges used here are BMP).
	max := d.maxCheckLength
	arabic, cyrillic, cjk, khmer, tamil := 0, 0, 0, 0, 0
	greek, devanagari, thai, hebrew, hangul := 0, 0, 0, 0, 0
	significant := 0
	n := 0
	for _, val := range str {
		if n >= max {
			break
		}
		// Java counts each charAt once; for BMP this is 1 per character.
		n++
		if !isWhitespaceRune(val) && !isDigitRune(val) && val != '.' {
			significant++
		}
		v := int(val)
		if v >= 0x0600 && v <= 0x06FF {
			arabic++
		}
		if v >= 0x0400 && v <= 0x04FF {
			cyrillic++
		}
		if (v >= 0x4E00 && v <= 0x9FFF) || (v >= 0x3040 && v <= 0x309F) || (v >= 0x30A0 && v <= 0x30FF) {
			cjk++
		}
		if v >= 0x1780 && v <= 0x17FF {
			khmer++
		}
		if v >= 0xB82 && v <= 0xBFA {
			tamil++
		}
		if (v >= 0x0370 && v <= 0x03FF) || (v >= 0x1F00 && v <= 0x1FFF) {
			greek++
		}
		if v >= 0x0900 && v <= 0x097F {
			devanagari++
		}
		if v >= 0x0E00 && v <= 0x0E7F {
			thai++
		}
		if (v >= 0x0590 && v <= 0x05FF) || (v >= 0xFB1D && v <= 0xFB40) {
			hebrew++
		}
		if (v >= 0xAC00 && v <= 0xD7AF) || (v >= 0x1100 && v <= 0x11FF) ||
			(v >= 0x3130 && v <= 0x318F) || (v >= 0xA960 && v <= 0xA97F) ||
			(v >= 0xD7B0 && v <= 0xD7FF) {
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

func isWhitespaceRune(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\u00A0'
}

func isDigitRune(r rune) bool {
	return r >= '0' && r <= '9'
}
