package de

import (
	"sort"
	"strings"
)

// LanguageNames ports org.languagetool.rules.de.LanguageNames — set of language adjectives.
var languageNames = map[string]struct{}{
	"Angelsächsisch":     {},
	"Afrikanisch":        {},
	"Albanisch":          {},
	"Altarabisch":        {},
	"Altchinesisch":      {},
	"Altgriechisch":      {},
	"Althochdeutsch":     {},
	"Altpersisch":        {},
	"Amerikanisch":       {},
	"Arabisch":           {},
	"Armenisch":          {},
	"Bairisch":           {},
	"Baskisch":           {},
	"Bengalisch":         {},
	"Bulgarisch":         {},
	"Chinesisch":         {},
	"Dänisch":            {},
	"Deutsch":            {},
	"Englisch":           {},
	"Estnisch":           {},
	"Finnisch":           {},
	"Französisch":        {},
	"Frühneuhochdeutsch": {},
	"Germanisch":         {},
	"Georgisch":          {},
	"Griechisch":         {},
	"Hebräisch":          {},
	"Hocharabisch":       {},
	"Hochchinesisch":     {},
	"Hochdeutsch":        {},
	"Holländisch":        {},
	"Indonesisch":        {},
	"Irisch":             {},
	"Isländisch":         {},
	"Italienisch":        {},
	"Japanisch":          {},
	"Jiddisch":           {},
	"Jugoslawisch":       {},
	"Kantonesisch":       {},
	"Katalanisch":        {},
	"Klingonisch":        {},
	"Koreanisch":         {},
	"Kroatisch":          {},
	"Kurdisch":           {},
	"Lateinisch":         {},
	"Lettisch":           {},
	"Litauisch":          {},
	"Luxemburgisch":      {},
	"Mittelhochdeutsch":  {},
	"Mongolisch":         {},
	"Neuhochdeutsch":     {},
	"Niederländisch":     {},
	"Norwegisch":         {},
	"Persisch":           {},
	"Plattdeutsch":       {},
	"Polnisch":           {},
	"Portugiesisch":      {},
	"Rätoromanisch":      {},
	"Rumänisch":          {},
	"Russisch":           {},
	"Sächsisch":          {},
	"Schwäbisch":         {},
	"Schwedisch":         {},
	"Schweizerisch":      {},
	"Serbisch":           {},
	"Serbokroatisch":     {},
	"Slawisch":           {},
	"Slowakisch":         {},
	"Slowenisch":         {},
	"Spanisch":           {},
	"Syrisch":            {},
	"Tamilisch":          {},
	"Tibetisch":          {},
	"Tschechisch":        {},
	"Tschetschenisch":    {},
	"Türkisch":           {},
	"Turkmenisch":        {},
	"Uigurisch":          {},
	"Ukrainisch":         {},
	"Ungarisch":          {},
	"Usbekisch":          {},
	"Vietnamesisch":      {},
	"Walisisch":          {},
	"Weißrussisch":       {},
}

// IsLanguageName reports whether s is a known German language-name adjective.
func IsLanguageName(s string) bool {
	_, ok := languageNames[s]
	return ok
}

// LanguageNamesGetAsRegex ports LanguageNames.getAsRegex().
func LanguageNamesGetAsRegex() string {
	keys := make([]string, 0, len(languageNames))
	for k := range languageNames {
		keys = append(keys, k)
	}
	sort.Strings(keys) // deterministic (Java HashSet join order is not; patterns still match)
	return strings.Join(keys, "|")
}

// LanguageNames is the Java-name twin for the language-adjective set.
type LanguageNames struct{}

func (LanguageNames) Contains(s string) bool { return IsLanguageName(s) }

func (LanguageNames) GetAsRegex() string { return LanguageNamesGetAsRegex() }

func (LanguageNames) All() []string {
	out := make([]string, 0, len(languageNames))
	for k := range languageNames {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
