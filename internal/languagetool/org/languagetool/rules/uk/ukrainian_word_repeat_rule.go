package uk

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// UkrainianWordRepeatRule ports org.languagetool.rules.uk.UkrainianWordRepeatRule.
// Ignore is POS-gated (date|time|number.*, abbr/initials); without tags, untagged
// doubles are ignored (Java returns true when no non-initial POS) — no surface invent
// of preposition lists.
type UkrainianWordRepeatRule struct {
	*rules.WordRepeatRule
}

var (
	// Java DATE_TIME_NUM_PATTERN = Pattern.compile("date|time|number.*") — full match on POS.
	ukDateTimeNum = regexp.MustCompile(`^(?:date|time|number.*)$`)
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

// ukIgnore ports UkrainianWordRepeatRule.ignore (does not call super — same as Java override).
func (r *UkrainianWordRepeatRule) ukIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if position < 0 || position >= len(tokens) || tokens[position] == nil {
		return true
	}
	atr := tokens[position]
	token := atr.GetToken()

	// від добра добра не шукають
	if position > 2 && token == "добра" &&
		strings.EqualFold(tokens[position-2].GetToken(), "від") {
		return true
	}

	// Тому що що?
	if position > 1 && token == "що" &&
		strings.EqualFold(tokens[position-2].GetToken(), "тому") {
		return true
	}

	// ні так, ні ні
	if position > 3 && token == "ні" &&
		tokens[position-2].GetToken() == "," &&
		strings.EqualFold(tokens[position-3].GetToken(), "так") {
		return true
	}

	if ukAllowed[strings.ToLower(token)] {
		return true
	}
	if ukAllowedCaps[token] {
		return true
	}

	// Java: PosTagHelper.hasPosTag(analyzedTokenReadings, DATE_TIME_NUM_PATTERN)
	if hasPosTagRE(atr, ukDateTimeNum) {
		return true
	}

	// Java: for each reading, if posTag != null && !isInitial && !SENT_END → return false
	// If no such reading (all null / initial / SENT_END) → return true
	for _, analyzedToken := range atr.GetReadings() {
		if analyzedToken == nil {
			continue
		}
		posTag := analyzedToken.GetPOSTag()
		if posTag == nil {
			continue
		}
		if *posTag == languagetool.SentenceEndTagName {
			continue
		}
		// Soft Analyze may attach SENT_END/PARA_END; skip those as non-morph POS
		if *posTag == languagetool.ParagraphEndTagName {
			continue
		}
		if !isInitialReading(analyzedToken, tokens, position) {
			return false
		}
	}
	return true
}

func isInitialReading(at *languagetool.AnalyzedToken, tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	// Java: posTag contains "abbr" OR (len==1 && uppercase && next is ".")
	pos := at.GetPOSTag()
	if pos != nil && strings.Contains(*pos, "abbr") {
		return true
	}
	tok := at.GetToken()
	runes := []rune(tok)
	if len(runes) == 1 && unicode.IsUpper(runes[0]) &&
		position+1 < len(tokens) && tokens[position+1] != nil &&
		tokens[position+1].GetToken() == "." {
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
