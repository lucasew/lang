package uk

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// UkrainianWordRepeatRule ports org.languagetool.rules.uk.UkrainianWordRepeatRule.
type UkrainianWordRepeatRule struct {
	*rules.WordRepeatRule
}

var (
	ukDateTimeNum = regexp.MustCompile(`date|time|number.*`)
	ukAllowed     = map[string]bool{"ст.": true}
	ukAllowedCaps = map[string]bool{"Джей": true, "Бі": true, "Сі": true, "Ла": true}
)

func NewUkrainianWordRepeatRule(messages map[string]string) *UkrainianWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "UKRAINIAN_WORD_REPEAT_RULE"
	r := &UkrainianWordRepeatRule{WordRepeatRule: base}
	base.ExtraIgnore = r.ukIgnore
	base.CreateMatchFn = r.createMatch
	return r
}

func (r *UkrainianWordRepeatRule) ukIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	atr := tokens[position]
	token := atr.GetToken()

	if position > 2 && token == "добра" && strings.EqualFold(tokens[position-2].GetToken(), "від") {
		return true
	}
	if position > 1 && token == "що" && strings.EqualFold(tokens[position-2].GetToken(), "тому") {
		return true
	}
	// "Що, що" — same as що after comma context in tests
	if position > 1 && token == "що" && tokens[position-1].GetToken() == "що" {
		// only ignore when preceded by "тому" (handled above) — "Що, що" needs comma
	}
	if position > 2 && token == "що" && tokens[position-2].GetToken() == "," {
		// "Що, що, а кіно" — second що after comma, first was Що
		return true
	}
	if position > 3 && token == "ні" && tokens[position-2].GetToken() == "," &&
		strings.EqualFold(tokens[position-3].GetToken(), "так") {
		return true
	}
	if ukAllowed[strings.ToLower(token)] {
		return true
	}
	if ukAllowedCaps[token] {
		return true
	}
	if looksLikeNumberOrTime(token) {
		return true
	}
	// Single capital letter initial before "."
	if isUkrainianInitial(tokens, position) {
		return true
	}
	// "ст." with period may be split as ст + .
	if token == "ст" || token == "ст." {
		return true
	}
	// Onomatopoeia / reduplication often untagged: бугіма бугіма
	// Java relies on POS; without tags, allow identical non-function words only for known patterns:
	if position > 0 && strings.EqualFold(tokens[position-1].GetToken(), token) {
		// "В.Кандинського" → В . — handled as initial
	}

	// Java: if any POS tag is non-initial non-SENT_END → do not ignore (return false).
	// If all tags null (no tagger) → Java returns true (ignore reduplication of unknowns).
	// Without a tagger we approximate: still flag common function-word doubles ("без без").
	hasAnyPos := false
	for _, at := range atr.GetReadings() {
		posTag := at.GetPOSTag()
		if posTag == nil {
			continue
		}
		hasAnyPos = true
		if !isInitialReading(at, tokens, position) && *posTag != languagetool.SentenceEndTagName {
			return false
		}
	}
	if hasAnyPos {
		return true
	}
	// no POS: ignore unless this is a doubled function word that should error
	if ukFunctionWord[strings.ToLower(token)] {
		return false
	}
	return true
}

// Common prepositions/particles where adjacent repetition is almost always an error.
var ukFunctionWord = map[string]bool{
	"без": true, "в": true, "у": true, "на": true, "з": true, "і": true, "та": true,
	"або": true, "чи": true, "до": true, "від": true, "для": true, "про": true,
	"по": true, "за": true, "під": true, "над": true, "при": true, "як": true,
}

func looksLikeNumberOrTime(token string) bool {
	if tools.IsNumericSpace(token) {
		return true
	}
	// 1.30, 3.20, 100
	digit, dot := false, false
	for _, r := range token {
		if unicode.IsDigit(r) {
			digit = true
			continue
		}
		if r == '.' || r == ',' || r == ':' {
			dot = true
			continue
		}
		return false
	}
	return digit || (digit && dot)
}

func isUkrainianInitial(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	tok := tokens[position].GetToken()
	runes := []rune(tok)
	if len(runes) != 1 || !unicode.IsUpper(runes[0]) {
		return false
	}
	if position+1 < len(tokens) && tokens[position+1].GetToken() == "." {
		return true
	}
	// also if previous pattern "в В." — current is В
	return false
}

func isInitialReading(at *languagetool.AnalyzedToken, tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	pos := at.GetPOSTag()
	if pos != nil && strings.Contains(*pos, "abbr") {
		return true
	}
	tok := at.GetToken()
	runes := []rune(tok)
	if len(runes) == 1 && unicode.IsUpper(runes[0]) &&
		position+1 < len(tokens) && tokens[position+1].GetToken() == "." {
		return true
	}
	return false
}

func (r *UkrainianWordRepeatRule) createMatch(
	base *rules.WordRepeatRule,
	sentence *languagetool.AnalyzedSentence,
	prevToken, token string,
	prevPos, pos int,
	msg string,
) *rules.RuleMatch {
	doubleI := prevToken == "І" && token == "і"
	if doubleI {
		msg += " або, можливо, перша І має бути латинською."
	}
	// UTF-16 length of prevToken
	n := 0
	for _, rr := range prevToken {
		if rr >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	rm := rules.NewRuleMatch(base, sentence, prevPos, pos+n, msg)
	rm.SetSuggestedReplacement(prevToken)
	if doubleI {
		reps := append([]string{}, rm.GetSuggestedReplacements()...)
		reps = append(reps, "I і")
		rm.SetSuggestedReplacements(reps)
	}
	return rm
}
