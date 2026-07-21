package ca

import (
	"embed"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
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

// SimpleReplaceBalearicRule ports org.languagetool.rules.ca.SimpleReplaceBalearicRule.
// isTokenException is POS/immunize only (Java); no title-case surface invent.
type SimpleReplaceBalearicRule struct {
	*rules.AbstractSimpleReplaceRule
}

func NewSimpleReplaceBalearicRule(messages map[string]string) *SimpleReplaceBalearicRule {
	base := &rules.AbstractSimpleReplaceRule{
		Messages:           messages,
		WrongWords:         loadBalearicWords(),
		CaseSensitive:      false,
		CheckLemmas:        false,
		ID:                 "CA_SIMPLE_REPLACE_BALEARIC",
		LanguageCode:       "ca",
		SubRuleSpecificIDs: true,
		Description:        "Suggeriments per a formes balears: $match",
		ShortMsg:           "Possible error ortogràfic.",
		MessageFn: func(tokenStr string, replacements []string) string {
			return "Possible error ortogràfic (forma verbal vàlida en la varietat balear)."
		},
		// Java isTokenException:
		// hasPosTagStartingWith("NP") || isImmunized || isIgnoredBySpeller
		// || hasPosTag("_english_ignore_") || hasPosTag("_Latin_")
		// (IsIgnoredBySpeller also checked in AbstractSimpleReplaceRule.Match.)
		TokenException: balearicTokenException,
	}
	return &SimpleReplaceBalearicRule{AbstractSimpleReplaceRule: base}
}

func balearicTokenException(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return false
	}
	if token.IsImmunized() || token.IsIgnoredBySpeller() {
		return true
	}
	if token.HasPosTagStartingWith("NP") {
		return true
	}
	if token.HasPosTag("_english_ignore_") || token.HasPosTag("_Latin_") {
		return true
	}
	return false
}

func (r *SimpleReplaceBalearicRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractSimpleReplaceRule.Match(sentence)
}
