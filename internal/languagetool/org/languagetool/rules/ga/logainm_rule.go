package ga

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/placenames.txt
var placenamesFS embed.FS

var (
	placenamesOnce sync.Once
	placenamesMap  map[string][]string
)

func loadPlacenames() map[string][]string {
	placenamesOnce.Do(func() {
		f, err := placenamesFS.Open("data/placenames.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		placenamesMap = m
	})
	return placenamesMap
}

// LogainmRule ports org.languagetool.rules.ga.LogainmRule (English placenames → Irish).
type LogainmRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewLogainmRule(messages map[string]string) *LogainmRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadPlacenames(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "GA_LOGAINM",
		Description:   "Logainm Béarla, m.sh., 'Dublin' in áit 'Baile Átha Cliath'.",
		ShortMsg:      "Logainm Béarla",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Is logainm Béarla é \"" + tokenStr + "\" ar \"" + joinCommaGA(replacements) + "\"."
		},
	}
	return &LogainmRule{AbstractSimpleReplaceRule: base}
}

func (r *LogainmRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
