package morfologik

import (
	"unicode"
)

// IsMisspelled ports morfologik.speller.Speller.isMisspelled (2.2.0).
// Empty word → false (not misspelled). Uses Dictionary .info flags + Contains.
func (d *Dictionary) IsMisspelled(word string) bool {
	if d == nil || word == "" {
		return false
	}
	wordToCheck := applyConversionPairs(word, d.InputConversion)
	if wordToCheck == "" {
		return false
	}
	// Java: isAlphabetic = word.length() != 1 || isAlphabetic(charAt(0))
	r := []rune(wordToCheck)
	isAlpha := len(r) != 1 || isAlphabeticRune(r[0])
	// (!ignorePunctuation || isAlphabetic)
	if d.IgnorePunctuation && !isAlpha {
		return false
	}
	// (!ignoreNumbers || containsNoDigit)
	if d.IgnoreNumbers && containsDigitRunes(wordToCheck) {
		return false
	}
	if d.IgnoreCamelCase && isCamelCase(wordToCheck) {
		return false
	}
	if d.IgnoreAllUppercase && isAlpha && isAllUppercase(wordToCheck) {
		return false
	}
	if d.Contains(wordToCheck) {
		return false
	}
	// convertCase arm: accept lower / initial-upper of non-mixed-case words
	if d.ConvertCase && !isMixedCase(wordToCheck) {
		low := d.ToLower(wordToCheck)
		if d.Contains(low) {
			return false
		}
		if isAllUppercase(wordToCheck) {
			iu := d.initialUppercase(wordToCheck)
			if iu != wordToCheck && d.Contains(iu) {
				return false
			}
		}
	}
	return true
}

// isAlphabeticRune ports Speller.isAlphabetic Unicode letter categories.
func isAlphabeticRune(codePoint rune) bool {
	return unicode.Is(unicode.Lu, codePoint) || unicode.Is(unicode.Ll, codePoint) ||
		unicode.Is(unicode.Lt, codePoint) || unicode.Is(unicode.Lm, codePoint) ||
		unicode.Is(unicode.Lo, codePoint) || unicode.Is(unicode.Nl, codePoint)
}

func containsDigitRunes(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// isAllUppercase ports Speller.isAllUppercase: true unless a letter is lowercase.
// Empty / digits-only → true (Java returns true when no lowercase letter found).
func isAllUppercase(str string) bool {
	for _, c := range str {
		if unicode.IsLetter(c) && unicode.IsLower(c) {
			return false
		}
	}
	return true
}

// isNotAllLowercase ports Speller.isNotAllLowercase.
func isNotAllLowercase(str string) bool {
	for _, c := range str {
		if unicode.IsLetter(c) && !unicode.IsLower(c) {
			return true
		}
	}
	return false
}

// isNotCapitalizedWord ports Speller.isNotCapitalizedWord.
func isNotCapitalizedWord(str string) bool {
	if str != "" {
		r := []rune(str)
		if unicode.IsUpper(r[0]) {
			for i := 1; i < len(r); i++ {
				c := r[i]
				if unicode.IsLetter(c) && !unicode.IsLower(c) {
					return true
				}
			}
			return false
		}
	}
	return true
}

// isMixedCase ports Speller.isMixedCase.
func isMixedCase(str string) bool {
	return !isAllUppercase(str) && isNotCapitalizedWord(str) && isNotAllLowercase(str)
}

// isCamelCase ports Speller.isCamelCase (German dash compounds included by this definition).
func isCamelCase(str string) bool {
	if str == "" {
		return false
	}
	r := []rune(str)
	if isAllUppercase(str) || !isNotCapitalizedWord(str) {
		return false
	}
	if !unicode.IsUpper(r[0]) {
		return false
	}
	if len(r) > 1 && !unicode.IsLower(r[1]) {
		return false
	}
	return isNotAllLowercase(str)
}

// initialUppercase ports Speller.initialUppercase with dictionary locale.
func (d *Dictionary) initialUppercase(wordToCheck string) string {
	if wordToCheck == "" {
		return wordToCheck
	}
	// Java: substring(0,1) + substring(1).toLowerCase(locale)
	// Use rune-aware first character for non-BMP safety; EN dicts are BMP.
	r := []rune(wordToCheck)
	first := string(r[0])
	rest := ""
	if len(r) > 1 {
		rest = d.ToLower(string(r[1:]))
	}
	// first char: Character.toUpperCase on char — use locale upper of first then take first rune
	up := d.ToUpper(first)
	ur := []rune(up)
	if len(ur) == 0 {
		return wordToCheck
	}
	return string(ur[0]) + rest
}

// initialUppercase is a package helper for tests without Dictionary (locale Und).
func initialUppercase(wordToCheck string) string {
	d := &Dictionary{}
	return d.initialUppercase(wordToCheck)
}
