package de

// LanguageNames ports org.languagetool.rules.de.LanguageNames — set of language adjectives.
var languageNames = map[string]struct{}{
	"Angelsächsisch": {}, "Afrikanisch": {}, "Albanisch": {}, "Altarabisch": {},
	"Altchinesisch": {}, "Altgriechisch": {}, "Althochdeutsch": {}, "Altpersisch": {},
	"Amerikanisch": {}, "Arabisch": {}, "Armenisch": {}, "Bairisch": {}, "Baskisch": {},
	"Bengalisch": {}, "Bulgarisch": {}, "Chinesisch": {}, "Dänisch": {}, "Deutsch": {},
	"Englisch": {}, "Estnisch": {}, "Finnisch": {}, "Französisch": {}, "Frühneuhochdeutsch": {},
	"Germanisch": {}, "Georgisch": {}, "Griechisch": {}, "Hebräisch": {}, "Hocharabisch": {},
	"Hochchinesisch": {}, "Hochdeutsch": {}, "Holländisch": {}, "Indonesisch": {},
	"Irisch": {}, "Isländisch": {}, "Italienisch": {}, "Japanisch": {}, "Jiddisch": {},
	"Jugoslawisch": {}, "Kantonesisch": {}, "Katalanisch": {}, "Klingonisch": {},
	"Koreanisch": {}, "Kroatisch": {}, "Kurdisch": {}, "Lateinisch": {}, "Lettisch": {},
	"Litauisch": {}, "Luxemburgisch": {}, "Mittelhochdeutsch": {}, "Mongolisch": {},
	"Neuhochdeutsch": {}, "Niederländisch": {}, "Norwegisch": {}, "Persisch": {},
	"Polnisch": {}, "Portugiesisch": {}, "Rumänisch": {}, "Russisch": {}, "Schwedisch": {},
	"Schweizerdeutsch": {}, "Serbisch": {}, "Slowakisch": {}, "Slowenisch": {},
	"Spanisch": {}, "Syrisch": {}, "Tschechisch": {}, "Türkisch": {}, "Ukrainisch": {},
	"Ungarisch": {}, "Vietnamesisch": {}, "Walisisch": {},
}

// IsLanguageName reports whether s is a known German language-name adjective.
func IsLanguageName(s string) bool {
	_, ok := languageNames[s]
	return ok
}

// LanguageNames is the Java-name twin for the language-adjective set.
type LanguageNames struct{}

func (LanguageNames) Contains(s string) bool { return IsLanguageName(s) }

func (LanguageNames) All() []string {
	out := make([]string, 0, len(languageNames))
	for k := range languageNames {
		out = append(out, k)
	}
	return out
}
