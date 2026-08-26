package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
)

// POSType is GermanToken.POSType as used by GermanHelper.hasReadingOfType.
// Values are the Java POSType.toString() strings, shared with tagging/de.
type POSType = detag.POSType

// POS type constants (GermanToken.POSType). Aliases keep existing rules/de call sites.
const (
	POSNomen      = detag.POSNomen
	POSVerb       = detag.POSVerb
	POSAdjective  = detag.POSAdjektiv // Java ADJEKTIV
	POSAdjektiv   = detag.POSAdjektiv
	POSDeterminer = detag.POSDeterminer
	POSPronoun    = detag.POSPronomen // Java PRONOMEN
	POSPronomen   = detag.POSPronomen
	POSPartizip   = detag.POSPartizip
	POSProperNoun = detag.POSProperNoun
	POSOther      = detag.POSOther
)

// GetNounCase ports GermanHelper.getNounCase (index 1 of SUB:CASE:NUM:GENDER…).
func GetNounCase(posTag string) string { return getIndexOrEmptyString(posTag, 1) }

// GetNounNumber ports GermanHelper.getNounNumber.
func GetNounNumber(posTag string) string { return getIndexOrEmptyString(posTag, 2) }

// GetNounGender ports GermanHelper.getNounGender.
func GetNounGender(posTag string) string { return getIndexOrEmptyString(posTag, 3) }

// GetDeterminerDefiniteness ports GermanHelper.getDeterminerDefiniteness.
func GetDeterminerDefiniteness(posTag string) string { return getIndexOrEmptyString(posTag, 1) }

// GetDeterminerCase ports GermanHelper.getDeterminerCase.
func GetDeterminerCase(posTag string) string { return getIndexOrEmptyString(posTag, 2) }

// GetDeterminerNumber ports GermanHelper.getDeterminerNumber.
func GetDeterminerNumber(posTag string) string { return getIndexOrEmptyString(posTag, 3) }

// GetDeterminerGender ports GermanHelper.getDeterminerGender.
func GetDeterminerGender(posTag string) string { return getIndexOrEmptyString(posTag, 4) }

// GetComparison ports GermanHelper.getComparison — GRU, KOM, or SUP (index 4, else 2 for PA2:PRD:GRU…).
func GetComparison(posTag string) string {
	// input e.g. ADJ:AKK:PLU:MAS:KOM:SOL
	cmp := getIndexOrEmptyString(posTag, 4)
	if cmp != "GRU" && cmp != "KOM" && cmp != "SUP" {
		// for cases like "PA2:PRD:GRU:VER"
		cmp = getIndexOrEmptyString(posTag, 2)
	}
	return cmp
}

// getIndexOrEmptyString ports GermanHelper.getIndexOrEmptyString.
// Java: null → ""; else posTag.split(":") and return array[idx] if length > idx else "".
// Go has no string null; empty string covers the null case.
func getIndexOrEmptyString(posTag string, idx int) string {
	if posTag == "" {
		return ""
	}
	array := strings.Split(posTag, ":")
	if len(array) > idx {
		return array[idx]
	}
	return ""
}

// HasReadingOfType ports GermanHelper.hasReadingOfType.
// Java: early false on SENTENCE_END / PARAGRAPH_END; else new AnalyzedGermanToken(token).getType() == type.
func HasReadingOfType(tokenReadings *languagetool.AnalyzedTokenReadings, typ POSType) bool {
	if tokenReadings == nil {
		return false
	}
	for _, token := range tokenReadings.GetReadings() {
		if token == nil {
			continue
		}
		posTag := token.GetPOSTag()
		if posTag != nil &&
			(*posTag == languagetool.SentenceEndTagName ||
				*posTag == languagetool.ParagraphEndTagName) {
			return false
		}
		// Java: AnalyzedGermanToken germanToken = new AnalyzedGermanToken(token);
		//       if (germanToken.getType() == type) return true;
		// Single type-resolution path: tagging/de.ParseGermanPOS (AnalyzedGermanToken ctor twin).
		var tag string
		if posTag != nil {
			tag = *posTag
		}
		germanToken := detag.ParseGermanPOS(tag)
		// Java getType() null never equals a non-null POSType; empty Type is the Go twin of null.
		if germanToken.Type == typ {
			return true
		}
	}
	return false
}
