package rules

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ContextWords ports WrongWordInContextRule.ContextWords.
type ContextWords struct {
	Matches      [2]string
	Explanations [2]string
	Words        [2]*regexp.Regexp
	Contexts     [2]*regexp.Regexp
}

// WrongWordInContextRule ports org.languagetool.rules.WrongWordInContextRule.
type WrongWordInContextRule struct {
	Messages           map[string]string
	ID                 string
	Description        string
	MessageString      string // $SUGGESTION, $WRONGWORD
	LongMessageString  string // + $EXPLANATION_*
	ShortMessageString string
	LanguageCode       string
	MatchLemmas        bool
	Entries            []ContextWords
}

// LoadWrongWordInContext loads tab-separated context confusion entries.
// Format: word1 word2 match1 match2 context1 context2 [explanation1 explanation2]
func LoadWrongWordInContext(r io.Reader) ([]ContextWords, error) {
	var set []ContextWords
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" || line[0] == '#' {
			continue
		}
		column := strings.Split(line, "\t")
		if len(column) < 6 {
			continue
		}
		var cw ContextWords
		var err error
		cw.Words[0], err = compileBoundary(column[0])
		if err != nil {
			return nil, err
		}
		cw.Words[1], err = compileBoundary(column[1])
		if err != nil {
			return nil, err
		}
		cw.Matches[0] = column[2]
		cw.Matches[1] = column[3]
		cw.Contexts[0], err = compileBoundary(column[4])
		if err != nil {
			return nil, err
		}
		cw.Contexts[1], err = compileBoundary(column[5])
		if err != nil {
			return nil, err
		}
		if len(column) > 6 {
			cw.Explanations[0] = column[6]
			if len(column) > 7 {
				cw.Explanations[1] = column[7]
			}
		}
		set = append(set, cw)
	}
	return set, sc.Err()
}

func compileBoundary(str string) (*regexp.Regexp, error) {
	ignoreCase := ""
	if strings.HasPrefix(str, "(?i)") {
		str = str[4:]
		ignoreCase = "(?i)"
	}
	// RE2: (?i) must be at start; word boundary around group
	pat := ignoreCase + `\b(?:` + str + `)\b`
	return regexp.Compile(pat)
}

func (r *WrongWordInContextRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "WRONG_WORD_IN_CONTEXT"
}

// Match ports WrongWordInContextRule.match.
func (r *WrongWordInContextRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	for _, contextWords := range r.Entries {
		matchedWord := [2]bool{}
		matchedPos := [2]int{-1, -1}
		matchedEnd := [2]int{-1, -1}
		matchedTok := [2]string{}
		for k := 1; k < len(tokens) && (!matchedWord[0] || !matchedWord[1]); k++ {
			t, endK := joinNoSpaceTokens(tokens, k)
			if !matchedWord[0] && contextWords.Words[0].MatchString(t) {
				matchedWord[0] = true
				matchedPos[0] = k
				matchedEnd[0] = endK
				matchedTok[0] = t
			}
			if !matchedWord[1] && contextWords.Words[1].MatchString(t) {
				matchedWord[1] = true
				matchedPos[1] = k
				matchedEnd[1] = endK
				matchedTok[1] = t
			}
		}
		foundWord, notFoundWord := -1, -1
		startPos, endPos := 0, 0
		matchedToken := ""
		if matchedWord[0] && !matchedWord[1] {
			foundWord, notFoundWord = 0, 1
			startPos = tokens[matchedPos[0]].GetStartPos()
			endPos = tokens[matchedEnd[0]].GetEndPos()
			matchedToken = matchedTok[0]
		} else if matchedWord[1] && !matchedWord[0] {
			foundWord, notFoundWord = 1, 0
			startPos = tokens[matchedPos[1]].GetStartPos()
			endPos = tokens[matchedEnd[1]].GetEndPos()
			matchedToken = matchedTok[1]
		}
		if foundWord == -1 {
			continue
		}
		matchedContext := [2]bool{}
		for k := 1; k < len(tokens) && (!matchedContext[foundWord] || !matchedContext[notFoundWord]); k++ {
			if r.MatchLemmas {
				// Try real lemmas, then surface + light stem guesses (no tagger).
				cands := lemmaCandidates(tokens[k])
				for _, cand := range cands {
					if !matchedContext[foundWord] {
						matchedContext[foundWord] = contextWords.Contexts[foundWord].MatchString(cand)
					}
					if !matchedContext[notFoundWord] {
						matchedContext[notFoundWord] = contextWords.Contexts[notFoundWord].MatchString(cand)
					}
				}
			} else {
				token := tokens[k].GetToken()
				if !matchedContext[foundWord] {
					matchedContext[foundWord] = contextWords.Contexts[foundWord].MatchString(token)
				}
				if !matchedContext[notFoundWord] {
					matchedContext[notFoundWord] = contextWords.Contexts[notFoundWord].MatchString(token)
				}
			}
		}
		if matchedContext[notFoundWord] && !matchedContext[foundWord] {
			originalStr := contextWords.Matches[foundWord]
			replacementStr := contextWords.Matches[notFoundWord]
			// Java: matchedToken.replaceFirst("(?i)" + originalStr, replacementStr)
			re, err := regexp.Compile("(?i)" + originalStr)
			if err != nil {
				continue
			}
			loc := re.FindStringIndex(matchedToken)
			if loc == nil {
				continue
			}
			tmp := matchedToken[:loc[0]] + replacementStr + matchedToken[loc[1]:]
			repl := tools.PreserveCase(tmp, matchedToken)
			msg := r.buildMessage(matchedToken, repl, contextWords.Explanations[notFoundWord], contextWords.Explanations[foundWord])
			rm := NewRuleMatch(r, sentence, startPos, endPos, msg)
			rm.ShortMessage = r.ShortMessageString
			rm.SetSuggestedReplacement(repl)
			ruleMatches = append(ruleMatches, rm)
		}
	}
	return ruleMatches
}

func (r *WrongWordInContextRule) buildMessage(wrongWord, suggestion, explanationSuggestion, explanationWrongWord string) string {
	if explanationSuggestion == "" || explanationWrongWord == "" {
		msg := r.MessageString
		if msg == "" {
			msg = "Possibly confused word: Did you mean <suggestion>$SUGGESTION</suggestion> instead of '$WRONGWORD'?"
		}
		msg = strings.Replace(msg, "$SUGGESTION", suggestion, 1)
		msg = strings.Replace(msg, "$WRONGWORD", wrongWord, 1)
		return msg
	}
	msg := r.LongMessageString
	if msg == "" {
		msg = "Possibly confused word: Did you mean <suggestion>$SUGGESTION</suggestion> (= $EXPLANATION_SUGGESTION) instead of '$WRONGWORD' (= $EXPLANATION_WRONGWORD)?"
	}
	msg = strings.Replace(msg, "$SUGGESTION", suggestion, 1)
	msg = strings.Replace(msg, "$WRONGWORD", wrongWord, 1)
	msg = strings.Replace(msg, "$EXPLANATION_SUGGESTION", explanationSuggestion, 1)
	msg = strings.Replace(msg, "$EXPLANATION_WRONGWORD", explanationWrongWord, 1)
	return msg
}

func lemmaCandidates(tok *languagetool.AnalyzedTokenReadings) []string {
	seen := map[string]bool{}
	var out []string
	add := func(s string) {
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		out = append(out, s)
	}
	add(tok.GetToken())
	for _, rd := range tok.GetReadings() {
		if rd.GetLemma() != nil {
			add(*rd.GetLemma())
		}
	}
	// Light Romance/Germanic stem guesses when untagged.
	t := strings.ToLower(tok.GetToken())
	add(t)
	if len(t) > 3 && strings.HasSuffix(t, "s") {
		add(t[:len(t)-1])
	}
	if len(t) > 4 && strings.HasSuffix(t, "es") {
		add(t[:len(t)-2])
	}
	// Catalan/Spanish imperfect: seguia → seguir
	if len(t) > 4 && strings.HasSuffix(t, "ia") {
		stem := t[:len(t)-2]
		add(stem + "ir")
		add(stem + "ar")
		add(stem + "er")
		add(stem)
	}
	if len(t) > 4 && strings.HasSuffix(t, "ava") {
		stem := t[:len(t)-3]
		add(stem + "ar")
		add(stem)
	}
	return out
}

// joinNoSpaceTokens joins middle-dot compounds (Catalan pal·li → pal + · + li)
// without swallowing sentence punctuation like trailing periods.
func joinNoSpaceTokens(tokens []*languagetool.AnalyzedTokenReadings, k int) (string, int) {
	if k < 0 || k >= len(tokens) {
		return "", k
	}
	t := tokens[k].GetToken()
	end := k
	for end+1 < len(tokens) && !tokens[end+1].IsWhitespaceBefore() {
		next := tokens[end+1].GetToken()
		if next == "·" || next == "‧" || next == "•" {
			end++
			t += next
			continue
		}
		if strings.HasSuffix(t, "·") || strings.HasSuffix(t, "‧") {
			end++
			t += next
			continue
		}
		break
	}
	return t, end
}
