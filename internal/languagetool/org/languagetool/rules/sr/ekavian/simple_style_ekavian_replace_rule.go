package ekavian

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

// SimpleStyleEkavianReplaceRule ports org.languagetool.rules.sr.ekavian.SimpleStyleEkavianReplaceRule.
type SimpleStyleEkavianReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleStyleEkavianReplaceRule(messages map[string]string) *SimpleStyleEkavianReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadStyle(),
		CaseSensitive: false,
		CheckLemmas:   true, // Java default checkLemmas true
		ID:            "SR_EKAVIAN_SIMPLE_STYLE_REPLACE_RULE",
		Description:   "Провера стилски лоших речи или израза",
		ShortMsg:      "Стил",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Уместо «" + tokenStr + "» боље је рећи: " + joinComma(replacements) + "."
		},
	}
	return &SimpleStyleEkavianReplaceRule{AbstractSimpleReplaceRule: base}
}

func joinComma(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}

func (r *SimpleStyleEkavianReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
