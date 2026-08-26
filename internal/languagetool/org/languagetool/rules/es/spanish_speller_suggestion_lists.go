package es

// Static lists from MorfologikSpanishSpellerRule (Java).

var spanishRemoveFromSuggestions = func() map[string]struct{} {
	words := []string{
		"abu", "abue", "abus", "anarco", "anarcos", "arbi", "arbis", "arqui", "arquis", "barna", "bibe", "bibes",
		"biblio", "biblios", "bolche", "bolches", "cami", "camis", "capi", "capis", "celu", "celus", "ceni", "cenis",
		"cerve", "cerves", "chiqui", "chiquis", "chuche", "chuches", "chumi", "chumis", "cintu", "cintus", "comi",
		"comis", "compu", "compus", "confe", "confes", "confi", "confis", "conge", "conges", "copi", "copis",
		"cosquis", "coti", "cotis", "cíber", "deco", "decos", "deli", "delis", "depa", "depas", "díver", "facu",
		"facus", "festi", "festis", "frigo", "frigos", "fácul", "gili", "gilis", "gine", "gineco", "ginecos", "gines",
		"graná", "hospi", "hospis", "ilu", "ilus", "impeque", "impeques", "inge", "inges", "joputa", "joputas",
		"jueputa", "jueputas", "lesbi", "lesbis", "lipo", "lipos", "lito", "litos", "mani", "manifa", "manifas",
		"manis", "mari", "maris", "masoca", "masocas", "milqui", "milquis", "munipa", "munipas", "ofi", "ofis",
		"pandi", "pandis", "pasti", "pastis", "pelu", "pelus", "pendeviejo", "pendeviejos", "peni", "penis", "pisci",
		"piscis", "piti", "pitis", "porfaplís", "porfi", "porfiplís", "porfis", "porsi", "porsiaca", "porsiacas",
		"porsis", "prefe", "prefes", "prince", "princes", "pringui", "pringuis", "prosti", "prostis", "prota",
		"protas", "prote", "protes", "psico", "psicos", "psiqui", "psiquis", "publi", "publis", "puti", "putis",
		"quillo", "quillos", "refri", "refris", "regu", "regus", "repe", "repes", "resi", "resis", "ridi", "ridis",
		"rotu", "rotus", "sado", "sados", "soco", "socos", "sufi", "sufis", "suje", "sujes", "tatu", "tatus",
		"torti", "tortis", "tranqui", "tranquis", "trici", "tricis", "ulti", "ultis", "urba", "urbas", "vice",
		"vices", "vitro", "vitros", "ñero", "ñeros",
	}
	m := make(map[string]struct{}, len(words))
	for _, w := range words {
		m[w] = struct{}{}
	}
	return m
}()

var spanishPrefixWithWhitespace = func() map[string]struct{} {
	words := []string{
		"ultra", "eco", "tele", "anti", "auto", "ex", "extra", "macro", "mega", "meta", "micro", "multi",
		"mono", "mini", "post", "retro", "semi", "super", "hiper", "trans", "re", "g", "l", "m",
	}
	m := make(map[string]struct{}, len(words))
	for _, w := range words {
		m[w] = struct{}{}
	}
	return m
}()

var spanishParticulaFinal = map[string]struct{}{
	"que": {}, "cual": {},
}

// Java PRONOMBRE_INICIAL
var spanishPronombreInicial = func() map[string]struct{} {
	words := []string{"me", "te", "se", "nos", "os", "lo", "le", "la", "los", "las"}
	m := make(map[string]struct{}, len(words))
	for _, w := range words {
		m[w] = struct{}{}
	}
	return m
}()
