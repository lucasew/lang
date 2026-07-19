package language

import "strings"

// Dutch rule priorities from org.languagetool.language.Dutch (id2prio + getPriorityForId).
// Java is king — do not invent extra IDs.

// dutchSimpleReplaceRuleID ports SimpleReplaceRule.DUTCH_SIMPLE_REPLACE_RULE.
const dutchSimpleReplaceRuleID = "NL_SIMPLE_REPLACE"

var dutchID2Prio = map[string]int{
	"BE": -3,
	"DE_ONVERWACHT": -20,
	"DOUBLE_PUNCTUATION": -3,
	"EINDE_ZIN_ONVERWACHT": -5,
	"ERG_LANG_WOORD": -20,
	"ET_AL": 1,
	"HET_FIETS": -2,
	"HOOFDLETTERS_OVERBODIG_A": 1,
	"JIJ_JOU_JOUW": -2,
	"JOU_JOUW": -3,
	"KOMMA_KOMMA": -1,
	"KOMMA_ONTBR": -1,
	"N_PERSOONS": 1,
	"SINT_X": 3,
	"STAM_ZONDER_IK": -1,
	"TOO_LONG_PARAGRAPH": -15,
	"VERSCHILLENDE_AANHALINGSTEKENS": 1,
}

// DutchPriorityMap ports Dutch.getPriorityMap (defensive copy).
func DutchPriorityMap() map[string]int {
	out := make(map[string]int, len(dutchID2Prio))
	for k, v := range dutchID2Prio {
		out[k] = v
	}
	return out
}

// DutchPriorityForId ports Dutch.getPriorityForId (then Language base).
// Java checks NL_SIMPLE_REPLACE / NL_SPACE_IN_COMPOUND prefixes BEFORE the map.
func DutchPriorityForId(id string) int {
	if strings.HasPrefix(id, dutchSimpleReplaceRuleID) || strings.HasPrefix(id, "NL_SPACE_IN_COMPOUND") {
		return 1
	}
	if p, ok := dutchID2Prio[id]; ok {
		return p
	}
	if strings.HasPrefix(id, "AI_NL_HYDRA_LEO") {
		if strings.HasPrefix(id, "AI_NL_HYDRA_LEO_MISSING_COMMA") {
			return -51
		}
		return -5
	}
	return languagePriorityForId(id)
}
