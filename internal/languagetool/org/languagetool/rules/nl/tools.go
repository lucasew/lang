package nl

import (
	"regexp"
	"strings"
	"unicode"
)

// Tools ports org.languagetool.rules.nl.Tools (compound gluing helpers).
// Pattern checks use full-string match semantics (Java Matcher.matches()).

var spelledWordsSet = func() map[string]struct{} {
	words := strings.Split("abc|adv|aed|apk|b2b|bh|bhv|bso|btw|bv|cao|cd|cfk|ckv|cv|dc|dj|dtp|dvd|fte|gft|ggo|ggz|gm|gmo|gps|gsm|hbo|"+
		"hd|hiv|hr|hrm|hst|ic|ivf|kmo|lcd|lp|lpg|lsd|mbo|mdf|mkb|mms|msn|mt|ngo|nv|ob|ov|ozb|p2p|pc|pcb|pdf|pk|pps|"+
		"pr|pvc|roc|rvs|sms|tbc|tbs|tl|tv|uv|vbo|vj|vmbo|vsbo|vwo|wc|wo|xtc|zzp", "|")
	m := make(map[string]struct{}, len(words))
	for _, w := range words {
		m[w] = struct{}{}
	}
	return m
}()

const spelledAlt = `abc|adv|aed|apk|b2b|bh|bhv|bso|btw|bv|cao|cd|cfk|ckv|cv|dc|dj|dtp|dvd|fte|gft|ggo|ggz|gm|gmo|gps|gsm|hbo|hd|hiv|hr|hrm|hst|ic|ivf|kmo|lcd|lp|lpg|lsd|mbo|mdf|mkb|mms|msn|mt|ngo|nv|ob|ov|ozb|p2p|pc|pcb|pdf|pk|pps|pr|pvc|roc|rvs|sms|tbc|tbs|tl|tv|uv|vbo|vj|vmbo|vsbo|vwo|wc|wo|xtc|zzp`

var (
	endsInDigit          = regexp.MustCompile(`^[0-9]*[0-9]$|.*[0-9]$`) // string ends with digit
	startsWithDigit      = regexp.MustCompile(`^[0-9]`)
	endsInHyphenAndChar  = regexp.MustCompile(`^.+-[a-z]$`)
	startsWithCharHyphen = regexp.MustCompile(`^[a-z]-.+$`)
	// Java: (^(^|.+-)?(spelled))$ via matches — whole string is optional xxx- + spelled acronym
	hyphenChars = regexp.MustCompile(`^(.+-)?(` + spelledAlt + `)$`)
	// Java: (spelled)(-.+|$)
	charsHyphen = regexp.MustCompile(`^(` + spelledAlt + `)(-.+)?$`)
)

var vowelPairs = []string{
	"aa", "ae", "ai", "ao", "au", "ee", "ei", "eu", "ée", "éi", "éu",
	"ie", "ii", "oe", "oi", "oo", "ou", "ui", "uu", "ij",
}

// GlueParts ports Tools.glueParts for Dutch compound suggestions.
func GlueParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	compound := parts[0]
	for i := 1; i < len(parts); i++ {
		word2 := parts[i]
		if len(compound) > 2 || spelledWord(compound) {
			runes1 := []rune(compound)
			runes2 := []rune(word2)
			lastChar := runes1[len(runes1)-1]
			firstChar := runes2[0]
			connection := string(lastChar) + string(firstChar)
			needHyphen := containsAny(connection, vowelPairs) ||
				(unicode.IsUpper(firstChar) && unicode.IsLower(lastChar)) ||
				(unicode.IsUpper(lastChar) && unicode.IsLower(firstChar)) ||
				(unicode.IsUpper(lastChar) && unicode.IsUpper(firstChar)) ||
				endsWithDigit(compound) ||
				startsWithDigit.MatchString(word2) ||
				hyphenChars.MatchString(compound) ||
				charsHyphen.MatchString(word2) ||
				endsInHyphenAndChar.MatchString(compound) ||
				startsWithCharHyphen.MatchString(word2)
			if needHyphen {
				compound = compound + "-" + word2
			} else {
				compound = compound + word2
			}
		} else {
			compound = compound + word2
		}
	}
	return compound
}

func endsWithDigit(s string) bool {
	if s == "" {
		return false
	}
	r := []rune(s)
	return unicode.IsDigit(r[len(r)-1])
}

func spelledWord(s string) bool {
	_, ok := spelledWordsSet[s]
	return ok
}

func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
