package language

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

func init() {
	languagetool.FrenchPriorityForIdHook = FrenchPriorityForId
}

// French rule priorities from org.languagetool.language.French (id2prio + getPriorityForId).
// Java is king — do not invent extra IDs.

var frenchID2Prio = map[string]int{
	"ACCORD_CHAQUE": 100,
	"ACCORD_PLURIEL_ORDINAUX": 10,
	"ACCORD_R_PERS_VERBE": 10,
	"ADJ_ADJ_SENT_END": 10,
	"AGREEMENT_EXCEPTIONS": 100,
	"AGREEMENT_POSTPONED_ADJ": -50,
	"AGREEMENT_TOUT_LE": 10,
	"AN_EN": 10,
	"APOS_M": 10,
	"A_ACCENT": 10,
	"A_ACCENT_A": 10,
	"A_A_ACCENT": 10,
	"A_A_ACCENT2": 10,
	"A_INFINITIF": 100,
	"A_LE": -50,
	"A_VERBE_INFINITIF": 20,
	"BYTES": 10,
	"CEST_A_DIRE": 100,
	"CONFUSION_AL_LA": -50,
	"CONFUSION_PARLEZ_PARLER": 10,
	"COTE": 10,
	"DE_OU_DES": 20,
	"DU_DU": 100,
	"D_N_E_OU_E": 100,
	"D_VPPA": 20,
	"ELISION": -200,
	"EMPLOI_EMPLOIE": 20,
	"EN_CE_QUI_CONCERNE": -152,
	"EN_MEME_TEMPS": -152,
	"ESPACE_UNITES": 10,
	"EST_CE_QUE": 20,
	"ET_AUSSI": -152,
	"ET_SENT_START": -151,
	"EXPRESSIONS_VU": 100,
	"FAIRE_VPPA": 100,
	"FRENCH_WHITESPACE": -400,
	"FRENCH_WHITESPACE_STRICT": -350,
	"FRENCH_WORD_REPEAT_BEGINNING_RULE": -350,
	"FRENCH_WORD_REPEAT_RULE": -20,
	"FR_SPELLING_RULE": -100,
	"FR_SPLIT_WORDS_HYPHEN": 100,
	"GENS_ACCORD": 100,
	"ILS_VERBE": -50,
	"IL_VERBE": -50,
	"IMP_PRON": -10,
	"INTERROGATIVE_DIRECTE": -10,
	"JE_M_APPEL": 10,
	"JE_SUI": 10,
	"JE_TES": 100,
	"J_N": -10,
	"J_N2": 100,
	"LA_LA2": -20,
	"LEURS_LEUR": 100,
	"LE_COVID": -60,
	"MA": 100,
	"MAIS_AUSSI": -152,
	"MAIS_SENT_START": -151,
	"MOTS_INCOMP": -400,
	"MOT_TRAIT_MOT": -400,
	"MULTI_ADJ": -50,
	"ON_ONT": 100,
	"OU_PAS": 10,
	"PARENTHESES": -50,
	"PAS_DE_SOUCIS": 10,
	"PAS_DE_TRAIT_UNION": 50,
	"PAS_DE_VERBE_APRES_POSSESSIF_DEMONSTRATIF": -20,
	"PEUTETRE": 10,
	"PLACE_DE_LA_VIRGULE": 10,
	"PLURIEL_AL2": 100,
	"POINT": -200,
	"POINTS_2": -400,
	"POINTS_SUSPENSIONS_SPACE": -250,
	"PREP_VERBECONJUGUE": -20,
	"QUASI_NOM": 100,
	"REPETITIONS_STYLE": -250,
	"REP_ESSENTIEL": -50,
	"R_VAVOIR_VINF": 10,
	"SA_CA_SE": 100,
	"SE_CE": -10,
	"SIL_VOUS_PLAIT": 100,
	"SOCIOCULTUREL": 40,
	"SON_SONT": 100,
	"SUJET_AUXILIAIRE": 10,
	"TE_NV": -20,
	"TE_NV2": -10,
	"TOO_LONG_PARAGRAPH": -15,
	"TOUT_MAJUSCULES": -400,
	"TRAIT_UNION": 100,
	"TRES_TRES_ADJ": -10,
	"UPPERCASE_SENTENCE_START": -300,
	"VERBES_FAMILIERS": -25,
	"VERB_PRONOUN": -50,
	"VIRGULE_EXPRESSIONS_FIGEES": 100,
	"VIRGULE_VERBE": -20,
	"VIRG_INF": -100,
	"VIRG_NON_TROUVEE": -400,
	"VOIR_VOIRE": 20,
	"V_J_A_R": -10,
	"Y_A": 10,
}

// FrenchPriorityMap ports French.getPriorityMap (defensive copy).
func FrenchPriorityMap() map[string]int {
	out := make(map[string]int, len(frenchID2Prio))
	for k, v := range frenchID2Prio {
		out[k] = v
	}
	return out
}

// FrenchPriorityForId ports French.getPriorityForId (then Language base).
func FrenchPriorityForId(id string) int {
	if p, ok := frenchID2Prio[id]; ok {
		return p
	}
	if strings.HasPrefix(id, "FR_COMPOUNDS") {
		return 500
	}
	switch id {
	case "CAT_TYPOGRAPHIE", "CAT_TOURS_CRITIQUES", "CAT_HOMONYMES_PARONYMES":
		return 20
	case "SON":
		return -5
	case "CONFUSION_RULE_PREMIUM":
		return -50
	}
	if strings.HasPrefix(id, "CAR") {
		return -50
	}
	if strings.HasPrefix(id, "FR_MULTITOKEN_SPELLING_") {
		return -90
	}
	if strings.HasPrefix(id, "FR_SIMPLE_REPLACE") {
		return 150
	}
	if strings.HasPrefix(id, "grammalecte_") {
		return -150
	}
	if strings.HasPrefix(id, "AI_FR_HYDRA_LEO") {
		return -101
	}
	if strings.HasPrefix(id, "AI_FR_GGEC_REPLACEMENT_ORTHOGRAPHY") {
		return -101
	}
	return languagePriorityForId(id)
}
