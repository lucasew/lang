package nl

import (
	"bufio"
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/preferredwords.csv
var preferredFS embed.FS

type preferredPair struct {
	oldWord, newWord string
}

var (
	preferredOnce sync.Once
	preferredList []preferredPair
)

func loadPreferredWords() []preferredPair {
	preferredOnce.Do(func() {
		f, err := preferredFS.Open("data/preferredwords.csv")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || line[0] == '#' {
				continue
			}
			parts := strings.Split(line, ";")
			if len(parts) != 2 {
				continue
			}
			preferredList = append(preferredList, preferredPair{
				oldWord: parts[0],
				newWord: parts[1],
			})
		}
	})
	return preferredList
}

// PreferredWordRule ports org.languagetool.rules.nl.PreferredWordRule.
type PreferredWordRule struct {
	Messages map[string]string
}

func NewPreferredWordRule(messages map[string]string) *PreferredWordRule {
	return &PreferredWordRule{Messages: messages}
}

func (r *PreferredWordRule) GetID() string { return "NL_PREFERRED_WORD_RULE" }

func (r *PreferredWordRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var out []*rules.RuleMatch
	for _, pair := range loadPreferredWords() {
		// Case-sensitive token match of oldWord (may be multi-word).
		parts := strings.Split(pair.oldWord, " ")
		tokens := sentence.GetTokensWithoutWhitespace()
		for i := 1; i+len(parts)-1 < len(tokens); i++ {
			ok := true
			for j, p := range parts {
				if tokens[i+j].GetToken() != p {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			from := tokens[i].GetStartPos()
			to := tokens[i+len(parts)-1].GetEndPos()
			matched := sentence.GetText()
			// rebuild matched text from tokens
			var b strings.Builder
			for j := range parts {
				if j > 0 {
					b.WriteByte(' ')
				}
				b.WriteString(tokens[i+j].GetToken())
			}
			matchedText := b.String()
			suggestion := strings.Replace(matchedText, pair.oldWord, pair.newWord, 1)
			if suggestion == matchedText {
				continue
			}
			msg := "Voor dit woord is een gebruikelijker alternatief."
			rm := rules.NewRuleMatch(r, sentence, from, to, msg)
			rm.ShortMessage = "Gebruikelijker woord"
			rm.SetSuggestedReplacement(suggestion)
			out = append(out, rm)
			_ = matched
		}
	}
	return out
}
