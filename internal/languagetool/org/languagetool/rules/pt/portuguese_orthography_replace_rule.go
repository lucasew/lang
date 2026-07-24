package pt

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_orthography.txt
var orthographyFS embed.FS

var (
	orthographyOnce  sync.Once
	orthographyWords map[string][]string
)

func loadOrthographyWords() map[string][]string {
	orthographyOnce.Do(func() {
		f, err := orthographyFS.Open("data/replace_orthography.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		orthographyWords = m
	})
	return orthographyWords
}

// PortugueseOrthographyReplaceRule ports org.languagetool.rules.pt.PortugueseOrthographyReplaceRule.
// Multiword immunization (e.g. Italian "sotto voce") is analysis/disambiguation owned;
// without immunized tokens, "voce" still matches (fail closed — no surface invent skip).
type PortugueseOrthographyReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewPortugueseOrthographyReplaceRule(messages map[string]string) *PortugueseOrthographyReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadOrthographyWords(),
		CaseSensitive: false,
		CheckLemmas:   true, // Java default checkLemmas true
		ID:            "PT_SIMPLE_REPLACE_ORTHOGRAPHY",
		LanguageCode:         "pt",
		SubRuleSpecificIDs:   true,
		Description:   "Possible spelling mistake found.",
		ShortMsg:      "Spelling mistake",
		Category:      rules.CatTypos.GetCategory(messages),
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Possible spelling mistake found."
		},
	}
	return &PortugueseOrthographyReplaceRule{AbstractSimpleReplaceRule: base}
}

func (r *PortugueseOrthographyReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
