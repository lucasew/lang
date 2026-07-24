package rules

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const wrongWordMinPrefixLength = 3

// ContextWords ports WrongWordInContextRule.ContextWords.
type ContextWords struct {
	Matches      [2]string
	Explanations [2]string
	Words        [2]*regexp.Regexp
	Contexts     [2]*regexp.Regexp
	// wordPrefixes ports ContextWords.wordPrefixes (couldMatchSentence filter).
	wordPrefixes [2]string
}

// WrongWordInContextRule ports org.languagetool.rules.WrongWordInContextRule.
// Java ctor: setCategory(new Category(CONFUSED_WORDS, getCategoryString())),
// setLocQualityIssueType(Misspelling). Default getCategoryString() is messages
// "category_misc"; language subclasses override the display name.
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
	// Category ports Rule.category (Java CONFUSED_WORDS + getCategoryString name).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Misspelling).
	IssueType ITSIssueType
	// Premium ports Rule.isPremium (SpecificIdRule).
	Premium bool
	// Tags ports Rule.getTags (SpecificIdRule).
	Tags []Tag
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
}

// InitWrongWordInContextMeta ports WrongWordInContextRule constructor meta.
// categoryString is getCategoryString(); empty uses default (messages "category_misc").
func InitWrongWordInContextMeta(r *WrongWordInContextRule, messages map[string]string, categoryString string) {
	if r == nil {
		return
	}
	name := categoryString
	if name == "" {
		name = "MISC"
		if messages != nil {
			if s := messages["category_misc"]; s != "" {
				name = s
			}
		}
	}
	r.Category = NewCategory(CategoryConfusedWords, name)
	r.IssueType = ITSMisspelling
}

// GetCategory ports Rule.getCategory.
func (r *WrongWordInContextRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// GetLocQualityIssueType ports Rule.getLocQualityIssueType.
func (r *WrongWordInContextRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSMisspelling
	}
	return r.IssueType
}

// IsPremium ports Rule.isPremium.
func (r *WrongWordInContextRule) IsPremium() bool {
	return r != nil && r.Premium
}

// GetTags ports Rule.getTags.
func (r *WrongWordInContextRule) GetTags() []Tag {
	if r == nil || len(r.Tags) == 0 {
		return nil
	}
	out := make([]Tag, len(r.Tags))
	copy(out, r.Tags)
	return out
}

// AddExamplePair ports Rule.addExamplePair.
func (r *WrongWordInContextRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *WrongWordInContextRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *WrongWordInContextRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// LoadWrongWordInContext loads tab-separated context confusion entries.
// Format: word1 word2 match1 match2 context1 context2 [explanation1 explanation2]
// Ports WrongWordInContextRule.loadContextWords.
func LoadWrongWordInContext(r io.Reader) ([]ContextWords, error) {
	var set []ContextWords
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if tools.JavaStringTrim(line) == "" || line[0] == '#' {
			continue
		}
		column := strings.Split(line, "\t")
		if len(column) < 6 {
			continue
		}
		var cw ContextWords
		var err error
		if err = cw.setWord(0, column[0]); err != nil {
			return nil, err
		}
		if err = cw.setWord(1, column[1]); err != nil {
			return nil, err
		}
		cw.Matches[0] = column[2]
		cw.Matches[1] = column[3]
		if err = cw.setContext(0, column[4]); err != nil {
			return nil, err
		}
		if err = cw.setContext(1, column[5]); err != nil {
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

// addBoundaries ports ContextWords.addBoundaries.
func addBoundaries(str string) string {
	ignoreCase := ""
	if strings.HasPrefix(str, "(?i)") {
		str = str[4:]
		ignoreCase = "(?i)"
	}
	// RE2: (?i) must be at start; non-capturing group is equivalent for matching.
	return ignoreCase + `\b(?:` + str + `)\b`
}

func (cw *ContextWords) setWord(i int, word string) error {
	re, err := regexp.Compile(addBoundaries(word))
	if err != nil {
		return err
	}
	cw.Words[i] = re
	pattern := word
	if strings.HasPrefix(pattern, "(?i)") {
		pattern = pattern[4:]
	} else if strings.HasPrefix(pattern, "(?-i)") {
		pattern = pattern[5:]
	}
	cw.wordPrefixes[i] = extractWordPrefix(pattern)
	return nil
}

func (cw *ContextWords) setContext(i int, context string) error {
	re, err := regexp.Compile(addBoundaries(context))
	if err != nil {
		return err
	}
	cw.Contexts[i] = re
	return nil
}

// CouldMatchSentence ports ContextWords.couldMatchSentence.
func (cw ContextWords) CouldMatchSentence(sentenceLower string) bool {
	word0Could := cw.wordPrefixes[0] == "" || strings.Contains(sentenceLower, cw.wordPrefixes[0])
	word1Could := cw.wordPrefixes[1] == "" || strings.Contains(sentenceLower, cw.wordPrefixes[1])
	return word0Could || word1Could
}

// extractWordPrefix ports ContextWords.extractWordPrefix.
func extractWordPrefix(pattern string) string {
	alternatives := splitTopLevelAlternatives(pattern)
	var prefixes []string
	for _, alt := range alternatives {
		p := extractSimplePrefix(alt)
		if len(p) < wrongWordMinPrefixLength {
			return ""
		}
		prefixes = append(prefixes, strings.ToLower(p))
	}
	if len(prefixes) == 0 {
		return ""
	}
	common := prefixes[0]
	for i := 1; i < len(prefixes); i++ {
		common = commonPrefix(common, prefixes[i])
		if len(common) < wrongWordMinPrefixLength {
			return ""
		}
	}
	return common
}

// splitTopLevelAlternatives ports ContextWords.splitTopLevelAlternatives.
func splitTopLevelAlternatives(pattern string) []string {
	var parts []string
	depth := 0
	start := 0
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		if c == '\\' {
			i++
			continue
		}
		if c == '(' || c == '[' {
			depth++
		} else if c == ')' || c == ']' {
			depth--
		} else if c == '|' && depth == 0 {
			parts = append(parts, pattern[start:i])
			start = i + 1
		}
	}
	parts = append(parts, pattern[start:])
	return parts
}

// extractSimplePrefix ports ContextWords.extractSimplePrefix.
func extractSimplePrefix(alt string) string {
	var sb strings.Builder
	for i := 0; i < len(alt); i++ {
		c := alt[i]
		if c == '\\' && i+1 < len(alt) {
			i++
			escaped := alt[i]
			if i+1 < len(alt) && (alt[i+1] == '?' || alt[i+1] == '*') {
				break // escaped char is optional
			}
			sb.WriteByte(escaped)
			if i+1 < len(alt) && alt[i+1] == '+' {
				i++ // skip quantifier
			}
		} else if strings.ContainsRune(`[(|.^$\{`, rune(c)) {
			break
		} else if c == '?' || c == '*' {
			s := sb.String()
			if len(s) > 0 {
				sb.Reset()
				sb.WriteString(s[:len(s)-1]) // preceding char was optional
			}
			break
		} else if c == '+' {
			break
		} else {
			if i+1 < len(alt) && (alt[i+1] == '?' || alt[i+1] == '*') {
				break // this char is optional
			}
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

func commonPrefix(a, b string) string {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return a[:i]
		}
	}
	return a[:n]
}

func (r *WrongWordInContextRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "WRONG_WORD_IN_CONTEXT"
}

func (r *WrongWordInContextRule) GetDescription() string {
	if r != nil && r.Description != "" {
		return r.Description
	}
	// Java base default when not overridden.
	return "Confusion of words"
}

// Match ports WrongWordInContextRule.match.
func (r *WrongWordInContextRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	// Java: sentence.getText().toLowerCase(Locale.ROOT)
	sentenceLower := strings.ToLower(sentence.GetText())
	for _, contextWords := range r.Entries {
		if !contextWords.CouldMatchSentence(sentenceLower) {
			continue
		}
		matchedWord := [2]bool{}
		matchedPos := [2]int{-1, -1}
		matchedTok := [2]string{}
		// search for word1 and word2 in a single pass, ignoring token 0 (SENT_START)
		for k := 1; k < len(tokens) && (!matchedWord[0] || !matchedWord[1]); k++ {
			// Java: if (!tokens[k].hasPartialPosTag("IS_URL"))
			if tokens[k].HasPartialPosTag("IS_URL") {
				continue
			}
			t, _ := joinNoSpaceTokens(tokens, k)
			if !matchedWord[0] && contextWords.Words[0].MatchString(t) {
				matchedWord[0] = true
				matchedPos[0] = k
				matchedTok[0] = t
			}
			if !matchedWord[1] && contextWords.Words[1].MatchString(t) {
				matchedWord[1] = true
				matchedPos[1] = k
				matchedTok[1] = t
			}
		}
		foundWord, notFoundWord := -1, -1
		startPos, endPos := 0, 0
		matchedToken := ""
		// when both words have been found, we cannot determine if one of them is wrong
		if matchedWord[0] && !matchedWord[1] {
			foundWord, notFoundWord = 0, 1
			// Java: startPos = tokens[matchedPos0].getStartPos();
			// endPos = tokens[matchedPos0].getStartPos() + token1.length();
			startPos = tokens[matchedPos[0]].GetStartPos()
			endPos = startPos + utf16TokenLen(matchedTok[0])
			matchedToken = matchedTok[0]
		} else if matchedWord[1] && !matchedWord[0] {
			foundWord, notFoundWord = 1, 0
			startPos = tokens[matchedPos[1]].GetStartPos()
			endPos = startPos + utf16TokenLen(matchedTok[1])
			matchedToken = matchedTok[1]
		}
		if foundWord == -1 {
			continue
		}
		matchedContext := [2]bool{}
		// search for both context patterns in a single pass, ignoring token 0 (SENT_START)
		for k := 1; k < len(tokens) && (!matchedContext[foundWord] || !matchedContext[notFoundWord]); k++ {
			if r.MatchLemmas {
				// Java: only non-empty lemmas from readings. When readings carry lemmas
				// (tagged analysis), match those only. lemmaCandidates also includes the
				// surface token so untagged AnalyzePlain can still exercise matchLemmas
				// languages (ES/CA); EN does not set MatchLemmas.
				for _, cand := range lemmaCandidates(tokens[k]) {
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
			// Java: SpecificIdRule(toId(getId()+"_"+matchedToken+"_"+repl, lang),
			// getDescription().replace("$match", matchedToken+"/"+repl), ...)
			id := tools.ToId(r.GetID()+"_"+matchedToken+"_"+repl, r.LanguageCode)
			desc := strings.ReplaceAll(r.GetDescription(), "$match", matchedToken+"/"+repl)
			specificIDRule := NewSpecificIdRule(id, desc, r.IsPremium(), r.GetCategory(), r.GetLocQualityIssueType(), r.GetTags())
			rm := NewRuleMatch(specificIDRule, sentence, startPos, endPos, msg)
			rm.ShortMessage = r.ShortMessageString
			// Java RuleMatch message parser also picks <suggestion>; set explicitly for parity.
			rm.SetSuggestedReplacement(repl)
			ruleMatches = append(ruleMatches, rm)
		}
	}
	return ruleMatches
}

func (r *WrongWordInContextRule) buildMessage(wrongWord, suggestion, explanationSuggestion, explanationWrongWord string) string {
	// Java uses Matcher.quoteReplacement then replaceFirst("\\$NAME", …).
	// Go strings.Replace is literal, so quoteReplacement is unnecessary (and would corrupt \$).
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

// lemmaCandidates returns surface form plus any reading lemmas (and light stem
// guesses used by MatchLemmas languages under untagged AnalyzePlain).
// Not used by EnglishWrongWordInContextRule (MatchLemmas=false).
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
	t := strings.ToLower(tok.GetToken())
	add(t)
	if len(t) > 3 && strings.HasSuffix(t, "s") {
		add(t[:len(t)-1])
	}
	if len(t) > 4 && strings.HasSuffix(t, "es") {
		add(t[:len(t)-2])
	}
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
// English tokens are single surface forms; this is a no-op for typical EN input.
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
