package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// TypographyRule ports org.languagetool.rules.uk.TypographyRule.
// Java: setCategory(Categories.TYPOGRAPHY.getCategory(messages)).
type TypographyRule struct {
	Messages map[string]string
	Category *rules.Category
}

// NewTypographyRule ports TypographyRule(ResourceBundle messages).
func NewTypographyRule(messages map[string]string) *TypographyRule {
	return &TypographyRule{
		Messages: messages,
		Category: rules.CatTypography.GetCategory(messages),
	}
}

// GetID ports TypographyRule.getId → "DASH".
func (r *TypographyRule) GetID() string { return "DASH" }

// GetDescription ports TypographyRule.getDescription.
func (r *TypographyRule) GetDescription() string {
	return "Коротка риска замість дефісу"
}

// GetShort ports TypographyRule.getShort.
func (r *TypographyRule) GetShort() string {
	return "Коротка риска"
}

// GetCategory ports Rule.getCategory (Java TYPOGRAPHY).
func (r *TypographyRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// Java TypographyRule.PATTERN = Pattern.compile(".*[а-яїієґА-ЯІЇЄҐ].*")
var typographyCyrillicPattern = regexp.MustCompile(`.*[а-яїієґА-ЯІЇЄҐ].*`)

// Java SHORT_DASH_WORD: CASE_INSENSITIVE|UNICODE_CASE
var shortDashWord = regexp.MustCompile(`(?i)[а-яіїєґ']{2,}([\x{2013}\x{2014}][а-яіїєґ']{2,})+`)

// Java BAD_LATIN (case-sensitive)
var badLatinDash = regexp.MustCompile(`[ХІXIV]+[\x{2013}\x{2014}][ХІXIV]+`)

// Java replaceAll("[\u2013\u2014]", ...)
var enEmDashChars = regexp.MustCompile(`[\x{2013}\x{2014}]`)

// Match ports TypographyRule.match.
func (r *TypographyRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	var ruleMatches []*rules.RuleMatch
	if sentence == nil {
		return ruleMatches
	}
	tokens := sentence.GetTokensWithoutWhitespace()

	for i := 1; i < len(tokens); i++ {
		// Java shortDashToken path
		if shortDashTok := shortDashToken(tokens[i]); shortDashTok != "" {
			replacements := []string{
				enEmDashChars.ReplaceAllString(shortDashTok, "-"),
				enEmDashChars.ReplaceAllString(shortDashTok, " \u2014 "),
			}
			// Правопис 2019 § 161. ТИРЕ (—), пп. 12-15
			msg := "Риска всередині слова. Всередині слова вживайте дефіс, між словами виокремлюйте риску пробілами."
			ruleMatches = append(ruleMatches, r.createRuleMatch(tokens[i], replacements, msg, sentence))
			continue
		}

		// Java: bare en/em dash token with missing spaces
		t := tokens[i].GetToken()
		if t != "\u2014" && t != "\u2013" {
			continue
		}
		noSpaceLeft := i > 1 && !tokens[i].IsWhitespaceBefore() &&
			tokens[i-1].GetToken() != "," &&
			tokens[i-1].GetToken() != "«"
		noSpaceRight := i < len(tokens)-1 && !tokens[i+1].IsWhitespaceBefore() &&
			tokens[i+1].GetToken() != ">"

		if !(noSpaceLeft || noSpaceRight) {
			continue
		}

		// Java: if both sides isNumber → continue
		if i > 1 && typographyIsNumber(tokens[i-1]) && i < len(tokens)-1 && typographyIsNumber(tokens[i+1]) {
			continue
		}

		var replacements []string

		// Java: both sides match PATTERN → left-right hyphen compound
		if i > 1 && i < len(tokens)-1 &&
			typographyCyrillicPattern.MatchString(tokens[i-1].GetToken()) &&
			typographyCyrillicPattern.MatchString(tokens[i+1].GetToken()) {
			replacements = append(replacements, tokens[i-1].GetToken()+"-"+tokens[i+1].GetToken())
		}

		startPos := tokens[i].GetStartPos()
		endPos := tokens[i].GetEndPos()
		repl := ""
		if i > 1 {
			repl = tokens[i-1].GetToken() + " "
			startPos = tokens[i-1].GetStartPos()
		}
		repl += "\u2014"
		if i < len(tokens)-1 {
			repl += " " + tokens[i+1].GetToken()
			// Java: endPos = tokens[i+1].getStartPos()
			endPos = tokens[i+1].GetStartPos()
		} else {
			endPos = tokens[i].GetEndPos()
		}
		replacements = append(replacements, repl)

		// Правопис 2019 § 161. ТИРЕ (—), пп. 12-15
		msg := "Риска всередині слова. Всередині слова вживайте дефіс, між словами виокремлюйте риску пробілами."
		// Java: new RuleMatch(this, sentence, startPos, endPos, msg, getShort())
		potential := rules.NewRuleMatch(r, sentence, startPos, endPos, msg)
		potential.SetShortMessage(r.GetShort())
		potential.SetSuggestedReplacements(replacements)
		ruleMatches = append(ruleMatches, potential)
	}

	return ruleMatches
}

// shortDashToken ports TypographyRule.shortDashToken.
func shortDashToken(tokenReadings *languagetool.AnalyzedTokenReadings) string {
	if tokenReadings == nil {
		return ""
	}
	readings := tokenReadings.GetReadings()
	// Java: readings.size() == 0 || tokenReadings.getToken() == null
	if len(readings) == 0 {
		return ""
	}
	// Go has no nil string; surface is always set by analysis. Proceed to last reading.
	last := readings[len(readings)-1]
	if last == nil {
		return ""
	}
	lastReadingToken := last.GetToken()
	if lastReadingToken == "" {
		return ""
	}
	// Java: indexOf('\u2013') > 0 || indexOf('\u2014') > 0
	if !(strings.IndexRune(lastReadingToken, '\u2013') > 0 || strings.IndexRune(lastReadingToken, '\u2014') > 0) {
		return ""
	}
	if shortDashWord.MatchString(lastReadingToken) && !badLatinDash.MatchString(lastReadingToken) {
		return lastReadingToken
	}
	return ""
}

// typographyIsNumber ports TypographyRule.isNumber (PosTagHelper.hasPosTagStart "number").
func typographyIsNumber(analyzedTokenReadings *languagetool.AnalyzedTokenReadings) bool {
	return HasPosTagStart(analyzedTokenReadings, "number")
}

// createRuleMatch ports TypographyRule.createRuleMatch.
func (r *TypographyRule) createRuleMatch(readings *languagetool.AnalyzedTokenReadings, replacements []string, msg string, sentence *languagetool.AnalyzedSentence) *rules.RuleMatch {
	potential := rules.NewRuleMatch(r, sentence, readings.GetStartPos(), readings.GetEndPos(), msg)
	potential.SetShortMessage(r.GetShort())
	potential.SetSuggestedReplacements(replacements)
	return potential
}
