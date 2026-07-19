package ca

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/replace.txt data/replace_custom.txt
var replaceFS embed.FS

var (
	replaceOnce  sync.Once
	replaceWords map[string][]string

	// Java SimpleReplaceRule.exceptions — skip ConvertToGenderAndNumberFilter.
	simpleReplaceExceptions = map[string]bool{
		"perquè":  true,
		"què":     true,
		"per què": true,
	}

	simpleReplaceGenderArgs = map[string]string{"lemmaSelect": "[NA].*"}
)

func loadReplaceWords() map[string][]string {
	replaceOnce.Do(func() {
		m := map[string][]string{}
		for _, name := range []string{"data/replace.txt", "data/replace_custom.txt"} {
			f, err := replaceFS.Open(name)
			if err != nil {
				continue
			}
			part, err := rules.LoadSimpleReplaceWords(f)
			f.Close()
			if err != nil {
				panic(err)
			}
			for k, v := range part {
				m[k] = v
			}
		}
		replaceWords = m
	})
	return replaceWords
}

// SimpleReplaceRule ports org.languagetool.rules.ca.SimpleReplaceRule.
// Java: setIgnoreTaggedWords() + setCheckLemmas(false). Match post-pass uses
// ConvertToGenderAndNumberFilter except for perquè/què/per què suggestions.
// Without Filter.Tag, surface replacements are kept (gender/number incomplete).
type SimpleReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
	// Filter optional; when Tag set, runs ConvertToGenderAndNumberFilter.
	Filter *ConvertToGenderAndNumberFilter
}

func NewSimpleReplaceRule(messages map[string]string) *SimpleReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:          messages,
		WrongWords:        loadReplaceWords(),
		CaseSensitive:     false,
		CheckLemmas:       false,
		IgnoreTaggedWords: true, // Java setIgnoreTaggedWords()
		ID:                "CA_SIMPLE_REPLACE_SIMPLE",
		Description:       "Paraula incorrecta: $match",
		ShortMsg:          "Paraula incorrecta",
		MessageFn: func(tokenStr string, replacements []string) string {
			if len(replacements) > 0 {
				return "¿Volíeu dir «" + replacements[0] + "»?"
			}
			return "Paraula incorrecta"
		},
		Category: rules.CatTypos.GetCategory(messages),
	}
	return &SimpleReplaceRule{AbstractSimpleReplaceRule: base}
}

// Match ports SimpleReplaceRule.match (super + gender/number filter + exceptions).
func (r *SimpleReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil {
		return nil
	}
	potential := r.AbstractSimpleReplaceRule.Match(sentence)
	filter := r.Filter
	// Without Tag the filter returns nil for suggestion-seed path — keep surface.
	if filter == nil || filter.Tag == nil {
		return potential
	}
	var out []*rules.RuleMatch
	for _, m := range potential {
		if m == nil {
			continue
		}
		if simpleReplaceHasException(m) {
			out = append(out, m)
			continue
		}
		final := filter.AcceptRuleMatch(m, simpleReplaceGenderArgs, 0, nil, nil)
		if final != nil {
			out = append(out, final)
		}
	}
	return out
}

func simpleReplaceHasException(m *rules.RuleMatch) bool {
	for _, s := range m.GetSuggestedReplacements() {
		if simpleReplaceExceptions[strings.ToLower(s)] {
			return true
		}
	}
	return false
}
