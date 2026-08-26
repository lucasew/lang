package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanWordRepeatBeginningRule ports org.languagetool.rules.de.GermanWordRepeatBeginningRule.
// Extends WordRepeatBeginningRule by a fixed set of German adverbs
// ("Konjunktionaladverbien") for a more precise error message (isAdverb).
//
// Java ctor takes ResourceBundle + Language; Language is unused by this type and
// the core WordRepeatBeginningRule match loop — Go takes messages only (same as
// other language WRB ports).
type GermanWordRepeatBeginningRule struct {
	*rules.WordRepeatBeginningRule
}

// ADVERBS ports Java private static final Set<String> ADVERBS (exact surfaces,
// case-sensitive; Ü in Überdies). Order matches Java Arrays.asList.
var germanWordRepeatBeginningAdverbs = func() map[string]bool {
	// Java:
	// "Auch", "Anschließend", "Außerdem", "Danach", "Ferner",
	// "Nebenher", "Nebenbei", "Überdies", "Weiterführend", "Zudem", "Zusätzlich"
	m := make(map[string]bool, 11)
	for _, w := range []string{
		"Auch", "Anschließend", "Außerdem", "Danach", "Ferner",
		"Nebenher", "Nebenbei", "Überdies", "Weiterführend", "Zudem", "Zusätzlich",
	} {
		m[w] = true
	}
	return m
}()

// NewGermanWordRepeatBeginningRule ports GermanWordRepeatBeginningRule(messages, language).
func NewGermanWordRepeatBeginningRule(messages map[string]string) *GermanWordRepeatBeginningRule {
	base := rules.NewWordRepeatBeginningRule(messages)
	// Java @Override getId
	base.IDOverride = "GERMAN_WORD_REPEAT_BEGINNING_RULE"
	// Java addExamplePair(Example.wrong(...), Example.fixed(...))
	base.AddExamplePair(
		rules.Wrong("Dann hatten wir Freizeit. Dann gab es Essen. <marker>Dann</marker> gingen wir schlafen."),
		rules.Fixed("Dann hatten wir Freizeit. Danach gab es Essen. <marker>Schließlich</marker> gingen wir schlafen."),
	)
	r := &GermanWordRepeatBeginningRule{WordRepeatBeginningRule: base}
	// Java @Override isAdverb
	base.IsAdverbFn = r.isAdverb
	return r
}

// isAdverb ports GermanWordRepeatBeginningRule.isAdverb:
// return ADVERBS.contains(token.getToken());
func (r *GermanWordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return false
	}
	return germanWordRepeatBeginningAdverbs[token.GetToken()]
}
