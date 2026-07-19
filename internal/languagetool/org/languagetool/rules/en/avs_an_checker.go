package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

func init() {
	// Wire faithful checker for languagetool.SimpleAvsAnChecker / demos.
	languagetool.PreferredAvsAnChecker = AvsAnSentenceChecker()
}

// AvsAnSentenceChecker returns a SentenceChecker using the faithful AvsAnRule.
// Injects DT on a/an/the (Java EnglishTagger always tags these) so AnalyzePlain demos work
// without inventing a separate soft phonetic lexicon.
func AvsAnSentenceChecker() languagetool.SentenceChecker {
	r := NewAvsAnRule(nil)
	return func(sentence *languagetool.AnalyzedSentence) []languagetool.LocalMatch {
		if sentence == nil {
			return nil
		}
		injectClosedClassDT(sentence)
		return rules.ToLocalMatches(r.Match(sentence))
	}
}

// injectClosedClassDT marks a/an/the with DT (English closed-class articles).
// Not a surface invent of which words take a/an — only POS wiring for untagged AnalyzePlain.
func injectClosedClassDT(sentence *languagetool.AnalyzedSentence) {
	if sentence == nil {
		return
	}
	dt := "DT"
	for _, tok := range sentence.GetTokensWithoutWhitespace() {
		if tok == nil {
			continue
		}
		switch strings.ToLower(tok.GetToken()) {
		case "a", "an", "the":
			if tok.HasPosTag("DT") {
				continue
			}
			lem := strings.ToLower(tok.GetToken())
			tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &dt, &lem), "avs_an_dt_inject")
		}
	}
}
