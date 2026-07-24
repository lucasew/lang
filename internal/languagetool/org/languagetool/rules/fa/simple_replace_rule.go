package fa

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt
var replaceFS embed.FS

var (
	replaceOnce sync.Once
	replaceMap  map[string][]string
)

func loadReplace() map[string][]string {
	replaceOnce.Do(func() {
		f, err := replaceFS.Open("data/replace.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		replaceMap = m
	})
	return replaceMap
}

// SimpleReplaceRule ports org.languagetool.rules.fa.SimpleReplaceRule.
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadReplace(),
		CaseSensitive: false,
		CheckLemmas:   true, // Java default isCheckLemmas true
		ID:            "FA_SIMPLE_REPLACE",
		Description:   "اشتباه محتمل املائی",
		ShortMsg:      "اشتباه محتمل املائی",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "اشتباه محتمل املائی پیداشده: " + joinPersian(replacements) + "."
		},
	}
	// Java: حاظر → حاضر
	base.AddExamplePair(
		rules.Wrong("وی <marker>حاظر</marker> به همکاری شد."),
		rules.Fixed("وی <marker>حاضر</marker> به همکاری شد."),
	)
	return &SimpleReplaceRule{AbstractSimpleReplaceRule: base}
}

func joinPersian(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += "، "
		}
		out += s
	}
	return out
}

func (r *SimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
