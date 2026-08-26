package language

// British English id2prio from org.languagetool.language.BritishEnglish.
// Java getPriorityForId: map then super (English). Do not invent extra IDs.
var britishEnglishID2Prio = map[string]int{
	"OXFORD_SPELLING_ISATION_NOUNS": -20,
	"OXFORD_SPELLING_ISE_VERBS":     -21,
	"OXFORD_SPELLING_IZE":           -22,
}

// BritishEnglishPriorityMap ports BritishEnglish.getPriorityMap (defensive copy of variant map only).
func BritishEnglishPriorityMap() map[string]int {
	out := make(map[string]int, len(britishEnglishID2Prio))
	for k, v := range britishEnglishID2Prio {
		out[k] = v
	}
	return out
}

// BritishEnglishPriorityForId ports BritishEnglish.getPriorityForId → English.
func BritishEnglishPriorityForId(id string) int {
	if p, ok := britishEnglishID2Prio[id]; ok {
		return p
	}
	return EnglishPriorityForId(id)
}

// EnglishPriorityForIdForCode selects BritishEnglish or English by language code
// (en-GB / en_GB → British; otherwise English).
func EnglishPriorityForIdForCode(langCode string) func(string) int {
	if isBritishEnglishCode(langCode) {
		return BritishEnglishPriorityForId
	}
	return EnglishPriorityForId
}

func isBritishEnglishCode(langCode string) bool {
	// Match common Go/Java short codes for British English.
	switch langCode {
	case "en-GB", "en_GB", "en-gb", "en_gb":
		return true
	}
	// Suffix forms: en-GB-oxendict etc.
	if len(langCode) >= 5 {
		// en-GB… or en_GB…
		if (langCode[2] == '-' || langCode[2] == '_') &&
			(langCode[3] == 'G' || langCode[3] == 'g') &&
			(langCode[4] == 'B' || langCode[4] == 'b') {
			return true
		}
	}
	return false
}
