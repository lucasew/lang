package jekavian

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace-style.txt
var styleFS embed.FS

var (
	styleOnce sync.Once
	styleMap  map[string][]string
)

func loadStyle() map[string][]string {
	styleOnce.Do(func() {
		f, err := styleFS.Open("data/replace-style.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		styleMap = m
	})
	return styleMap
}

// SimpleStyleJekavianReplaceRule ports org.languagetool.rules.sr.jekavian.SimpleStyleJekavianReplaceRule.
type SimpleStyleJekavianReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleStyleJekavianReplaceRule(messages map[string]string) *SimpleStyleJekavianReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadStyle(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "SR_JEKAVIAN_SIMPLE_STYLE_REPLACE_RULE",
		Description:   "Провера стилски лоших ријечи или израза",
		ShortMsg:      "Стилски лоша ријеч тј. израз",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Умјесто израза „" + tokenStr + "“ било би боље да користите: " + joinComma(replacements) + "."
		},
	}
	return &SimpleStyleJekavianReplaceRule{AbstractSimpleReplaceRule: base}
}

func (r *SimpleStyleJekavianReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
