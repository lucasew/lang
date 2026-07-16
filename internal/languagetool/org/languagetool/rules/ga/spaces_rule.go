package ga

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/spaces.txt
var spacesFS embed.FS

var (
	spacesOnce sync.Once
	spacesMap  map[string][]string
)

func loadSpaces() map[string][]string {
	spacesOnce.Do(func() {
		f, err := spacesFS.Open("data/spaces.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		spacesMap = m
	})
	return spacesMap
}

// SpacesRule ports org.languagetool.rules.ga.SpacesRule (missing spaces between words).
type SpacesRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSpacesRule(messages map[string]string) *SpacesRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadSpaces(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "GA_SPASANNA",
		Description:   "Spás ar iarraidh",
		ShortMsg:      "Spás ar iarraidh",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Spás ar iarraidh: \"" + joinCommaGA(replacements) + "\"."
		},
	}
	return &SpacesRule{AbstractSimpleReplaceRule: base}
}

func (r *SpacesRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
