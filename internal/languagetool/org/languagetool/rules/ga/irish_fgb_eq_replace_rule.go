package ga

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace-fgb-eq.txt
var fgbEqFS embed.FS

var (
	fgbEqOnce sync.Once
	fgbEqMap  map[string][]string
)

func loadFGBEq() map[string][]string {
	fgbEqOnce.Do(func() {
		f, err := fgbEqFS.Open("data/replace-fgb-eq.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		fgbEqMap = m
	})
	return fgbEqMap
}

// IrishFGBEqReplaceRule ports org.languagetool.rules.ga.IrishFGBEqReplaceRule.
type IrishFGBEqReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewIrishFGBEqReplaceRule(messages map[string]string) *IrishFGBEqReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadFGBEq(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "GA_FGB_EQ_REPLACE",
		Description:   "Ceannfhocal FGB neamhchoitianta, m.sh., \"urlamh\" in áit \"ullamh\"",
		ShortMsg:      "Neamhchoitianta",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Focal ceart ach tá \"" + joinCommaGA(replacements) + "\" níos coitianta."
		},
	}
	// Java: urlamh → ullamh
	base.AddExamplePair(
		rules.Wrong("An bhfuil tú <marker>urlamh</marker>?"),
		rules.Fixed("An bhfuil tú <marker>ullamh</marker>?"),
	)
	return &IrishFGBEqReplaceRule{AbstractSimpleReplaceRule: base}
}

func (r *IrishFGBEqReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
