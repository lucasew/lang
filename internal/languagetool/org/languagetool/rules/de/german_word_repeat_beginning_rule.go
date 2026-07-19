package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanWordRepeatBeginningRule ports org.languagetool.rules.de.GermanWordRepeatBeginningRule.
type GermanWordRepeatBeginningRule struct {
	*rules.WordRepeatBeginningRule
}

var germanAdverbs = map[string]bool{
	"Auch": true, "Anschließend": true, "Außerdem": true, "Danach": true, "Ferner": true,
	"Nebenher": true, "Nebenbei": true, "Überdies": true, "Weiterführend": true,
	"Zudem": true, "Zusätzlich": true,
}

func NewGermanWordRepeatBeginningRule(messages map[string]string) *GermanWordRepeatBeginningRule {
	base := rules.NewWordRepeatBeginningRule(messages)
	base.IDOverride = "GERMAN_WORD_REPEAT_BEGINNING_RULE"
	// Java: Dann… Dann… <marker>Dann</marker> → Schließlich
	base.AddExamplePair(
		rules.Wrong("Dann hatten wir Freizeit. Dann gab es Essen. <marker>Dann</marker> gingen wir schlafen."),
		rules.Fixed("Dann hatten wir Freizeit. Danach gab es Essen. <marker>Schließlich</marker> gingen wir schlafen."),
	)
	r := &GermanWordRepeatBeginningRule{WordRepeatBeginningRule: base}
	base.IsAdverbFn = r.isAdverb
	return r
}

func (r *GermanWordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	return germanAdverbs[token.GetToken()]
}
