package rules

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractStyleRepeatedWordRule ports
// org.languagetool.rules.AbstractStyleRepeatedWordRule (text-level; default off in Java).
// Java: STYLE, Style, setDefaultOff().
type AbstractStyleRepeatedWordRule struct {
	ID                     string
	Description            string
	MaxDistanceOfSentences int  // default 1
	ExcludeDirectSpeech    bool // default true in Java
	// Category ports Rule.category (Java STYLE).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Style).
	IssueType ITSIssueType
	// DefaultOff ports setDefaultOff (Java true).
	DefaultOff bool
	// IsTokenToCheck: Java isTokenToCheck(tokens, n). Nil → letters len≥2.
	IsTokenToCheck func(tokens []*languagetool.AnalyzedTokenReadings, n int) bool
	// IsTokenPair excludes pairs like "Arm in Arm".
	IsTokenPair func(tokens []*languagetool.AnalyzedTokenReadings, n int, before bool) bool
	// IsPartOfWord for compound languages (German).
	IsPartOfWord func(testTokenText, tokenText string) bool
	// IsExceptionPair after same-lemma match.
	IsExceptionPair func(token1, token2 *languagetool.AnalyzedTokenReadings) bool
	// SetURL ports setURL(token) — optional match-level URL (e.g. DE OpenThesaurus).
	// Nil → no URL on matches (Java default null).
	SetURL func(token *languagetool.AnalyzedTokenReadings) string
	// Messages
	MessageSameSentence   func() string
	MessageSentenceBefore func() string
	MessageSentenceAfter  func() string
}

const maxTokenToCheckStyle = 5

var (
	styleOpenQuotes   = regexp.MustCompile(`^["“„»«]$`)
	styleEndQuotes    = regexp.MustCompile(`^["“”»«]$`)
	styleSingleQuotes = regexp.MustCompile(`^['‚‘’'›‹]$`)
)

func NewAbstractStyleRepeatedWordRule() *AbstractStyleRepeatedWordRule {
	return &AbstractStyleRepeatedWordRule{
		ID:                     "STYLE_REPEATED_WORD_RULE",
		Description:            "Repeated words in consecutive sentences",
		MaxDistanceOfSentences: 1,
		ExcludeDirectSpeech:    true,
		// Category filled when language sets Messages via InitStyleRepeatedWordMeta.
		IssueType:  ITSStyle,
		DefaultOff: true,
	}
}

// InitStyleRepeatedWordMeta applies Java AbstractStyleRepeatedWordRule ctor metadata.
func InitStyleRepeatedWordMeta(r *AbstractStyleRepeatedWordRule, messages map[string]string) {
	if r == nil {
		return
	}
	if r.Category == nil {
		r.Category = CatStyle.GetCategory(messages)
	}
	if r.IssueType == "" {
		r.IssueType = ITSStyle
	}
	r.DefaultOff = true
}

func (r *AbstractStyleRepeatedWordRule) GetID() string {
	if r != nil && r.ID != "" {
		return r.ID
	}
	return "STYLE_REPEATED_WORD_RULE"
}

func (r *AbstractStyleRepeatedWordRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *AbstractStyleRepeatedWordRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

func (r *AbstractStyleRepeatedWordRule) IsDefaultOff() bool { return r != nil && r.DefaultOff }

func (r *AbstractStyleRepeatedWordRule) GetDescription() string {
	if r != nil && r.Description != "" {
		return r.Description
	}
	return "Repeated words in consecutive sentences"
}

func hasBreakTokenStyleAbstract(tokens []*languagetool.AnalyzedTokenReadings) bool {
	for i := 0; i < len(tokens) && i < maxTokenToCheckStyle; i++ {
		if tokens[i] == nil {
			continue
		}
		t := tokens[i].GetToken()
		if t == "-" || t == "—" || t == "–" {
			return true
		}
	}
	return false
}

func isQuestionResponseStyle(nAct, nTest int, tokenList [][]*languagetool.AnalyzedTokenReadings) bool {
	dist := nAct - nTest
	if dist != 1 && dist != -1 {
		return false
	}
	if nAct < 0 || nTest < 0 || nAct >= len(tokenList) || nTest >= len(tokenList) {
		return false
	}
	actTokens := tokenList[nAct]
	testTokens := tokenList[nTest]
	if len(actTokens) < 2 || len(testTokens) < 2 {
		return false
	}
	actToken := actTokens[len(actTokens)-1].GetToken()
	if styleEndQuotes.MatchString(actToken) && len(actTokens) >= 2 {
		actToken = actTokens[len(actTokens)-2].GetToken()
	}
	testToken := testTokens[len(testTokens)-1].GetToken()
	if styleEndQuotes.MatchString(testToken) && len(testTokens) >= 2 {
		testToken = testTokens[len(testTokens)-2].GetToken()
	}
	return (actToken == "?" && testToken != "?") || (testToken == "?" && actToken != "?")
}

func isInQuotesStyle(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if i <= 0 || i >= len(tokens)-1 || tokens[i-1] == nil || tokens[i+1] == nil {
		return false
	}
	prev, next := tokens[i-1].GetToken(), tokens[i+1].GetToken()
	return (styleOpenQuotes.MatchString(prev) || styleSingleQuotes.MatchString(prev)) &&
		(styleEndQuotes.MatchString(next) || styleSingleQuotes.MatchString(next))
}

func (r *AbstractStyleRepeatedWordRule) isTokenToCheckDefault(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
	if r.IsTokenToCheck != nil {
		return r.IsTokenToCheck(tokens, n)
	}
	if n < 0 || n >= len(tokens) || tokens[n] == nil {
		return false
	}
	t := tokens[n].GetToken()
	return len([]rune(t)) >= 2
}

func (r *AbstractStyleRepeatedWordRule) isPartOfWord(a, b string) bool {
	if r.IsPartOfWord != nil {
		return r.IsPartOfWord(a, b)
	}
	return false
}

func (r *AbstractStyleRepeatedWordRule) isExceptionPair(a, b *languagetool.AnalyzedTokenReadings) bool {
	if r.IsExceptionPair != nil {
		return r.IsExceptionPair(a, b)
	}
	return false
}

func (r *AbstractStyleRepeatedWordRule) isTokenPair(tokens []*languagetool.AnalyzedTokenReadings, n int, before bool) bool {
	if r.IsTokenPair != nil {
		return r.IsTokenPair(tokens, n, before)
	}
	return false
}

func tokenLemmas(t *languagetool.AnalyzedTokenReadings) []string {
	if t == nil {
		return nil
	}
	var lemmas []string
	for _, rd := range t.GetReadings() {
		if rd == nil {
			continue
		}
		if l := rd.GetLemma(); l != nil && *l != "" {
			lemmas = append(lemmas, *l)
		}
	}
	return lemmas
}

func (r *AbstractStyleRepeatedWordRule) isTokenInSentence(
	testToken *languagetool.AnalyzedTokenReadings,
	tokens []*languagetool.AnalyzedTokenReadings,
	notCheck int,
	isDirectSpeech bool,
) bool {
	if testToken == nil || tokens == nil {
		return false
	}
	lemmas := tokenLemmas(testToken)
	excludeDS := r.ExcludeDirectSpeech
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == nil {
			continue
		}
		if excludeDS && !isDirectSpeech && styleOpenQuotes.MatchString(tokens[i].GetToken()) &&
			i < len(tokens)-1 && tokens[i+1] != nil && !tokens[i+1].IsWhitespaceBefore() {
			isDirectSpeech = true
		} else if excludeDS && isDirectSpeech && styleEndQuotes.MatchString(tokens[i].GetToken()) &&
			i > 1 && !tokens[i].IsWhitespaceBefore() {
			isDirectSpeech = false
		} else if i != notCheck && !isDirectSpeech && !isInQuotesStyle(tokens, i) && r.isTokenToCheckDefault(tokens, i) {
			sameLemma := len(lemmas) > 0 && tokens[i].HasAnyLemma(lemmas...) && !r.isExceptionPair(testToken, tokens[i])
			partOf := r.isPartOfWord(testToken.GetToken(), tokens[i].GetToken())
			// surface fallback when no lemmas: equal ignore case
			if !sameLemma && len(lemmas) == 0 &&
				strings.EqualFold(testToken.GetToken(), tokens[i].GetToken()) {
				sameLemma = true
			}
			if sameLemma || partOf {
				if notCheck >= 0 {
					if notCheck == i-2 {
						return !r.isTokenPair(tokens, i, true)
					} else if notCheck == i+2 {
						return !r.isTokenPair(tokens, i, false)
					} else if (notCheck == i+1 || notCheck == i-1) &&
						testToken.GetToken() == tokens[i].GetToken() {
						return false
					}
				}
				return true
			}
		}
	}
	return false
}

func getStartsWithDirectSpeech(n int, sentences []*languagetool.AnalyzedSentence, isDirectSpeech bool) bool {
	if n <= 0 {
		return false
	}
	sentence := sentences[n-1].GetTokensWithoutWhitespace()
	for i := 0; i < len(sentence); i++ {
		token := sentence[i]
		if token == nil {
			continue
		}
		if !isDirectSpeech && styleOpenQuotes.MatchString(token.GetToken()) &&
			i < len(sentence)-1 && sentence[i+1] != nil && !sentence[i+1].IsWhitespaceBefore() {
			isDirectSpeech = true
		} else if isDirectSpeech && styleEndQuotes.MatchString(token.GetToken()) &&
			i > 1 && !token.IsWhitespaceBefore() {
			isDirectSpeech = false
		}
	}
	return isDirectSpeech
}

// MatchList ports AbstractStyleRepeatedWordRule.match.
func (r *AbstractStyleRepeatedWordRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	if r == nil {
		return nil
	}
	maxDist := r.MaxDistanceOfSentences
	if maxDist < 0 {
		maxDist = 1
	}
	var ruleMatches []*RuleMatch
	var tokenList [][]*languagetool.AnalyzedTokenReadings
	var isDSList []bool
	pos := 0
	startsWithDirectSpeech := false
	excludeDS := r.ExcludeDirectSpeech

	for n := 0; n < maxDist && n < len(sentences); n++ {
		if sentences[n] == nil {
			tokenList = append(tokenList, nil)
			isDSList = append(isDSList, false)
			continue
		}
		tokenList = append(tokenList, sentences[n].GetTokensWithoutWhitespace())
		startsWithDirectSpeech = getStartsWithDirectSpeech(n, sentences, startsWithDirectSpeech)
		isDSList = append(isDSList, startsWithDirectSpeech)
	}

	isDirectSpeech := false
	for n := 0; n < len(sentences); n++ {
		if n+maxDist < len(sentences) && sentences[n+maxDist] != nil {
			tokenList = append(tokenList, sentences[n+maxDist].GetTokensWithoutWhitespace())
			startsWithDirectSpeech = getStartsWithDirectSpeech(n+maxDist, sentences, startsWithDirectSpeech)
			isDSList = append(isDSList, startsWithDirectSpeech)
		}
		if len(tokenList) > 2*maxDist+1 {
			tokenList = tokenList[1:]
			isDSList = isDSList[1:]
		}
		nTok := maxDist
		if n < maxDist {
			nTok = n
		} else if n >= len(sentences)-maxDist {
			nTok = len(tokenList) - (len(sentences) - n)
		}
		if nTok < 0 || nTok >= len(tokenList) || tokenList[nTok] == nil {
			if sentences[n] != nil {
				pos += sentences[n].GetCorrectedTextLength()
			}
			continue
		}
		if !hasBreakTokenStyleAbstract(tokenList[nTok]) {
			for i := 0; i < len(tokenList[nTok]); i++ {
				token := tokenList[nTok][i]
				if token == nil {
					continue
				}
				if excludeDS && !isDirectSpeech && styleOpenQuotes.MatchString(token.GetToken()) &&
					i < len(tokenList[nTok])-1 && tokenList[nTok][i+1] != nil &&
					!tokenList[nTok][i+1].IsWhitespaceBefore() {
					isDirectSpeech = true
				} else if excludeDS && isDirectSpeech && styleEndQuotes.MatchString(token.GetToken()) &&
					i > 1 && !token.IsWhitespaceBefore() {
					isDirectSpeech = false
				} else if !isDirectSpeech && !isInQuotesStyle(tokenList[nTok], i) &&
					r.isTokenToCheckDefault(tokenList[nTok], i) {
					isRepeated := 0
					dsTok := false
					if nTok < len(isDSList) {
						dsTok = isDSList[nTok]
					}
					if r.isTokenInSentence(token, tokenList[nTok], i, dsTok) {
						isRepeated = 1
					}
					for j := nTok - 1; isRepeated == 0 && j >= 0 && j >= nTok-maxDist; j-- {
						if !isQuestionResponseStyle(nTok, j, tokenList) {
							dsj := false
							if j < len(isDSList) {
								dsj = isDSList[j]
							}
							if r.isTokenInSentence(token, tokenList[j], -1, dsj) {
								isRepeated = 2
							}
						}
					}
					for j := nTok + 1; isRepeated == 0 && j < len(tokenList) && j <= nTok+maxDist; j++ {
						if !isQuestionResponseStyle(nTok, j, tokenList) {
							dsj := false
							if j < len(isDSList) {
								dsj = isDSList[j]
							}
							if r.isTokenInSentence(token, tokenList[j], -1, dsj) {
								isRepeated = 3
							}
						}
					}
					if isRepeated != 0 {
						var msg string
						switch isRepeated {
						case 1:
							if r.MessageSameSentence != nil {
								msg = r.MessageSameSentence()
							} else {
								msg = "Repeated word in the same sentence"
							}
						case 2:
							if r.MessageSentenceBefore != nil {
								msg = r.MessageSentenceBefore()
							} else {
								msg = "Repeated word in a previous sentence"
							}
						default:
							if r.MessageSentenceAfter != nil {
								msg = r.MessageSentenceAfter()
							} else {
								msg = "Repeated word in a following sentence"
							}
						}
						// Java RuleMatch sentence may be null; attach current sentence for offsets
						sent := sentences[n]
						rm := NewRuleMatch(r, sent, pos+token.GetStartPos(), pos+token.GetEndPos(), msg)
						// Java: URL url = setURL(token); if (url != null) ruleMatch.setUrl(url);
						if r.SetURL != nil {
							if u := r.SetURL(token); u != "" {
								rm.SetURL(u)
							}
						}
						ruleMatches = append(ruleMatches, rm)
					}
				}
			}
		}
		if sentences[n] != nil {
			pos += sentences[n].GetCorrectedTextLength()
		}
	}
	return ruleMatches
}
