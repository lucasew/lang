package uk

import (
	"regexp"
	"strings"

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
		// Path A: short dash inside word (Java shortDashToken on last reading surface)
		if tok := shortDashToken(tokens[i]); tok != "" {
			rm := rules.NewRuleMatch(r, sentence, tokens[i].GetStartPos(), tokens[i].GetEndPos(), msg)
			rm.ShortMessage = "Коротка риска"
			rm.SetSuggestedReplacements([]string{
				dashChars.ReplaceAllString(tok, "-"),
				dashChars.ReplaceAllString(tok, " \u2014 "),
			})
			ruleMatches = append(ruleMatches, rm)
			continue
		}

		// Path B: bare en/em dash with missing spaces (Java)
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
		// Java isNumber = hasPosTagStart(..., "number")
		if i > 1 && typographyIsNumber(tokens[i-1]) && i < len(tokens)-1 && typographyIsNumber(tokens[i+1]) {
			continue
		}

		// Tokenizer may split a single Java token "word–word" into three tokens.
		// Only treat fully-glued spans as path A when recombined form passes shortDashWord
		// (so А–Т / roman ranges stay silent like Java one-token shortDashToken).
		if noSpaceLeft && noSpaceRight && i > 1 && i < len(tokens)-1 {
			combined := tokens[i-1].GetToken() + t + tokens[i+1].GetToken()
			if shortDashWord.MatchString(combined) && !badLatinDash.MatchString(combined) {
				rm := rules.NewRuleMatch(r, sentence, tokens[i-1].GetStartPos(), tokens[i+1].GetEndPos(), msg)
				rm.ShortMessage = "Коротка риска"
				rm.SetSuggestedReplacements([]string{
					dashChars.ReplaceAllString(combined, "-"),
					dashChars.ReplaceAllString(combined, " \u2014 "),
				})
				ruleMatches = append(ruleMatches, rm)
				i++
				continue
			}
			// glued but not a short-dash word → no match (Java single-token path would also reject)
			continue
		}

		var replacements []string
		// Java: both sides contain Cyrillic → suggest left-right hyphen compound
		if i > 1 && i < len(tokens)-1 &&
			typoCyrRE.MatchString(tokens[i-1].GetToken()) &&
			typoCyrRE.MatchString(tokens[i+1].GetToken()) {
			replacements = append(replacements, tokens[i-1].GetToken()+"-"+tokens[i+1].GetToken())
		}

		// Java: startPos/endPos from prev/next; repl = [prev ]—[ next]
		startPos := tokens[i].GetStartPos()
		endPos := tokens[i].GetEndPos()
		var b strings.Builder
		if i > 1 {
			b.WriteString(tokens[i-1].GetToken())
			b.WriteByte(' ')
			startPos = tokens[i-1].GetStartPos()
		}
		b.WriteString("\u2014")
		if i < len(tokens)-1 {
			b.WriteByte(' ')
			b.WriteString(tokens[i+1].GetToken())
			// Java uses next startPos as endPos (not full end of next token)
			endPos = tokens[i+1].GetStartPos()
		} else {
			endPos = tokens[i].GetEndPos()
		}
		replacements = append(replacements, b.String())
		replacements = uniqueNonEmpty(replacements)

		rm := rules.NewRuleMatch(r, sentence, startPos, endPos, msg)
		rm.ShortMessage = "Коротка риска"
		rm.SetSuggestedReplacements(replacements)
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}

// shortDashToken ports TypographyRule.shortDashToken (last reading token surface).
func shortDashToken(atr *languagetool.AnalyzedTokenReadings) string {
	if atr == nil {
		return ""
	}
	rds := atr.GetReadings()
	if len(rds) == 0 {
		return ""
	}
	last := rds[len(rds)-1]
	if last == nil {
		return ""
	}
	tok := last.GetToken()
	if tok == "" {
		tok = atr.GetToken()
	}
	if tok == "" {
		return ""
	}
	if !strings.ContainsRune(tok, '\u2013') && !strings.ContainsRune(tok, '\u2014') {
		return ""
	}
	// dash not at start (indexOf > 0)
	if strings.IndexRune(tok, '\u2013') == 0 || strings.IndexRune(tok, '\u2014') == 0 {
		return ""
	}
	if shortDashWord.MatchString(tok) && !badLatinDash.MatchString(tok) {
		return tok
	}
	return ""
}

// typographyIsNumber ports TypographyRule.isNumber (POS starts with "number").
func typographyIsNumber(atr *languagetool.AnalyzedTokenReadings) bool {
	return HasPosTagStart(atr, "number")
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
