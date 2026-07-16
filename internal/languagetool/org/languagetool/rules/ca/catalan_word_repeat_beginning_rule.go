package ca

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// CatalanWordRepeatBeginningRule ports org.languagetool.rules.ca.CatalanWordRepeatBeginningRule.
type CatalanWordRepeatBeginningRule struct {
	*rules.WordRepeatBeginningRule
}

var (
	caAddAdverbs = map[string]bool{
		"Igualment": true, "També": true, "Addicionalment": true,
	}
	caContrastConj = map[string]bool{"Però": true, "Emperò": true, "Mes": true}
	caCauseConj    = map[string]bool{"Perquè": true, "Car": true}
	caEmphasis     = map[string]bool{
		"Òbviament": true, "Clarament": true, "Absolutament": true, "Definitivament": true,
	}
	caExplain = map[string]bool{
		"Específicament": true, "Concretament": true, "Particularment": true, "Precisament": true,
	}
	caAddExpr      = []string{"Així mateix", "A més a més"}
	caContrastExpr = []string{"Així i tot", "D'altra banda", "Per altra part"}
	caCauseExpr    = []string{"Ja que", "Per tal com", "Pel fet que", "Puix que"}
	caPronouns     = map[string]bool{
		"jo": true, "tu": true, "ell": true, "ella": true, "nosaltres": true, "vosaltres": true,
		"ells": true, "elles": true, "vostè": true, "vostès": true, "vosté": true, "vostés": true, "vós": true,
	}
	caExceptions = map[string]bool{
		"l'": true, "el": true, "la": true, "els": true, "les": true, "punt": true, "article": true,
		"mòdul": true, "part": true, "sessió": true, "unitat": true, "tema": true,
		"a": true, "per": true, "en": true, "com": true,
	}
)

func NewCatalanWordRepeatBeginningRule(messages map[string]string) *CatalanWordRepeatBeginningRule {
	base := rules.NewWordRepeatBeginningRule(messages)
	base.IDOverride = "CATALAN_WORD_REPEAT_BEGINNING_RULE"
	r := &CatalanWordRepeatBeginningRule{WordRepeatBeginningRule: base}
	base.IsExceptionFn = r.isException
	base.IsAdverbFn = r.isAdverb
	base.GetSuggestionsFn = r.getSuggestions
	return r
}

func (r *CatalanWordRepeatBeginningRule) isException(token string) bool {
	if token == "" {
		return false
	}
	if unicode.IsDigit([]rune(token)[0]) {
		return true
	}
	return caExceptions[strings.ToLower(token)]
}

func (r *CatalanWordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	if token.HasPosTag("RG") || token.HasPosTag("LOC_ADV") {
		return true
	}
	tok := token.GetToken()
	return caAddAdverbs[tok] || caContrastConj[tok] || caEmphasis[tok] || caExplain[tok] || caCauseConj[tok]
}

func (r *CatalanWordRepeatBeginningRule) getSuggestions(token *languagetool.AnalyzedTokenReadings) []string {
	tok := token.GetToken()
	lower := strings.ToLower(tok)
	if caPronouns[lower] {
		return []string{
			"A més a més, " + lower,
			"Igualment, " + lower,
			"No sols aixó, sinó que " + lower,
		}
	}
	if caAddAdverbs[tok] {
		s := differentMap(tok, caAddAdverbs)
		return append(s, caAddExpr...)
	}
	if caContrastConj[tok] {
		return append([]string(nil), caContrastExpr...)
	}
	if caEmphasis[tok] {
		return differentMap(tok, caEmphasis)
	}
	if caExplain[tok] {
		return differentMap(tok, caExplain)
	}
	if caCauseConj[tok] {
		s := differentMap(tok, caCauseConj)
		return append(s, caCauseExpr...)
	}
	return nil
}

func differentMap(adverb string, category map[string]bool) []string {
	var out []string
	for adv := range category {
		if adv != adverb {
			out = append(out, adv)
		}
	}
	return out
}
