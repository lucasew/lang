package ca

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/replace_balearic.txt
var balearicFS embed.FS

var (
	balearicOnce  sync.Once
	balearicWords map[string][]string
)

func loadBalearicWords() map[string][]string {
	balearicOnce.Do(func() {
		f, err := balearicFS.Open("data/replace_balearic.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSimpleReplaceWords(f)
		if err != nil {
			panic(err)
		}
		balearicWords = m
	})
	return balearicWords
}

// SimpleReplaceBalearicRule ports org.languagetool.rules.ca.SimpleReplaceBalearicRule
// without POS-tag NP immunization (surface heuristics for proper names).
type SimpleReplaceBalearicRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleReplaceBalearicRule(messages map[string]string) *SimpleReplaceBalearicRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:      messages,
		WrongWords:    loadBalearicWords(),
		CaseSensitive: false,
		CheckLemmas:   false,
		ID:            "CA_SIMPLE_REPLACE_BALEARIC",
		Description:   "Suggeriments per a formes balears: $match",
		ShortMsg:      "Possible error ortogràfic.",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Possible error ortogràfic (forma verbal vàlida en la varietat balear)."
		},
		// Stand-in for NP / multiword proper-name exceptions (Prosper, Index, …).
		TokenException: func(token *languagetool.AnalyzedTokenReadings) bool {
			if token.IsImmunized() {
				return true
			}
			t := token.GetToken()
			// Title case (not ALL CAPS): treat as possible proper name / Latin title word.
			// Java uses hasPosTagStartingWith("NP"); without a tagger this is surface-only.
			if tools.IsCapitalizedWord(t) && !tools.IsAllUppercase(t) {
				switch strings.ToLower(t) {
				case "prosper", "index":
					return true
				}
			}
			return false
		},
	}
	return &SimpleReplaceBalearicRule{AbstractSimpleReplaceRule: base}
}

func (r *SimpleReplaceBalearicRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
