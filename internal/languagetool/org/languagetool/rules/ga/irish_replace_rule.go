package ga

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

// IrishReplaceRule ports org.languagetool.rules.ga.IrishReplaceRule.
type IrishReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewIrishReplaceRule(messages map[string]string) *IrishReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadReplace(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "GA_REPLACE",
		Description:   "Litriú mícheart, m.sh., \"agsu\" in áit \"agus\"",
		ShortMsg:      "Litriú mícheart",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Litriú mícheart. An bhfuil «" + joinCommaGA(replacements) + "» i gceist agat?"
		},
	}
	// Java: bhúr → bhur
	base.AddExamplePair(
		rules.Wrong("Níl beann agam oraibh, ar <marker>bhúr</marker> gcuid cainte."),
		rules.Fixed("Níl beann agam oraibh, ar <marker>bhur</marker> gcuid cainte."),
	)
	return &IrishReplaceRule{AbstractSimpleReplaceRule: base}
}

func joinCommaGA(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}

func (r *IrishReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
