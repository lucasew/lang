package uk

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// TypographyRule ports org.languagetool.rules.uk.TypographyRule.
// Java: setCategory(Categories.TYPOGRAPHY).
type TypographyRule struct {
	Messages map[string]string
	Category *rules.Category
}

func NewTypographyRule(messages map[string]string) *TypographyRule {
	return &TypographyRule{
		Messages: messages,
		Category: rules.CatTypography.GetCategory(messages),
	}
}

func (r *TypographyRule) GetID() string { return "DASH" }

func (r *TypographyRule) GetDescription() string {
	return "Коротка риска замість дефісу"
}

// GetCategory ports Rule.getCategory (Java TYPOGRAPHY).
func (r *TypographyRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

var (
	typoCyrRE     = regexp.MustCompile(`[а-яїієґА-ЯІЇЄҐ]`)
	shortDashWord = regexp.MustCompile(`(?i)[а-яіїєґ']{2,}([\x{2013}\x{2014}][а-яіїєґ']{2,})+`)
	badLatinDash  = regexp.MustCompile(`[ХІXIV]+[\x{2013}\x{2014}][ХІXIV]+`)
	dashChars     = regexp.MustCompile(`[\x{2013}\x{2014}]`)
)

func (r *TypographyRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var ruleMatches []*rules.RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	msg := "Риска всередині слова. Всередині слова вживайте дефіс, між словами виокремлюйте риску пробілами."

	for i := 1; i < len(tokens); i++ {
		// Path A: compound with en/em dash as single token (rare with WordTokenizer)
		if tok := shortDashInToken(tokens[i].GetToken()); tok != "" {
			rm := rules.NewRuleMatch(r, sentence, tokens[i].GetStartPos(), tokens[i].GetEndPos(), msg)
			rm.ShortMessage = "Коротка риска"
			rm.SetSuggestedReplacements([]string{
				dashChars.ReplaceAllString(tok, "-"),
				dashChars.ReplaceAllString(tok, " \u2014 "),
			})
			ruleMatches = append(ruleMatches, rm)
			continue
		}

		// Path A2: tokenizer split "word–word" into three tokens
		if i+1 < len(tokens) {
			dash := tokens[i].GetToken()
			if (dash == "\u2013" || dash == "\u2014") &&
				!tokens[i].IsWhitespaceBefore() &&
				!tokens[i+1].IsWhitespaceBefore() {
				left, right := tokens[i-1].GetToken(), tokens[i+1].GetToken()
				combined := left + dash + right
				if shortDashInToken(combined) != "" {
					rm := rules.NewRuleMatch(r, sentence, tokens[i-1].GetStartPos(), tokens[i+1].GetEndPos(), msg)
					rm.ShortMessage = "Коротка риска"
					rm.SetSuggestedReplacements([]string{
						dashChars.ReplaceAllString(combined, "-"),
						dashChars.ReplaceAllString(combined, " \u2014 "),
					})
					ruleMatches = append(ruleMatches, rm)
					i++ // skip right
					continue
				}
			}
		}

		// Path B: bare en/em dash with missing spaces
		t := tokens[i].GetToken()
		if t != "\u2014" && t != "\u2013" {
			continue
		}
		noSpaceLeft := i > 1 && !tokens[i].IsWhitespaceBefore() &&
			tokens[i-1].GetToken() != "," && tokens[i-1].GetToken() != "«"
		noSpaceRight := i < len(tokens)-1 && !tokens[i+1].IsWhitespaceBefore() &&
			tokens[i+1].GetToken() != ">"
		if !noSpaceLeft && !noSpaceRight {
			continue
		}
		if i > 1 && isDigitsOnly(tokens[i-1].GetToken()) && i < len(tokens)-1 && isDigitsOnly(tokens[i+1].GetToken()) {
			continue
		}
		// А–Т / ХХ–ХХІ: short or roman-like sides kept as non-errors in Java
		if i > 1 && i < len(tokens)-1 && !tokens[i].IsWhitespaceBefore() && !tokens[i+1].IsWhitespaceBefore() {
			left, right := tokens[i-1].GetToken(), tokens[i+1].GetToken()
			if isRomanishDash(left) && isRomanishDash(right) {
				continue
			}
			if len([]rune(left)) <= 1 && len([]rune(right)) <= 1 {
				continue
			}
		}

		var replacements []string
		if i > 1 && i < len(tokens)-1 &&
			typoCyrRE.MatchString(tokens[i-1].GetToken()) &&
			typoCyrRE.MatchString(tokens[i+1].GetToken()) &&
			(noSpaceLeft || noSpaceRight) {
			// only add hyphen compound if both sides glued or left glued?
			// Java always adds if both cyrillic when noSpace left or right
			if noSpaceLeft && noSpaceRight {
				replacements = append(replacements, tokens[i-1].GetToken()+"-"+tokens[i+1].GetToken())
			} else if noSpaceRight && tokens[i].IsWhitespaceBefore() {
				// "цукерок —знову" — space left, no space right: still adds hyphen form
				replacements = append(replacements, tokens[i-1].GetToken()+"-"+tokens[i+1].GetToken())
			} else if noSpaceLeft && i < len(tokens)-1 {
				replacements = append(replacements, tokens[i-1].GetToken()+"-"+tokens[i+1].GetToken())
			}
		}

		// spaced em dash form
		startPos := tokens[i].GetStartPos()
		endPos := tokens[i].GetEndPos()
		var b strings.Builder
		if i > 1 && tokens[i-1].GetToken() != "," {
			b.WriteString(tokens[i-1].GetToken())
			b.WriteByte(' ')
			startPos = tokens[i-1].GetStartPos()
		}
		b.WriteString("\u2014")
		if i < len(tokens)-1 {
			b.WriteByte(' ')
			b.WriteString(tokens[i+1].GetToken())
			endPos = tokens[i+1].GetEndPos()
		}
		// Edge: dash at start "—знову"
		if i == 1 || (i > 1 && tokens[i-1].GetToken() == ",") {
			startPos = tokens[i].GetStartPos()
			if i < len(tokens)-1 {
				replacements = []string{"\u2014 " + tokens[i+1].GetToken()}
				// also keep hyphen form if applicable already in list
			} else {
				replacements = []string{"\u2014"}
			}
			// rebuild
			if i < len(tokens)-1 {
				endPos = tokens[i+1].GetEndPos()
			}
		} else if i == len(tokens)-1 {
			replacements = []string{tokens[i-1].GetToken() + " \u2014"}
			startPos = tokens[i-1].GetStartPos()
			endPos = tokens[i].GetEndPos()
		} else {
			spaced := b.String()
			// Avoid "left — right" when left was comma case
			replacements = append(replacements, spaced)
		}

		replacements = uniqueNonEmpty(replacements)
		if len(replacements) == 0 {
			continue
		}
		rm := rules.NewRuleMatch(r, sentence, startPos, endPos, msg)
		rm.ShortMessage = "Коротка риска"
		rm.SetSuggestedReplacements(replacements)
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}

func shortDashInToken(tok string) string {
	if tok == "" {
		return ""
	}
	if !strings.ContainsRune(tok, '\u2013') && !strings.ContainsRune(tok, '\u2014') {
		return ""
	}
	// must not start with dash
	if strings.HasPrefix(tok, "\u2013") || strings.HasPrefix(tok, "\u2014") {
		return ""
	}
	if shortDashWord.MatchString(tok) && !badLatinDash.MatchString(tok) {
		return tok
	}
	return ""
}

func isDigitsOnly(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func isRomanishDash(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch r {
		case 'I', 'V', 'X', 'І', 'Х', 'i', 'v', 'x', 'і', 'х':
		default:
			return false
		}
	}
	return true
}

func uniqueNonEmpty(ss []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range ss {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}
