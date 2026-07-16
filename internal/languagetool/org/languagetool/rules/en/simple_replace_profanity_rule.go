package en

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace_profanity.txt
var profanityFS embed.FS

var (
	profanityOnce sync.Once
	profanityBase *rules.AbstractSimpleReplaceRule2
)

func loadProfanity() *rules.AbstractSimpleReplaceRule2 {
	profanityOnce.Do(func() {
		f, err := profanityFS.Open("data/replace_profanity.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base := &rules.AbstractSimpleReplaceRule2{
			ID:              "PROFANITY",
			Description:     "Profanity",
			ShortMsg:        "Profanity",
			MessageTemplate: "This expression can be considered offensive.",
			CaseSens:        rules.CaseInsensitive,
			LanguageCode:    "en",
		}
		if err := base.LoadSimpleReplaceRule2Data(f, "/en/replace_profanity.txt"); err != nil {
			panic(err)
		}
		profanityBase = base
	})
	return profanityBase
}

// SimpleReplaceProfanityRule ports org.languagetool.rules.en.SimpleReplaceProfanityRule
// (suggestion-less dictionary matches).
type SimpleReplaceProfanityRule struct {
	*rules.AbstractSimpleReplaceRule2
}

func NewSimpleReplaceProfanityRule(messages map[string]string) *SimpleReplaceProfanityRule {
	base := loadProfanity()
	r := *base
	r.Messages = messages
	return &SimpleReplaceProfanityRule{AbstractSimpleReplaceRule2: &r}
}

func (r *SimpleReplaceProfanityRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule2.Match(sentence)
}
