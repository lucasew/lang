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

// isBrazilianPortugueseCode reports pt-BR / pt_BR locale codes (BrazilianPortuguese).
// Used for getRelevantRules extras gating; priority remains base Portuguese only.
func isBrazilianPortugueseCode(langCode string) bool {
	switch langCode {
	case "pt-BR", "pt_BR", "pt-br", "pt_br":
		return true
	}
	if len(langCode) >= 5 {
		if (langCode[2] == '-' || langCode[2] == '_') &&
			(langCode[3] == 'B' || langCode[3] == 'b') &&
			(langCode[4] == 'R' || langCode[4] == 'r') {
			return true
		}
	}
	return false
}

// isAngolaPortugueseCode reports pt-AO / pt_AO locale codes (AngolaPortuguese).
// Used for getRelevantRules extras gating; priority remains base Portuguese only
// (Java AngolaPortuguese has no id2prio override).
func isAngolaPortugueseCode(langCode string) bool {
	switch langCode {
	case "pt-AO", "pt_AO", "pt-ao", "pt_ao":
		return true
	}
	if len(langCode) >= 5 {
		if (langCode[2] == '-' || langCode[2] == '_') &&
			(langCode[3] == 'A' || langCode[3] == 'a') &&
			(langCode[4] == 'O' || langCode[4] == 'o') {
			return true
		}
	}
	return false
}
