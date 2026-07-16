package rules

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AbstractWordCoherencyRule ports org.languagetool.rules.AbstractWordCoherencyRule.
type AbstractWordCoherencyRule struct {
	Messages    map[string]string
	ID          string
	Description string
	WordMap     map[string]map[string]struct{}
	// ToBase maps surface form → uninflected file form (lemma stand-in).
	ToBase map[string]string
	// MessageFn(word1, word2) — word1 is the later variant, word2 the established one.
	MessageFn         func(word1, word2 string) string
	ShortMsg          string
	CreateReplacement func(marked, token, otherSpelling string, tmpToken *languagetool.AnalyzedTokenReadings) string
}

func (r *AbstractWordCoherencyRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "WORD_COHERENCY"
}

// Match ports AbstractWordCoherencyRule.match over sentences.
func (r *AbstractWordCoherencyRule) Match(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	shouldNotAppearWord := make(map[string]string) // later form → established base form
	pos := 0
	for _, sentence := range sentences {
		for _, tmpToken := range sentence.GetTokensWithoutWhitespace() {
			candidates := coherencyCandidates(tmpToken)
			for _, cand := range candidates {
				key := strings.ToLower(cand)
				fromPos := pos + tmpToken.GetStartPos()
				toPos := pos + tmpToken.GetEndPos()
				if other, ok := shouldNotAppearWord[key]; ok {
					msg := r.message(cand, other)
					ruleMatch := NewRuleMatch(r, sentence, fromPos, toPos, msg)
					ruleMatch.ShortMessage = r.ShortMsg
					marked := tmpToken.GetToken()
					// Replace using current base (lemma stand-in) → established base (Java).
					curBase := key
					if r.ToBase != nil {
						if b, ok := r.ToBase[key]; ok {
							curBase = b
						}
					}
					replacement := r.createReplacement(marked, curBase, other, tmpToken)
					if tools.StartsWithUppercase(tmpToken.GetToken()) {
						replacement = tools.UppercaseFirstChar(replacement)
					}
					if !strings.EqualFold(marked, replacement) {
						ruleMatch.SetSuggestedReplacement(replacement)
						ruleMatches = append(ruleMatches, ruleMatch)
					}
					break
				} else if alts, ok := r.WordMap[key]; ok {
					established := key
					if r.ToBase != nil {
						if b, ok := r.ToBase[key]; ok {
							established = b
						}
					}
					for shouldNotAppear := range alts {
						shouldNotAppearWord[shouldNotAppear] = established
					}
				}
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}

func (r *AbstractWordCoherencyRule) message(word1, word2 string) string {
	if r.MessageFn != nil {
		return r.MessageFn(word1, word2)
	}
	return "Do not mix variants of the same word ('" + word1 + "' and '" + word2 + "') within a single text."
}

func (r *AbstractWordCoherencyRule) createReplacement(marked, token, otherSpelling string, tmpToken *languagetool.AnalyzedTokenReadings) string {
	if r.CreateReplacement != nil {
		return r.CreateReplacement(marked, token, otherSpelling, tmpToken)
	}
	re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(token))
	if err != nil {
		return otherSpelling
	}
	loc := re.FindStringIndex(marked)
	if loc == nil {
		// token base not a substring of marked (e.g. reelected vs reelect) — use other directly
		return otherSpelling
	}
	return marked[:loc[0]] + otherSpelling + marked[loc[1]:]
}

func coherencyCandidates(tmpToken *languagetool.AnalyzedTokenReadings) []string {
	var out []string
	seen := map[string]bool{}
	readings := tmpToken.GetReadings()
	if len(readings) == 0 {
		return []string{tmpToken.GetToken()}
	}
	for _, rd := range readings {
		tok := tmpToken.GetToken()
		if rd.GetLemma() != nil && *rd.GetLemma() != "" {
			tok = *rd.GetLemma()
		}
		if !seen[tok] {
			seen[tok] = true
			out = append(out, tok)
		}
	}
	if len(out) == 0 {
		out = []string{tmpToken.GetToken()}
	}
	return out
}
