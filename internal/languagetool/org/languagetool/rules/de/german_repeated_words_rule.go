package de

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/synonyms.txt
var synonymsFS embed.FS

var (
	synOnce sync.Once
	synMap  map[string]*rules.SynonymsData
)

func loadSynonyms() map[string]*rules.SynonymsData {
	synOnce.Do(func() {
		f, err := synonymsFS.Open("data/synonyms.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSynonymsWords(f)
		if err != nil {
			panic(err)
		}
		synMap = m
	})
	return synMap
}

// GermanRepeatedWordsRule ports org.languagetool.rules.de.GermanRepeatedWordsRule.
type GermanRepeatedWordsRule struct {
	*rules.AbstractRepeatedWordsRule
}

func NewGermanRepeatedWordsRule(messages map[string]string) *GermanRepeatedWordsRule {
	// Java getDescription / getMessage / getShortMessage (not invent English).
	// Java AbstractRepeatedWordsRule: REPETITIONS_STYLE + ITS Style (no picky tag on DE).
	base := &rules.AbstractRepeatedWordsRule{
		ID:          "DE_REPEATEDWORDS",
		Description: "Synonyme für wiederholte Wörter.",
		Message: "Dieses Wort kommt in einem nahe gelegenen vorherigen Satz bereits vor. " +
			"Verwenden Sie ein Synonym, um Ihren Text abwechslungsreicher zu gestalten, außer die Wiederholung ist beabsichtigt.",
		ShortMsg:     "Stil: Wortwiederholung",
		WordsToCheck: loadSynonyms(),
		LanguageCode: "de",
	}
	rules.InitRepeatedWordsMeta(base, messages)
	// Java GermanRepeatedWordsRule.isException:
	// isAllUppercase || (isCapitalized && !sentStart) || hasPosTagStartingWith("EIG:")
	base.IsException = germanRepeatedWordsIsException
	// Java getSynthesizer() → GermanSynthesizer.INSTANCE (fail-closed when dict missing).
	base.SynthesizeRE = germanRepeatedWordsSynthesizeRE
	return &GermanRepeatedWordsRule{AbstractRepeatedWordsRule: base}
}

// germanRepeatedWordsSynthesizeRE ports GermanSynthesizer.INSTANCE.synthesize(token, postag, true).
func germanRepeatedWordsSynthesizeRE(token *languagetool.AnalyzedToken, posTag string) []string {
	if token == nil {
		return nil
	}
	if gs := openDiscoveredGermanSynthesizer(); gs != nil {
		forms, err := gs.SynthesizeRE(token, posTag, true)
		if err != nil || len(forms) == 0 {
			return nil
		}
		return forms
	}
	if base := openDiscoveredGermanSynthBase(); base != nil {
		forms, err := base.SynthesizeRE(token, posTag, true)
		if err != nil || len(forms) == 0 {
			return nil
		}
		return forms
	}
	return nil
}

// germanRepeatedWordsIsException ports GermanRepeatedWordsRule.isException.
func germanRepeatedWordsIsException(
	tokens []*languagetool.AnalyzedTokenReadings,
	i int,
	sentStart, isCapitalized, isAllUppercase bool,
) bool {
	if isAllUppercase || (isCapitalized && !sentStart) {
		return true
	}
	if i >= 0 && i < len(tokens) && tokens[i] != nil && tokens[i].HasPosTagStartingWith("EIG:") {
		return true
	}
	return false
}

func (r *GermanRepeatedWordsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractRepeatedWordsRule.MatchList(sentences)
}

// MinToCheckParagraph ports AbstractRepeatedWordsRule.minToCheckParagraph (Java returns 1).
func (r *GermanRepeatedWordsRule) MinToCheckParagraph() int { return 1 }
