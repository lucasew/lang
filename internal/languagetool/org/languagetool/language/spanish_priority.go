package language

import "strings"

// Spanish rule priorities from org.languagetool.language.Spanish (id2prio + getPriorityForId).
// Java is king — do not invent extra IDs.

var spanishID2Prio = map[string]int{
	"AGREEMENT_ADJ_NOUN": -30,
	"AGREEMENT_ADJ_NOUN_AREA": 30,
	"AGREEMENT_DET_ABREV": 25,
	"AGREEMENT_DET_ADJ": 10,
	"AGREEMENT_DET_NOUN": 15,
	"AGREEMENT_DET_NOUN_EXCEPTIONS": 25,
	"AGREEMENT_PARTICIPLE_NOUN": -30,
	"AGREEMENT_POSTPONED_ADJ": -30,
	"COMMA_SINO": -40,
	"COMMA_SINO2": -40,
	"CONFUSION_ES_SE": 20,
	"DEGREE_CHAR": 30,
	"DE_TILDE": 50,
	"EL_NO_TILDE": 40,
	"EL_TILDE": -10,
	"ES_QUESTION_MARK": -250,
	"ES_SIMPLE_REPLACE_MULTIWORDS": 50,
	"ES_SPLIT_WORDS": -10,
	"ETCETERA": 30,
	"HALLA_HAYA": 10,
	"LOS_MAPUCHE": 50,
	"LO_LOS": 30,
	"MORFOLOGIK_RULE_ES": -100,
	"MUCHO_NF": 25,
	"MULTI_ADJ": -30,
	"NO_SEPARADO": 40,
	"PARTICIPIO_MS": 40,
	"PERSONAJES_FAMOSOS": 50,
	"PHRASE_REPETITION": -150,
	"PLURAL_SEPARADO": 50,
	"POR_CIERTO": 30,
	"PREP_VERB": -20,
	"PRIMER_PRIMERA": 20,
	"PRONOMBRE_SIN_VERBO": 25,
	"P_EJ": 30,
	"REPETITIONS_STYLE": -50,
	"SEPARADO": 1,
	"SE_CREO": 35,
	"SE_CREO2": 25,
	"SINGLE_CHARACTER": -15,
	"SI_AFIRMACION": 10,
	"SPANISH_WORD_REPEAT_RULE": -150,
	"SUBJUNTIVO_FUTURO": -30,
	"SUBJUNTIVO_INCORRECTO": -40,
	"SUBJUNTIVO_PASADO": -30,
	"SUBJUNTIVO_PASADO2": -30,
	"TE_TILDE": 50,
	"TE_TILDE2": 10,
	"TOO_LONG_PARAGRAPH": -15,
	"TYPOGRAPHY": 20,
	"UPPERCASE_SENTENCE_START": -200,
	"U_NO": -10,
	"VALLA_VAYA": 10,
	"VERBO_MODAL_INFINITIVO": 40,
	"VOSEO": -40,
}

// SpanishPriorityMap ports Spanish.getPriorityMap (defensive copy).
func SpanishPriorityMap() map[string]int {
	out := make(map[string]int, len(spanishID2Prio))
	for k, v := range spanishID2Prio {
		out[k] = v
	}
	return out
}

// SpanishPriorityForId ports Spanish.getPriorityForId (then Language base).
func SpanishPriorityForId(id string) int {
	// Java checks some ids before the map.
	switch id {
	case "CONFUSIONS2":
		return 50
	case "RARE_WORDS":
		return 50
	case "MISSPELLING":
		return 40
	case "CONFUSIONS":
		return 40
	case "INCORRECT_EXPRESSIONS":
		return 40
	case "DIACRITICS":
		return 30
	}
	if strings.HasPrefix(id, "ES_SIMPLE_REPLACE_SIMPLE") {
		return 30
	}
	if strings.HasPrefix(id, "ES_COMPOUNDS") {
		return 50
	}
	if p, ok := spanishID2Prio[id]; ok {
		return p
	}
	if strings.HasPrefix(id, "AI_ES_HYDRA_LEO") {
		return -101
	}
	if strings.HasPrefix(id, "AI_ES_GGEC") {
		if id == "AI_ES_GGEC_REPLACEMENT_OTHER" {
			return -300
		}
		return 0
	}
	if strings.HasPrefix(id, "ES_MULTITOKEN_SPELLING") {
		return -95
	}
	return languagePriorityForId(id)
}
