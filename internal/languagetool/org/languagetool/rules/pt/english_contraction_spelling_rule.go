package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/english_contractions.txt
var engContrFS embed.FS

var (
	engContrOnce sync.Once
	engContrMap  map[string][]string
)

func loadEnglishContractions() map[string][]string {
	engContrOnce.Do(func() {
		f, err := engContrFS.Open("data/english_contractions.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		engContrMap = m
	})
	return engContrMap
}

// EnglishContractionSpellingRule ports org.languagetool.rules.pt.EnglishContractionSpellingRule.
type EnglishContractionSpellingRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewEnglishContractionSpellingRule(messages map[string]string) *EnglishContractionSpellingRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadEnglishContractions(),
		CaseSensitive: true,
		CheckLemmas:   false,
		ID:            "PT_ENGLISH_CONTRACTION_ORTHOGRAPHY",
		Description:   "Ortografia de contrações inglesas",
		ShortMsg:      "Erro de ortografia inglesa",
		MessageFn: func(tokenStr string, replacements []string) string {
			if len(replacements) == 0 {
				return "Erro de ortografia inglesa"
			}
			return "Caso seja uma contração da língua inglesa, prefira \"" + replacements[0] + "\"."
		},
	}
	// Java: whats → what's
	base.AddExamplePair(
		rules.Wrong("Ele adorava assistir <marker>whats</marker> cooking às sextas-feiras."),
		rules.Fixed("Ele adorava assistir <marker>what's</marker> cooking às sextas-feiras."),
	)
	return &EnglishContractionSpellingRule{AbstractSimpleReplaceRule: base}
}

func (r *EnglishContractionSpellingRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
