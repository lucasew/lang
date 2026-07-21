package rules

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var (
	// Java NUMERALS_EN: "[a-z]|(m{0,4}(c[md]|d?c{0,3})(x[cl]|l?x{0,3})(i[xv]|v?i{0,3}))$"
	// String.matches → full match; roman numerals are lowercase only (no CASE_INSENSITIVE).
	numeralsEN          = regexp.MustCompile(`^(?:[a-z]|m{0,4}(?:c[md]|d?c{0,3})(?:x[cl]|l?x{0,3})(?:i[xv]|v?i{0,3}))$`)
	containsDigitRE     = regexp.MustCompile(`\d`)
	onlyLowercaseStart  = regexp.MustCompile(`^[a-z][A-Z]`)
	whitespaceOrQuote   = regexp.MustCompile(`^[ "'„«»‘’“”\n]$`)
	digitDot            = regexp.MustCompile(`^\d+\. .*`)
	linebreakDigitDot   = regexp.MustCompile(`\n\d+\. `)
	uppercaseExceptions = map[string]bool{
		"n": true, "w": true, "x86": true, "ⓒ": true, "ø": true,
		"cc": true, "pH": true, "heylogin": true,
	}
	dutchSpecialTokens = map[string]bool{"k": true, "m": true, "n": true, "r": true, "s": true, "t": true}
)

// UppercaseSentenceStartRule ports org.languagetool.rules.UppercaseSentenceStartRule.
// Java: CASING, Typographical.
type UppercaseSentenceStartRule struct {
	Messages map[string]string
	LangCode string // short code e.g. "en"
	// Category ports Rule.category (Java CASING).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Typographical).
	IssueType ITSIssueType
	// IsException skips this sentence's start check (language-specific).
	IsException func(tokens []*languagetool.AnalyzedTokenReadings, tokenIdx int) bool
	// AntiPatterns ports Rule.getAntiPatterns; MatchList uses SentenceWithImmunization.
	AntiPatterns []SentenceReplacer
}

func NewUppercaseSentenceStartRule(messages map[string]string, langCode string) *UppercaseSentenceStartRule {
	return &UppercaseSentenceStartRule{
		Messages:  messages,
		LangCode:  langCode,
		Category:  CatCasing.GetCategory(messages),
		IssueType: ITSTypographical,
	}
}

func (r *UppercaseSentenceStartRule) GetID() string { return "UPPERCASE_SENTENCE_START" }

// GetDescription ports getDescription (desc_uppercase_sentence).
func (r *UppercaseSentenceStartRule) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["desc_uppercase_sentence"]; s != "" {
			return s
		}
	}
	return "Checks that a sentence starts with an uppercase letter"
}

func (r *UppercaseSentenceStartRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *UppercaseSentenceStartRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSTypographical
	}
	return r.IssueType
}

// MinToCheckParagraph ports minToCheckParagraph (Java returns 0).
func (r *UppercaseSentenceStartRule) MinToCheckParagraph() int { return 0 }

// EstimateContextForSureMatch ports TextLevelRule (Java always -1).
func (r *UppercaseSentenceStartRule) EstimateContextForSureMatch() int { return -1 }

func (r *UppercaseSentenceStartRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	lastParagraphString := ""
	var ruleMatches []*RuleMatch
	if len(sentences) == 1 && len(sentences[0].GetTokens()) == 2 {
		return ruleMatches
	}
	pos := 0
	isPrevSentenceNumberedList := false
	for _, sentence := range sentences {
		// Java: getSentenceWithImmunization(sentence).getTokensWithoutWhitespace()
		work := sentence
		if r != nil {
			work = SentenceWithImmunization(sentence, r.AntiPatterns)
		}
		tokens := work.GetTokensWithoutWhitespace()
		if len(tokens) < 2 {
			// Java: return immediately (drop remaining sentences)
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
		if dutch := r.dutchSpecialCase(firstToken, secondToken, tokens); dutch != "" {
			thirdToken = dutch
			matchTokenPos = 3
		}

		// Java isException: return toRuleMatchArray (abort whole match)
		if r.IsException != nil && r.IsException(tokens, matchTokenPos) {
			return ruleMatches
		}

		checkToken := firstToken
		if thirdToken != "" {
			checkToken = thirdToken
		} else if secondToken != "" {
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
		if !sentenceEnd1Matches(lastParagraphString) && !isSentenceEnd(lastToken) {
			preventError = true
		}

		// Java: !sentence.getText().replace('\u00A0',' ').trim().isEmpty()
		if !tools.JavaStringTrimIsEmpty(strings.ReplaceAll(sentence.GetText(), "\u00A0", " ")) {
			lastParagraphString = lastToken
		}

		if matchTokenPos+1 < len(tokens) &&
			numeralsEN.MatchString(tokens[matchTokenPos].GetToken()) &&
			(tokens[matchTokenPos+1].GetToken() == "." || tokens[matchTokenPos+1].GetToken() == ")") {
			preventError = true
		}

		if isPrevSentenceNumberedList || tokenizers.IsURL(checkToken) || tokenizers.IsEMail(checkToken) ||
			firstTokenObj.IsImmunized() || tokens[matchTokenPos].HasPosTag("_IS_URL") {
			preventError = true
		}

		if javaStringLenUpper(checkToken) > 0 {
			// Java Character.isLowerCase(checkToken.charAt(0)) — first UTF-16 unit
			firstChar := javaFirstUTF16Rune(checkToken)
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
				// Java: setShortMessage(messages.getString("category_case"))
				if r.Messages != nil {
					if sm := r.Messages["category_case"]; sm != "" {
						rm.ShortMessage = sm
					}
				}
				ruleMatches = append(ruleMatches, rm)
			}
		}
		pos += sentence.GetCorrectedTextLength()
		isPrevSentenceNumberedList = digitDot.MatchString(sentence.GetText()) || linebreakDigitDot.MatchString(sentence.GetText())
	}
	return ruleMatches
}

// dutchSpecialCase ports UppercaseSentenceStartRule.dutchSpecialCase.
func (r *UppercaseSentenceStartRule) dutchSpecialCase(firstToken, secondToken string, tokens []*languagetool.AnalyzedTokenReadings) string {
	if r == nil || r.LangCode != "nl" {
		return ""
	}
	if len(tokens) > 3 && firstToken == "'" && dutchSpecialTokens[secondToken] {
		return tokens[3].GetToken()
	}
	return ""
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
	// Java: pt short code adds dialogue dashes
	if r != nil && r.LangCode == "pt" {
		for _, q := range []string{"-", "–", "—"} {
			if word == q {
				return true
			}
		}
	}
	return false
}

func javaStringLenUpper(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func javaFirstUTF16Rune(s string) rune {
	u := utf16.Encode([]rune(s))
	if len(u) == 0 {
		return 0
	}
	return rune(u[0])
}
