package km

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// KhmerWordRepeatRule ports org.languagetool.rules.km.KhmerWordRepeatRule.
//
// Java extends Rule directly (not WordRepeatRule). Custom match / isWord / ignore.
// Ctor: super(messages); setCategory(Categories.MISC.getCategory(messages)).
// Language ctor arg is unused in Java beyond the signature.
type KhmerWordRepeatRule struct {
	// Messages ports Rule.messages (ResourceBundle).
	Messages map[string]string
	// Category ports Rule.category (Java MISC).
	Category *rules.Category
}

// NewKhmerWordRepeatRule ports KhmerWordRepeatRule(ResourceBundle messages, Language language).
// Language is unused by Java match/isWord/ignore; Go omits it.
func NewKhmerWordRepeatRule(messages map[string]string) *KhmerWordRepeatRule {
	return &KhmerWordRepeatRule{
		Messages: messages,
		// Java: super.setCategory(Categories.MISC.getCategory(messages))
		Category: rules.CatMisc.GetCategory(messages),
	}
}

// GetID ports KhmerWordRepeatRule.getId → "KM_WORD_REPEAT_RULE".
func (r *KhmerWordRepeatRule) GetID() string { return "KM_WORD_REPEAT_RULE" }

// GetDescription ports KhmerWordRepeatRule.getDescription → messages.getString("desc_repetition").
func (r *KhmerWordRepeatRule) GetDescription() string {
	if r != nil && r.Messages != nil {
		return r.Messages["desc_repetition"]
	}
	return ""
}

// GetCategory ports Rule.getCategory.
func (r *KhmerWordRepeatRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// Match ports KhmerWordRepeatRule.match.
func (r *KhmerWordRepeatRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var ruleMatches []*rules.RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	tokensWithWS := sentence.GetTokens()

	prevToken := ""
	// we start from token 1, token 0 is SENT_START
	for i := 1; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		if isWordKM(token) && strings.EqualFold(prevToken, token) && !r.ignore(sentence, tokensWithWS, i) {
			prevPos := tokens[i-1].GetStartPos()
			pos := tokens[i].GetStartPos()
			// Java: new RuleMatch(this, sentence, prevPos, pos+prevToken.length(),
			//   messages.getString("repetition"), messages.getString("desc_repetition_short"))
			// String.length() is UTF-16 code units.
			msg := ""
			short := ""
			if r != nil && r.Messages != nil {
				msg = r.Messages["repetition"]
				short = r.Messages["desc_repetition_short"]
			}
			rm := rules.NewRuleMatch(r, sentence, prevPos, pos+utf16LenKM(prevToken), msg)
			rm.ShortMessage = short
			// Java replacement order:
			// 1. prevToken + " " + token  (replace ZWSP with real space)
			// 2. prevToken               (remove repeated word)
			// 3. prevToken + "ៗ"        (repetition character)
			rm.SetSuggestedReplacements([]string{
				prevToken + " " + token,
				prevToken,
				prevToken + "ៗ",
			})
			ruleMatches = append(ruleMatches, rm)
		}
		prevToken = token
	}
	return ruleMatches
}

// isWordKM ports KhmerWordRepeatRule.isWord — avoid "..." etc. to be matched.
// Java: token.length()==1 && !Character.isLetter(charAt(0)) → false; else true.
// Note: unlike WordRepeatRule.isWord, no emoji / numeric-space arms.
func isWordKM(token string) bool {
	if utf16LenKM(token) == 1 {
		// length 1 ⇒ single BMP code unit; charAt(0) / first rune agree
		c := []rune(token)[0]
		if !unicode.IsLetter(c) {
			return false
		}
	}
	return true
}

// ignore ports KhmerWordRepeatRule.ignore:
//
//	int origPos = sentence.getOriginalPosition(position);
//	if (position >= 1 && "\u0020".equals(tokensWithWhiteSpace[origPos-1].getToken())) {
//	  return true;
//	}
//	return false;
//
// Uses AnalyzedSentence.GetOriginalPosition (Java whPositions), not a hand-rolled map.
func (r *KhmerWordRepeatRule) ignore(sentence *languagetool.AnalyzedSentence, tokensWithWhiteSpace []*languagetool.AnalyzedTokenReadings, position int) bool {
	origPos := sentence.GetOriginalPosition(position)
	if position >= 1 && origPos >= 1 && origPos-1 < len(tokensWithWhiteSpace) &&
		tokensWithWhiteSpace[origPos-1].GetToken() == "\u0020" {
		return true
	}
	return false
}

// utf16LenKM ports Java String.length() (UTF-16 code units) for match spans / isWord.
func utf16LenKM(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}
