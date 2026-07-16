package uk

import (
	"bufio"
	"embed"
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/dash_prefixes.txt
var dashPrefixFS embed.FS

const ua1992Tag = ":alt"

var (
	dashOnce     sync.Once
	dashPrefixes map[string]string // lowercased prefix → tag (may be empty, ":alt", ":bad", ":slang")
	allLowerUK   = regexp.MustCompile(`^[а-яіїєґ'\-]+$`)
)

func loadDashPrefixes() map[string]string {
	dashOnce.Do(func() {
		f, err := dashPrefixFS.Open("data/dash_prefixes.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m := map[string]string{}
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || line[0] == '#' {
				continue
			}
			// strip trailing comments after #
			if i := strings.IndexByte(line, '#'); i >= 0 {
				line = strings.TrimSpace(line[:i])
			}
			parts := strings.Fields(line)
			if len(parts) == 0 {
				continue
			}
			key := parts[0]
			tag := ""
			if len(parts) > 1 {
				tag = parts[1]
			}
			// Java filters
			if key == "блок" || key == "рейтинг" {
				continue
			}
			if !allLowerUK.MatchString(strings.ToLower(key)) || strings.Contains(tag, ":bad") {
				continue
			}
			m[strings.ToLower(key)] = tag
		}
		// ensure special-case тайм for "тайм аут"
		if _, ok := m["тайм"]; !ok {
			m["тайм"] = ""
		}
		dashPrefixes = m
	})
	return dashPrefixes
}

// MissingHyphenRule ports org.languagetool.rules.uk.MissingHyphenRule without a word tagger
// (always suggests when prefix matches; Java only when compound is known to the dictionary).
type MissingHyphenRule struct {
	Messages map[string]string
}

func NewMissingHyphenRule(messages map[string]string) *MissingHyphenRule {
	_ = loadDashPrefixes()
	return &MissingHyphenRule{Messages: messages}
}

func (r *MissingHyphenRule) GetID() string { return "UK_MISSING_HYPHEN" }

func isCapitalizedUK(s string) bool {
	rs := []rune(s)
	if len(rs) < 2 {
		return false
	}
	return unicode.IsUpper(rs[0]) && unicode.IsLower(rs[1])
}

func uncapitalize(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToLower(r)) + s[size:]
}

func (r *MissingHyphenRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	prefixes := loadDashPrefixes()
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 1; i < len(tokens)-1; i++ {
		cur := tokens[i]
		next := tokens[i+1]
		nextTok := next.GetToken()
		if !allLowerUK.MatchString(strings.ToLower(nextTok)) {
			continue
		}
		isCap := isCapitalizedUK(cur.GetToken())
		key := cur.GetToken()
		if isCap {
			key = uncapitalize(key)
		} else {
			key = strings.ToLower(key)
		}
		extraTag, ok := prefixes[key]
		// special: тайм + аут
		if !ok && strings.EqualFold(cur.GetToken(), "тайм") && strings.EqualFold(nextTok, "аут") {
			ok = true
			extraTag = ""
		}
		if !ok {
			continue
		}
		// exceptions
		if strings.EqualFold(cur.GetToken(), "медіа") &&
			(nextTok == "країни" || nextTok == "півострова") {
			continue
		}
		if strings.EqualFold(cur.GetToken(), "шоу") && strings.Contains(nextTok, "-") {
			continue
		}
		var suggested, message string
		if extraTag == ua1992Tag {
			suggested = cur.GetToken() + next.GetToken()
			message = "Можливо, зайвий пробіл?"
		} else {
			suggested = cur.GetToken() + "-" + next.GetToken()
			message = "Можливо, пропущено дефіс?"
		}
		if isCap {
			// keep capitalization of first part as in original token
			_ = tools.UppercaseFirstChar
		}
		rm := rules.NewRuleMatch(r, sentence, cur.GetStartPos(), next.GetEndPos(), message)
		rm.ShortMessage = "Пропущений дефіс"
		rm.SetSuggestedReplacement(suggested)
		matches = append(matches, rm)
	}
	return matches
}
