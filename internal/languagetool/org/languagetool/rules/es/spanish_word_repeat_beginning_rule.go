package es

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SpanishWordRepeatBeginningRule ports org.languagetool.rules.es.SpanishWordRepeatBeginningRule.
type SpanishWordRepeatBeginningRule struct {
	*rules.WordRepeatBeginningRule
}

var (
	esAddAdverbs = map[string]bool{
		"Asimismo": true, "Igualmente": true, "Además": true, "También": true, "Adicionalmente": true,
	}
	esContrastConj = map[string]bool{
		"Pero": true, "Empero": true, "Mas": true,
	}
	esEmphasisAdverbs = map[string]bool{
		"Obviamente": true, "Claramente": true, "Absolutamente": true, "Definitivamente": true,
	}
	esExplainAdverbs = map[string]bool{
		"Específicamente": true, "Concretamente": true, "Particularmente": true, "Precisamente": true,
	}
	// Common sentence-start adverbs (POS RG in Java tagger; surface stand-in).
	esExtraAdverbs = map[string]bool{
		"Mañana": true, "Hoy": true, "Ayer": true, "Luego": true, "Después": true,
		"Entonces": true, "Así": true, "Ahora": true, "Antes": true, "Pronto": true,
	}
	esAddExpressions      = []string{"Así mismo"}
	esContrastExpressions = []string{"Aun así", "Por otra parte", "Sin embargo"}
	esPersonalPronouns    = map[string]bool{
		"yo": true, "tú": true, "él": true, "ella": true, "nosostros": true, "nosotras": true,
		"vosotros": true, "vosotras": true, "ellos": true, "ellas": true, "usted": true, "ustedes": true,
	}
	esExceptionStarts = map[string]bool{
		"el": true, "la": true, "los": true, "las": true, "punto": true, "artículo": true,
		"módulo": true, "parte": true, "sesión": true, "unidad": true, "tema": true, "n": true,
	}
	esSentenceExceptions = []string{"por un", "por otro", "por otra", "por una"}
)

func NewSpanishWordRepeatBeginningRule(messages map[string]string) *SpanishWordRepeatBeginningRule {
	base := rules.NewWordRepeatBeginningRule(messages)
	base.IDOverride = "SPANISH_WORD_REPEAT_BEGINNING_RULE"
	r := &SpanishWordRepeatBeginningRule{WordRepeatBeginningRule: base}
	base.IsExceptionFn = r.isException
	base.IsAdverbFn = r.isAdverb
	base.IsAdverbAtFn = r.isAdverbAt
	base.GetSuggestionsFn = r.getSuggestions
	base.IsSentenceException = r.isSentenceException
	return r
}

func (r *SpanishWordRepeatBeginningRule) isException(token string) bool {
	if token == "" {
		return false
	}
	if unicode.IsDigit([]rune(token)[0]) {
		return true
	}
	return esExceptionStarts[strings.ToLower(token)]
}

func (r *SpanishWordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	if token.HasPosTag("RG") || token.HasPosTag("LOC_ADV") {
		return true
	}
	tok := token.GetToken()
	return esAddAdverbs[tok] || esContrastConj[tok] || esEmphasisAdverbs[tok] ||
		esExplainAdverbs[tok] || esExtraAdverbs[tok]
}

func (r *SpanishWordRepeatBeginningRule) isAdverbAt(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if i >= 0 && i < len(tokens) && r.isAdverb(tokens[i]) {
		return true
	}
	// Multiword "Sin embargo" (split by WordTokenizer).
	if i >= 0 && i+1 < len(tokens) &&
		tokens[i].GetToken() == "Sin" && tokens[i+1].GetToken() == "embargo" {
		return true
	}
	return false
}

func (r *SpanishWordRepeatBeginningRule) getSuggestions(token *languagetool.AnalyzedTokenReadings) []string {
	tok := token.GetToken()
	lower := strings.ToLower(tok)
	if esPersonalPronouns[lower] {
		return []string{
			"Además, " + lower,
			"Igualmente, " + lower,
			"No solo eso, sino que " + lower,
		}
	}
	if esAddAdverbs[tok] {
		// Order matches Java test stringification expectation.
		order := []string{"Adicionalmente", "Asimismo", "Además", "Igualmente", "También"}
		var s []string
		for _, a := range order {
			if a != tok && esAddAdverbs[a] {
				s = append(s, a)
			}
		}
		s = append(s, esAddExpressions...)
		return s
	}
	if esContrastConj[tok] {
		return append([]string(nil), esContrastExpressions...)
	}
	if esEmphasisAdverbs[tok] {
		return differentFromMap(tok, esEmphasisAdverbs)
	}
	if esExplainAdverbs[tok] {
		return differentFromMap(tok, esExplainAdverbs)
	}
	return nil
}

func (r *SpanishWordRepeatBeginningRule) isSentenceException(sentence *languagetool.AnalyzedSentence) bool {
	s := strings.ToLower(sentence.GetText())
	for _, ex := range esSentenceExceptions {
		if strings.HasPrefix(s, ex) {
			return true
		}
	}
	return false
}

func differentFromMap(adverb string, category map[string]bool) []string {
	var out []string
	for adv := range category {
		if adv != adverb {
			out = append(out, adv)
		}
	}
	return out
}
