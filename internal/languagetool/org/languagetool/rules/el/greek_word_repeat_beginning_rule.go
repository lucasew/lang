package el

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GreekWordRepeatBeginningRule ports org.languagetool.rules.el.GreekWordRepeatBeginningRule.
type GreekWordRepeatBeginningRule struct {
	*rules.WordRepeatBeginningRule
}

var (
	elAddAdverbs = map[string]bool{
		"Επίσης": true, "Επιπρόσθετα": true, "Ακόμη": true, "Επιπλέον": true, "Συμπληρωματικά": true,
	}
	elContrastAdverbs = map[string]bool{
		"Αντίθετα": true, "Ωστόσο": true, "Εντούτοις": true, "Εξάλλου": true,
	}
	elExplainAdverbs = map[string]bool{
		"Δηλαδή": true, "Ειδικότερα": true, "Ειδικά": true, "Συγκεκριμένα": true,
	}
)

func NewGreekWordRepeatBeginningRule(messages map[string]string) *GreekWordRepeatBeginningRule {
	base := rules.NewWordRepeatBeginningRule(messages)
	base.IDOverride = "GREEK_WORD_REPEAT_BEGINNING_RULE"
	// Java: Επίσης → Επιπλέον
	base.AddExamplePair(
		rules.Wrong("Επίσης, παίζω ποδόσφαιρο. <marker>Επίσης</marker>, παίζω μπάσκετ."),
		rules.Fixed("Επίσης, παίζω ποδόσφαιρο. <marker>Επιπλέον</marker>, παίζω μπάσκετ."),
	)
	r := &GreekWordRepeatBeginningRule{WordRepeatBeginningRule: base}
	base.IsExceptionFn = r.isException
	base.IsAdverbFn = r.isAdverb
	base.GetSuggestionsFn = r.getSuggestions
	return r
}

func (r *GreekWordRepeatBeginningRule) isException(token string) bool {
	switch token {
	case "Ο", "Η", "Το", "Οι", "Τα", "Ένας", "Μία", "Ένα":
		return true
	}
	return false
}

func (r *GreekWordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	tok := token.GetToken()
	return elAddAdverbs[tok] || elContrastAdverbs[tok] || elExplainAdverbs[tok]
}

func (r *GreekWordRepeatBeginningRule) getSuggestions(token *languagetool.AnalyzedTokenReadings) []string {
	tok := token.GetToken()
	if elAddAdverbs[tok] {
		return differentFrom(tok, elAddAdverbs)
	}
	if elContrastAdverbs[tok] {
		return differentFrom(tok, elContrastAdverbs)
	}
	if elExplainAdverbs[tok] {
		return differentFrom(tok, elExplainAdverbs)
	}
	return nil
}

func differentFrom(adverb string, category map[string]bool) []string {
	var out []string
	for adv := range category {
		if adv != adverb {
			out = append(out, adv)
		}
	}
	return out
}
