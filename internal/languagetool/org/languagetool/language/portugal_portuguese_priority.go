package language

// Portugal Portuguese id2prio from org.languagetool.language.PortugalPortuguese.
// Java getPriorityForId: map then super (Portuguese). Do not invent extra IDs.
var portugalPortugueseID2Prio = map[string]int{
	"PT_COMPOUNDS_POST_REFORM":         1,
	"PORTUGUESE_OLD_SPELLING_INTERNAL": -9,
}

// PortugalPortuguesePriorityMap ports PortugalPortuguese.getPriorityMap (variant map only).
func PortugalPortuguesePriorityMap() map[string]int {
	out := make(map[string]int, len(portugalPortugueseID2Prio))
	for k, v := range portugalPortugueseID2Prio {
		out[k] = v
	}
	return out
}

// PortugalPortuguesePriorityForId ports PortugalPortuguese.getPriorityForId → Portuguese.
func PortugalPortuguesePriorityForId(id string) int {
	if p, ok := portugalPortugueseID2Prio[id]; ok {
		return p
	}
	return PortuguesePriorityForId(id)
}

// PortuguesePriorityForIdForCode selects PortugalPortuguese or Portuguese by language code
// (pt-PT / pt_PT → Portugal; otherwise base Portuguese — e.g. pt-BR uses base only).
func PortuguesePriorityForIdForCode(langCode string) func(string) int {
	if isPortugalPortugueseCode(langCode) {
		return PortugalPortuguesePriorityForId
	}
	return PortuguesePriorityForId
}

func isPortugalPortugueseCode(langCode string) bool {
	// Java Portuguese.getDefaultLanguageVariant() → pt-PT (PortugalPortuguese).
	// pt-BR / other regional codes keep base Portuguese priorities only.
	switch langCode {
	case "pt", "pt-PT", "pt_PT", "pt-pt", "pt_pt":
		return true
	}
	if len(langCode) >= 5 {
		// pt-PT… / pt_PT… but not pt-BR
		if (langCode[2] == '-' || langCode[2] == '_') &&
			(langCode[3] == 'P' || langCode[3] == 'p') &&
			(langCode[4] == 'T' || langCode[4] == 't') {
			return true
		}
	}
	return false
}
