package rules

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var (
	numeralsEN          = regexp.MustCompile(`^[a-z]$|^(?i)m{0,4}(c[md]|d?c{0,3})(x[cl]|l?x{0,3})(i[xv]|v?i{0,3})$`)
	containsDigitRE     = regexp.MustCompile(`\d`)
	onlyLowercaseStart  = regexp.MustCompile(`^[a-z][A-Z]`)
	whitespaceOrQuote   = regexp.MustCompile(`^[ "'„«»‘’“”\n]$`)
	digitDot            = regexp.MustCompile(`^\d+\. `)
	linebreakDigitDot   = regexp.MustCompile(`\n\d+\. `)
	uppercaseExceptions = map[string]bool{
		"n": true, "w": true, "x86": true, "ⓒ": true, "ø": true,
		"cc": true, "pH": true, "heylogin": true,
	}
)

// UppercaseSentenceStartRule ports org.languagetool.rules.UppercaseSentenceStartRule.
type UppercaseSentenceStartRule struct {
	Messages map[string]string
	LangCode string // short code e.g. "en"
	// IsException skips this sentence's start check (language-specific).
	IsException func(tokens []*languagetool.AnalyzedTokenReadings, tokenIdx int) bool
}

func NewUppercaseSentenceStartRule(messages map[string]string, langCode string) *UppercaseSentenceStartRule {
	return &UppercaseSentenceStartRule{Messages: messages, LangCode: langCode}
}

func (r *UppercaseSentenceStartRule) GetID() string { return "UPPERCASE_SENTENCE_START" }

func (r *UppercaseSentenceStartRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	lastParagraphString := ""
	var ruleMatches []*RuleMatch
	if len(sentences) == 1 && len(sentences[0].GetTokens()) == 2 {
		return ruleMatches
	}
	pos := 0
	isPrevSentenceNumberedList := false
	for _, sentence := range sentences {
		tokens := sentence.GetTokensWithoutWhitespace()
		if len(tokens) < 2 {
			return ruleMatches
		}
		matchTokenPos := 1
		firstTokenObj := tokens[matchTokenPos]
		firstToken := firstTokenObj.GetToken()
		var secondToken, thirdToken string
		if len(tokens) >= 3 && r.isQuoteStart(firstToken) {
			matchTokenPos = 2
			secondToken = tokens[matchTokenPos].GetToken()
		}
		// dutch special skipped for en demo
		_ = thirdToken

		if r.IsException != nil && r.IsException(tokens, matchTokenPos) {
			pos += sentence.GetCorrectedTextLength()
			continue
		}

		checkToken := firstToken
		if secondToken != "" {
			checkToken = secondToken
		}

		lastToken := tokens[len(tokens)-1].GetToken()
		if whitespaceOrQuote.MatchString(lastToken) && len(tokens) >= 2 {
			lastToken = tokens[len(tokens)-2].GetToken()
		}

		preventError := false
		if lastParagraphString == "," || lastParagraphString == ";" {
			preventError = true
		}
		if containsDigitRE.MatchString(tokens[matchTokenPos].GetToken()) {
			preventError = true
		}
		// Java: !SENTENCE_END1.matcher(lastParagraphString).matches() && !isSentenceEnd(lastToken)
		// SENTENCE_END1 matches "", ".", "?", "!", "…"
		if !sentenceEnd1Matches(lastParagraphString) && !isSentenceEnd(lastToken) {
			preventError = true
		}

		if strings.TrimSpace(strings.ReplaceAll(sentence.GetText(), "\u00A0", " ")) != "" {
			lastParagraphString = lastToken
		}

		if matchTokenPos+1 < len(tokens) &&
			numeralsEN.MatchString(tokens[matchTokenPos].GetToken()) &&
			(tokens[matchTokenPos+1].GetToken() == "." || tokens[matchTokenPos+1].GetToken() == ")") {
			preventError = true
		}

		if isPrevSentenceNumberedList || tokenizers.IsURL(checkToken) || tokenizers.IsEMail(checkToken) || firstTokenObj.IsImmunized() {
			preventError = true
		}

		if len(checkToken) > 0 {
			firstChar := []rune(checkToken)[0]
			capitalized := tools.UppercaseFirstChar(checkToken)
			if capitalized != checkToken && !preventError && unicode.IsLower(firstChar) &&
				!onlyLowercaseStart.MatchString(checkToken) &&
				!uppercaseExceptions[checkToken] && !tools.IsCamelCase(checkToken) {
				msg := "This sentence does not start with an uppercase letter"
				if r.Messages != nil {
					if m, ok := r.Messages["incorrect_case"]; ok {
						msg = m
					}
				}
				from := pos + tokens[matchTokenPos].GetStartPos()
				to := pos + tokens[matchTokenPos].GetEndPos()
				rm := NewRuleMatch(r, sentence, from, to, msg)
				rm.SetSuggestedReplacement(capitalized)
				ruleMatches = append(ruleMatches, rm)
			}
		}
		pos += sentence.GetCorrectedTextLength()
		isPrevSentenceNumberedList = digitDot.MatchString(sentence.GetText()) || linebreakDigitDot.MatchString(sentence.GetText())
	}
	return ruleMatches
}

func sentenceEnd1Matches(s string) bool {
	// Java Pattern.compile("[.?!…]|") — matches empty or one of those
	if s == "" {
		return true
	}
	return s == "." || s == "?" || s == "!" || s == "…"
}

func isSentenceEnd(word string) bool {
	return word == "." || word == "?" || word == "!" || word == "…"
}

func (r *UppercaseSentenceStartRule) isQuoteStart(word string) bool {
	base := []string{"\"", "'", "„", "»", "«", "“", "‘", "¡", "¿"}
	for _, q := range base {
		if word == q {
			return true
		}
	}
	return false
}
