package uk

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	taguk "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/uk"
)

// UkrainianWordRepeatRule ports org.languagetool.rules.uk.UkrainianWordRepeatRule.
// Extends WordRepeatRule; ignore() is fully overridden (does not call super.ignore).
type UkrainianWordRepeatRule struct {
	*rules.WordRepeatRule
}

var (
	// Java: Pattern.compile("date|time|number.*") — Matcher.matches() full tag.
	ukDateTimeNumPattern = regexp.MustCompile(`^(?:date|time|number.*)$`)

	// Java REPEAT_ALLOWED_SET
	ukRepeatAllowedSet = map[string]bool{
		"ст.": true,
	}
	// Java REPEAT_ALLOWED_CAPS_SET (case-sensitive contains)
	ukRepeatAllowedCapsSet = map[string]bool{
		"Джей": true,
		"Бі":   true,
		"Сі":   true,
		"Ла":   true,
	}
)

// NewUkrainianWordRepeatRule ports UkrainianWordRepeatRule(ResourceBundle, Language).
// Language is unused by WordRepeatRule beyond the Java super ctor signature.
func NewUkrainianWordRepeatRule(messages map[string]string) *UkrainianWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "UKRAINIAN_WORD_REPEAT_RULE"
	r := &UkrainianWordRepeatRule{WordRepeatRule: base}
	// Java: ignore() fully replaces WordRepeatRule.ignore (no super.ignore).
	// ExtraIgnore true short-circuits; when false, base name list still runs
	// (Phi/Li/…) — not present in UkrainianWordRepeatRuleTest.
	base.ExtraIgnore = r.ukIgnore
	base.CreateMatchFn = r.createMatch
	return r
}

// ukIgnore ports UkrainianWordRepeatRule.ignore (does not call super).
func (r *UkrainianWordRepeatRule) ukIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if position < 0 || position >= len(tokens) || tokens[position] == nil {
		// Java would NPE; fail closed for Go nil safety.
		return true
	}
	analyzedTokenReadings := tokens[position]
	token := analyzedTokenReadings.GetToken()

	// від добра добра не шукають
	if position > 2 &&
		token == "добра" &&
		tokens[position-2] != nil &&
		strings.EqualFold(tokens[position-2].GetToken(), "від") {
		return true
	}

	// Тому що що?
	if position > 1 &&
		token == "що" &&
		tokens[position-2] != nil &&
		strings.EqualFold(tokens[position-2].GetToken(), "тому") {
		return true
	}

	// ні так, ні ні
	if position > 3 &&
		token == "ні" &&
		tokens[position-2] != nil &&
		tokens[position-2].GetToken() == "," &&
		tokens[position-3] != nil &&
		strings.EqualFold(tokens[position-3].GetToken(), "так") {
		return true
	}

	// Java: REPEAT_ALLOWED_SET.contains(token.toLowerCase())
	if ukRepeatAllowedSet[strings.ToLower(token)] {
		return true
	}
	// Java: REPEAT_ALLOWED_CAPS_SET.contains(token)
	if ukRepeatAllowedCapsSet[token] {
		return true
	}

	// Java: PosTagHelper.hasPosTag(analyzedTokenReadings, DATE_TIME_NUM_PATTERN)
	if taguk.HasPosTagReadings(analyzedTokenReadings, ukDateTimeNumPattern) {
		return true
	}

	// Java: for each reading with non-null POS, if !isInitial && !SENTENCE_END → false
	for _, analyzedToken := range analyzedTokenReadings.GetReadings() {
		if analyzedToken == nil {
			continue
		}
		posTag := analyzedToken.GetPOSTag()
		if posTag == nil {
			continue
		}
		// Java commented-out SENTENCE_START check is not applied.
		if *posTag == languagetool.SentenceEndTagName {
			continue
		}
		if !isInitial(analyzedToken, tokens, position) {
			return false
		}
	}
	return true
}

// isInitial ports UkrainianWordRepeatRule.isInitial.
// POS is non-null when called from ignore()'s loop.
func isInitial(analyzedToken *languagetool.AnalyzedToken, tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	posTag := analyzedToken.GetPOSTag()
	if posTag != nil && strings.Contains(*posTag, taguk.IPOSAbbr.Text()) {
		return true
	}
	tok := analyzedToken.GetToken()
	// Java: token.length() == 1 (UTF-16) && Character.isUpperCase(charAt(0))
	runes := []rune(tok)
	if len(runes) == 1 && unicode.IsUpper(runes[0]) &&
		position < len(tokens)-1 &&
		tokens[position+1] != nil &&
		tokens[position+1].GetToken() == "." {
		return true
	}
	return false
}

// createMatch ports UkrainianWordRepeatRule.createRuleMatch.
func (r *UkrainianWordRepeatRule) createMatch(
	base *rules.WordRepeatRule,
	sentence *languagetool.AnalyzedSentence,
	prevToken, token string,
	prevPos, pos int,
	msg string,
) *rules.RuleMatch {
	// Java: prevToken.equals("І") && token.equals("і") — exact case
	doubleI := prevToken == "І" && token == "і"
	if doubleI {
		msg += " або, можливо, перша І має бути латинською."
	}
	// Java: super.createRuleMatch(...)
	ruleMatch := base.CreateRuleMatch(prevToken, token, prevPos, pos, msg, sentence)
	if doubleI {
		replacements := append([]string{}, ruleMatch.GetSuggestedReplacements()...)
		// Latin capital I (U+0049), not Cyrillic І (U+0406)
		replacements = append(replacements, "I і")
		ruleMatch.SetSuggestedReplacements(replacements)
	}
	return ruleMatch
}
