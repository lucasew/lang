package el

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GreekWordRepeatBeginningRule ports org.languagetool.rules.el.GreekWordRepeatBeginningRule.
// Adds Greek ADD/CONTRAST/EXPLAIN adverb sets, article exceptions, and synonym suggestions
// on top of core WordRepeatBeginningRule.
type GreekWordRepeatBeginningRule struct {
	*rules.WordRepeatBeginningRule
}

// Ports Java private static final Set<String> ADD_ADVERBS (static block order).
var elAddAdverbs = map[string]bool{
	"Επίσης": true, "Επιπρόσθετα": true, "Ακόμη": true, "Επιπλέον": true, "Συμπληρωματικά": true,
}

// Ports Java private static final Set<String> CONTRAST_ADVERBS.
var elContrastAdverbs = map[string]bool{
	"Αντίθετα": true, "Ωστόσο": true, "Εντούτοις": true, "Εξάλλου": true,
}

// Ports Java private static final Set<String> EXPLAIN_ADVERBS.
var elExplainAdverbs = map[string]bool{
	"Δηλαδή": true, "Ειδικότερα": true, "Ειδικά": true, "Συγκεκριμένα": true,
}

// NewGreekWordRepeatBeginningRule ports GreekWordRepeatBeginningRule(messages, language).
// Language is unused by this type and the core match loop (same as other language WRB ports).
func NewGreekWordRepeatBeginningRule(messages map[string]string) *GreekWordRepeatBeginningRule {
	base := rules.NewWordRepeatBeginningRule(messages)
	// Java @Override getId
	base.IDOverride = "GREEK_WORD_REPEAT_BEGINNING_RULE"
	// Java addExamplePair(Example.wrong(...), Example.fixed(...))
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

// isException ports GreekWordRepeatBeginningRule.isException:
//   return super.isException(token) || Ο|Η|Το|Οι|Τα|Ένας|Μία|Ένα
// Base WordRepeatBeginningRule.isException already ORs IsExceptionFn with the core list
// (: – - ✔️ ➡️ — ⭐️ ⚠️), so only Greek article extras live here (same pattern as EN The/A/An).
func (r *GreekWordRepeatBeginningRule) isException(token string) bool {
	switch token {
	case "Ο", "Η", "Το", "Οι", "Τα", "Ένας", "Μία", "Ένα":
		return true
	}
	return false
}

// isAdverb ports GreekWordRepeatBeginningRule.isAdverb (fixed surface sets only, not POS).
func (r *GreekWordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return false
	}
	tok := token.GetToken()
	return elAddAdverbs[tok] || elContrastAdverbs[tok] || elExplainAdverbs[tok]
}

// getSuggestions ports GreekWordRepeatBeginningRule.getSuggestions:
// other members of the same ADD/CONTRAST/EXPLAIN set (stream filter !equals(tok)).
func (r *GreekWordRepeatBeginningRule) getSuggestions(token *languagetool.AnalyzedTokenReadings) []string {
	if token == nil {
		return nil
	}
	tok := token.GetToken()
	if elAddAdverbs[tok] {
		return otherAdverbsInSet(tok, elAddAdverbs)
	}
	if elContrastAdverbs[tok] {
		return otherAdverbsInSet(tok, elContrastAdverbs)
	}
	if elExplainAdverbs[tok] {
		return otherAdverbsInSet(tok, elExplainAdverbs)
	}
	// Java: return Collections.emptyList();
	return nil
}

// otherAdverbsInSet ports set.stream().filter(adv -> !adv.equals(tok)).collect(toList()).
// Iteration order is map-undefined (Java HashSet stream is also unordered; tests use anyMatch).
func otherAdverbsInSet(adverb string, category map[string]bool) []string {
	var out []string
	for adv := range category {
		if adv != adverb {
			out = append(out, adv)
		}
	}
	return out
}
