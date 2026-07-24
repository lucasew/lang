package ca

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_diacritics_iec.txt
var diacriticsIECFS embed.FS

var (
	diacriticsIECOnce sync.Once
	diacriticsIECMap  map[string][]string
)

func loadDiacriticsIEC() map[string][]string {
	diacriticsIECOnce.Do(func() {
		f, err := diacriticsIECFS.Open("data/replace_diacritics_iec.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		diacriticsIECMap = m
	})
	return diacriticsIECMap
}

// SimpleReplaceDiacriticsIEC ports org.languagetool.rules.ca.SimpleReplaceDiacriticsIEC.
// Java isTokenException: hasPosTagStartingWith("NP") only.
type SimpleReplaceDiacriticsIEC struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleReplaceDiacriticsIEC(messages map[string]string) *SimpleReplaceDiacriticsIEC {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:           messages,
		WrongWords:         loadDiacriticsIEC(),
		CaseSensitive:      false,
		CheckLemmas:        false,
		ID:                 "CA_SIMPLE_REPLACE_DIACRITICS_IEC",
		LanguageCode:       "ca",
		SubRuleSpecificIDs: true,
		Description:        "Accents diacrítics segons les normes noves (2017): $match",
		ShortMsg:           "Hi sobra l'accent.",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Hi sobra l'accent diacrític (segons les normes noves)."
		},
		TokenException: diacriticsIECTokenException,
	}
	return &SimpleReplaceDiacriticsIEC{AbstractSimpleReplaceRule: base}
}

// diacriticsIECTokenException ports SimpleReplaceDiacriticsIEC.isTokenException.
func diacriticsIECTokenException(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return false
	}
	return token.HasPosTagStartingWith("NP")
}

func (r *SimpleReplaceDiacriticsIEC) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
