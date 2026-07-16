package en

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

// EnglishRepeatedWordsRule ports org.languagetool.rules.en.EnglishRepeatedWordsRule (surface lemmas).
type EnglishRepeatedWordsRule struct {
	*rules.AbstractRepeatedWordsRule
}

func NewEnglishRepeatedWordsRule(messages map[string]string) *EnglishRepeatedWordsRule {
	base := &rules.AbstractRepeatedWordsRule{
		Messages:     messages,
		ID:           "EN_REPEATEDWORDS",
		Description:  "Suggest synonyms for repeated words.",
		Message:      "This word already appears in one of the immediately preceding sentences. Use a synonym for more variety, unless the repetition is intentional.",
		ShortMsg:     "Style: Repeated word",
		WordsToCheck: loadSynonyms(),
	}
	return &EnglishRepeatedWordsRule{AbstractRepeatedWordsRule: base}
}

func (r *EnglishRepeatedWordsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractRepeatedWordsRule.MatchList(sentences)
}
