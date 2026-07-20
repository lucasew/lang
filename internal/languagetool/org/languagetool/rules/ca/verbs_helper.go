package ca

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// VerbsHelper ports org.languagetool.rules.ca.VerbsHelper.
var verbsDicendi = map[string]struct{}{
	"acceptar":     {},
	"aclarir":      {},
	"aconsellar":   {},
	"acusar":       {},
	"adduir":       {},
	"admetre":      {},
	"advertir":     {},
	"afegir":       {},
	"afirmar":      {},
	"afluixar":     {},
	"agregar":      {},
	"al·legar":     {},
	"al·ludir":     {},
	"amenaçar":     {},
	"amollar":      {},
	"amonestar":    {},
	"ampliar":      {},
	"anunciar":     {},
	"apuntar":      {},
	"argumentar":   {},
	"assegurar":    {},
	"assentir":     {},
	"assenyalar":   {},
	"atorgar":      {},
	"atribuir":     {},
	"avançar":      {},
	"avisar":       {},
	"barbotejar":   {},
	"bordar":       {},
	"bramar":       {},
	"calcular":     {},
	"callar":       {},
	"citar":        {},
	"comentar":     {},
	"concedir":     {},
	"concloure":    {},
	"concretar":    {},
	"confessar":    {},
	"confiar":      {},
	"confirmar":    {},
	"considerar":   {},
	"contestar":    {},
	"creure":       {},
	"cridar":       {},
	"culpar":       {},
	"decidir":      {},
	"declamar":     {},
	"declarar":     {},
	"decretar":     {},
	"defensar":     {},
	"definir":      {},
	"delimitar":    {},
	"demanar":      {},
	"descobrir":    {},
	"descriure":    {},
	"desitjar":     {},
	"desmentir":    {},
	"destacar":     {},
	"desvelar":     {},
	"detallar":     {},
	"determinar":   {},
	"dir":          {},
	"dogmatitzar":  {},
	"dubtar":       {},
	"elogiar":      {},
	"emfasitzar":   {},
	"emfatitzar":   {},
	"engaltar":     {},
	"engegar":      {},
	"entaferrar":   {},
	"enumerar":     {},
	"esclafir":     {},
	"escopir":      {},
	"escridassar":  {},
	"esgrimir":     {},
	"esmentar":     {},
	"especificar":  {},
	"espletar":     {},
	"establir":     {},
	"etzibar":      {},
	"exclamar":     {},
	"exigir":       {},
	"explicar":     {},
	"exposar":      {},
	"expressar":    {},
	"formular":     {},
	"garantir":     {},
	"gemegar":      {},
	"imaginar":     {},
	"implorar":     {},
	"imputar":      {},
	"increpar":     {},
	"indicar":      {},
	"informar":     {},
	"inquirir":     {},
	"insinuar":     {},
	"insistir":     {},
	"insultar":     {},
	"interrogar":   {},
	"intervenir":   {},
	"ironitzar":    {},
	"jurar":        {},
	"justificar":   {},
	"lamentar":     {},
	"lladrar":      {},
	"lloar":        {},
	"maleir":       {},
	"manar":        {},
	"manifestar":   {},
	"matisar":      {},
	"mentir":       {},
	"mostrar":      {},
	"murmurar":     {},
	"negar":        {},
	"observar":     {},
	"oferir":       {},
	"opinar":       {},
	"ordenar":      {},
	"pensar":       {},
	"plantejar":    {},
	"pontificar":   {},
	"pregar":       {},
	"preguntar":    {},
	"presumir":     {},
	"preveure":     {},
	"prometre":     {},
	"proposar":     {},
	"protestar":    {},
	"puntualitzar": {},
	"quequejar":    {},
	"ratificar":    {},
	"reafirmar":    {},
	"rebutjar":     {},
	"recalcar":     {},
	"recitar":      {},
	"reclamar":     {},
	"recomanar":    {},
	"reconèixer":   {},
	"referir":      {},
	"refermar":     {},
	"reflexionar":  {},
	"refusar":      {},
	"refutar":      {},
	"relatar":      {},
	"remarcar":     {},
	"rematar":      {},
	"remugar":      {},
	"renegar":      {},
	"renyar":       {},
	"repetir":      {},
	"replicar":     {},
	"reprendre":    {},
	"resar":        {},
	"respondre":    {},
	"retreure":     {},
	"revelar":      {},
	"soltar":       {},
	"sol·licitar":  {},
	"somicar":      {},
	"sospirar":     {},
	"sospitar":     {},
	"sostenir":     {},
	"subratllar":   {},
	"suggerir":     {},
	"suposar":      {},
	"trobar":       {},
	"xisclar":      {},
	"xiuxiuejar":   {},
}

// IsVerbDicendi reports whether lemma is a Catalan verbum dicendi.
func IsVerbDicendi(lemma string) bool {
	_, ok := verbsDicendi[strings.ToLower(lemma)]
	return ok
}

// IsVerbDicendiBefore scans lemmas from i backwards while keepLooking is true.
// keepLooking(i) should be true for V.*/RG.*/LOC_ADV-like tokens.
func IsVerbDicendiBefore(lemmas []string, i int, keepLooking func(int) bool) bool {
	for i > 0 && i < len(lemmas) && keepLooking(i) {
		if IsVerbDicendi(lemmas[i]) {
			return true
		}
		i--
	}
	return false
}

// Java VerbsHelper.pKeepLooking
var pKeepLooking = regexp.MustCompile(`^(V.*|RG.*|LOC_ADV)$`)

// IsVerbDicendiBeforeTokens ports VerbsHelper.isVerbDicendiBefore(tokens, i).
// Scans backward from i while a reading matches V.*|RG.*|LOC_ADV; true if lemma is dicendi.
func IsVerbDicendiBeforeTokens(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	for i > 0 && i < len(tokens) {
		tok := tokens[i]
		if tok == nil {
			return false
		}
		var reading *languagetool.AnalyzedToken
		for _, r := range tok.GetReadings() {
			if r == nil {
				continue
			}
			p := r.GetPOSTag()
			if p != nil && pKeepLooking.MatchString(*p) {
				reading = r
				break
			}
		}
		if reading == nil {
			return false
		}
		if lem := reading.GetLemma(); lem != nil && IsVerbDicendi(*lem) {
			return true
		}
		// Java continues while reading != null even if lemma not dicendi
		i--
	}
	return false
}
