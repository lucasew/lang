package rules

import (
	"bufio"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AbstractSpecificCaseRule ports org.languagetool.rules.AbstractSpecificCaseRule.
type AbstractSpecificCaseRule struct {
	Messages                   map[string]string
	LcToProper                 map[string]string
	MaxPhraseLen               int
	ID                         string
	Description                string
	InitialCapitalMessage      string
	OtherCapitalizationMessage string
	ShortMsg                   string
}

// LoadSpecificCasePhrases loads phrase list: each non-comment line is the proper spelling.
func LoadSpecificCasePhrases(r io.Reader) (map[string]string, int, error) {
	m := make(map[string]string)
	maxLen := 0
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) > maxLen {
			maxLen = len(parts)
		}
		m[strings.ToLower(line)] = line
	}
	return m, maxLen, sc.Err()
}

func (r *AbstractSpecificCaseRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "SPECIFIC_CASE"
}

func (r *AbstractSpecificCaseRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var matches []*RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	for i := 0; i < len(tokens); i++ {
		var l []string
		j := 0
		for len(l) < r.MaxPhraseLen && i+j < len(tokens) {
			l = append(l, tokens[i+j].GetToken())
			j++
			phrase := strings.Join(l, " ")
			lcPhrase := strings.ToLower(phrase)
			properSpelling, ok := r.LcToProper[lcPhrase]
			if !ok || tools.IsAllUppercase(phrase) || phrase == properSpelling {
				continue
			}
			// sentence start: avoid suggesting lowercase-start phrases
			if i > 0 && tokens[i-1].IsSentenceStart() && !tools.StartsWithUppercase(properSpelling) {
				continue
			}
			// Also: first real token after SENT_START is i==1 with tokens[0] SENT_START
			if i == 1 && tokens[0].IsSentenceStart() && !tools.StartsWithUppercase(properSpelling) {
				continue
			}
			msg := r.OtherCapitalizationMessage
			if msg == "" {
				msg = "The particular expression should follow the suggested capitalization."
			}
			if allWordsUppercase(properSpelling) {
				msg = r.InitialCapitalMessage
				if msg == "" {
					msg = "The initials of the particular phrase must be capitals."
				}
			}
			from := tokens[i].GetStartPos()
			to := tokens[i+j-1].GetEndPos()
			rm := NewRuleMatch(r, sentence, from, to, msg)
			short := r.ShortMsg
			if short == "" {
				short = "Special capitalization"
			}
			rm.ShortMessage = short
			rm.SetSuggestedReplacement(properSpelling)
			matches = append(matches, rm)
		}
	}
	return matches
}

func allWordsUppercase(s string) bool {
	for _, w := range strings.Split(s, " ") {
		if !tools.StartsWithUppercase(w) {
			return false
		}
	}
	return true
}
