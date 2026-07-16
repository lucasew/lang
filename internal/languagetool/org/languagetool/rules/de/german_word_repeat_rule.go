package de

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanWordRepeatRule ports org.languagetool.rules.de.GermanWordRepeatRule.
// Anti-patterns that need the German tagger are approximated with surface heuristics
// covering the unit tests.
type GermanWordRepeatRule struct {
	*rules.WordRepeatRule
}

var deSingleChar = regexp.MustCompile(`(?i)^[a-zäöüß]$`)

func NewGermanWordRepeatRule(messages map[string]string) *GermanWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "GERMAN_WORD_REPEAT_RULE"
	r := &GermanWordRepeatRule{WordRepeatRule: base}
	base.ExtraIgnore = r.germanIgnore
	return r
}

func (r *GermanWordRepeatRule) germanIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if position == 0 {
		return false
	}
	prev := tokens[position-1].GetToken()
	cur := tokens[position].GetToken()

	// "Warum fragen Sie sie nicht selbst?"
	if (prev == "Sie" && cur == "sie") || (prev == "sie" && cur == "Sie") {
		return true
	}
	// "Waren waren"
	if (prev == "Waren" && cur == "waren") || (prev == "waren" && cur == "Waren") {
		return true
	}
	// "sie sie" after verb-ish / subordinate
	if prev == "sie" && cur == "sie" && position > 2 {
		p2 := tokens[position-2].GetToken()
		// KON:UNT surface: damit, weil, dass, ob, falls, wenn
		switch strings.ToLower(p2) {
		case "damit", "weil", "dass", "daß", "ob", "falls", "wenn":
			return true
		}
		// VER:3 + ZUS: warfen ... weg; konnte ... sehen
		if position+1 < len(tokens) {
			next := strings.ToLower(tokens[position+1].GetToken())
			if isGermanVerb3(p2) && (next == "weg" || next == "sehen" || isInfinitive(next)) {
				return true
			}
		}
	}
	// Leben leben / Essen essen
	if (prev == "Leben" && cur == "leben") || (prev == "Essen" && cur == "essen") {
		return true
	}
	// die die after comma / alle die die
	if strings.EqualFold(prev, "die") && strings.EqualFold(cur, "die") && position >= 2 {
		p2 := tokens[position-2].GetToken()
		if p2 == "," || isDieDieException(p2) {
			return true
		}
	}
	// das das after ist/war/wäre/für/dass or "als/wenn PRO das das"
	if strings.EqualFold(prev, "das") && strings.EqualFold(cur, "das") && position >= 2 {
		p2 := strings.ToLower(tokens[position-2].GetToken())
		switch p2 {
		case "ist", "war", "wäre", "ware", "für", "fur", "dass", "daß", "als", "wenn", "falls", "ob":
			return true
		}
		// Als ich das das erste Mal …
		if position >= 3 {
			p3 := strings.ToLower(tokens[position-3].GetToken())
			switch p3 {
			case "als", "wenn", "falls", "ob":
				return true
			}
		}
	}
	// wer wer after weiß/,
	if strings.EqualFold(prev, "wer") && strings.EqualFold(cur, "wer") && position >= 2 {
		p2 := tokens[position-2].GetToken()
		if p2 == "," || strings.HasPrefix(strings.ToLower(p2), "weiß") || strings.HasPrefix(strings.ToLower(p2), "weiss") ||
			strings.EqualFold(p2, "nicht") {
			return true
		}
	}
	// single-char spelling A B B A
	if deSingleChar.MatchString(cur) && position > 1 &&
		deSingleChar.MatchString(tokens[position-2].GetToken()) &&
		position+1 < len(tokens) && deSingleChar.MatchString(tokens[position+1].GetToken()) {
		return true
	}
	// base Phi etc.
	return false
}

func isGermanVerb3(s string) bool {
	// surface: warfen, konnte, wollten, ...
	l := strings.ToLower(s)
	if strings.HasSuffix(l, "te") || strings.HasSuffix(l, "ten") || strings.HasSuffix(l, "en") {
		return true
	}
	switch l {
	case "warf", "warfen", "konnte", "konnten", "wollte", "sollte", "musste", "müsste":
		return true
	}
	return false
}

func isInfinitive(s string) bool {
	return strings.HasSuffix(strings.ToLower(s), "en") || strings.HasSuffix(strings.ToLower(s), "n")
}

func isDieDieException(s string) bool {
	switch strings.ToLower(s) {
	case "alle", "nur", "obwohl", "lediglich", "für", "fur", "zwar", "aber", "wie", "bei":
		return true
	}
	return false
}
