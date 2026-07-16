package pt

import (
	"embed"
	"strings"
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
type PortugueseOrthographyReplaceRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewPortugueseOrthographyReplaceRule(messages map[string]string) *PortugueseOrthographyReplaceRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadOrthographyWords(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "PT_SIMPLE_REPLACE_ORTHOGRAPHY",
		Description:   "Possible spelling mistake found.",
		ShortMsg:      "Spelling mistake",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Possible spelling mistake found."
		},
	}
	return &PortugueseOrthographyReplaceRule{AbstractSimpleReplaceRule: base}
}

func (r *PortugueseOrthographyReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	matches := r.AbstractSimpleReplaceRule.Match(sentence)
	if len(matches) == 0 {
		return matches
	}
	// Surface stand-in for multiword tagger immunization of Italian "sotto voce".
	tokens := sentence.GetTokensWithoutWhitespace()
	prevByStart := map[int]string{}
	prev := ""
	for _, tok := range tokens {
		if tok.IsSentenceStart() {
			continue
		}
		prevByStart[tok.GetStartPos()] = prev
		prev = tok.GetToken()
	}
	out := matches[:0]
	for _, m := range matches {
		from := m.GetFromPos()
		if strings.EqualFold(prevByStart[from], "sotto") {
			skip := false
			for _, tok := range tokens {
				if tok.GetStartPos() == from && strings.EqualFold(tok.GetToken(), "voce") {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}
		out = append(out, m)
	}
	return out
}
