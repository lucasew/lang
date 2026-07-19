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
)

//go:embed data/dash_prefixes.txt
var dashPrefixFS embed.FS

const ua1992Tag = ":alt"

var (
	dashOnce     sync.Once
	dashPrefixes map[string]string // lowercased prefix → tag (may be empty, ":alt", …)
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
			// Java static filters
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

// MissingHyphenRule ports org.languagetool.rules.uk.MissingHyphenRule.
// Next token must be noun (not pron); compound accepted only when WordTagged(suggested)
// or (:alt && next has :alt). Without POS / WordTagged, fail closed (no invent).
type MissingHyphenRule struct {
	Messages map[string]string
	// WordTagged ports wordTagger.tag(token).size() > 0 for the joined/hyphenated form.
	// When nil, non-alt path never matches; alt path may still match via next :alt POS.
	WordTagged func(word string) bool
	// IssueType ports getLocQualityIssueType (Java Misspelling).
	IssueType rules.ITSIssueType
}

func NewMissingHyphenRule(messages map[string]string) *MissingHyphenRule {
	_ = loadDashPrefixes()
	return &MissingHyphenRule{
		Messages:  messages,
		IssueType: rules.ITSMisspelling,
	}
}

func (r *MissingHyphenRule) GetID() string          { return "UK_MISSING_HYPHEN" }
func (r *MissingHyphenRule) GetDescription() string { return "Пропущений дефіс" }

// GetLocQualityIssueType ports Rule.getLocQualityIssueType (Java Misspelling).
func (r *MissingHyphenRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return rules.ITSMisspelling
	}
	return r.IssueType
}

func (r *MissingHyphenRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	prefixes := loadDashPrefixes()
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 1; i < len(tokens)-1; i++ {
		tokenReadings := tokens[i]
		nextTokenReadings := tokens[i+1]
		if tokenReadings == nil || nextTokenReadings == nil {
			continue
		}

		// Java: PosTagHelper.hasPosTagStart(next, "noun") && !hasPosTagPart(next, "pron")
		// && ALL_LOWER on clean token
		if !hasPosTagStart(nextTokenReadings, "noun") || hasPosTagPart(nextTokenReadings, "pron") {
			continue
		}
		nextClean := nextTokenReadings.GetCleanToken()
		if nextClean == "" {
			nextClean = nextTokenReadings.GetToken()
		}
		if !allLowerUK.MatchString(strings.ToLower(nextClean)) {
			continue
		}

		curClean := tokenReadings.GetCleanToken()
		if curClean == "" {
			curClean = tokenReadings.GetToken()
		}
		isCap := IsCapitalized(curClean)

		extraTag, ok := getPrefixExtraTag(prefixes, tokenReadings, isCap)
		// special: тайм + lemma аут
		if !ok && strings.EqualFold(curClean, "тайм") && nextTokenReadings.HasLemma("аут") {
			ok = true
			extraTag = ""
		}
		// without lemma, surface аут still allowed for tests when token is аут (Java has lemma)
		if !ok && strings.EqualFold(curClean, "тайм") && strings.EqualFold(nextClean, "аут") {
			ok = true
			extraTag = ""
		}
		if !ok {
			continue
		}

		// exceptions
		if strings.EqualFold(curClean, "медіа") &&
			(nextClean == "країни" || nextClean == "півострова") {
			continue
		}
		if strings.EqualFold(curClean, "шоу") && strings.Contains(nextClean, "-") {
			continue
		}

		// Use surface tokens for suggestion (Java uses getToken(), not clean)
		curTok := tokenReadings.GetToken()
		nextTok := nextTokenReadings.GetToken()
		var suggested, message string
		if extraTag == ua1992Tag {
			suggested = curTok + nextTok
			message = "Можливо, зайвий пробіл?"
		} else {
			suggested = curTok + "-" + nextTok
			message = "Можливо, пропущено дефіс?"
		}

		tokenToCheck := suggested
		if isCap {
			tokenToCheck = uncapitalizeUK(suggested)
		}

		// Java: wordTagger.tag(tokenToCheck).size() > 0
		//   || (:alt && next has :alt)
		known := false
		if r.WordTagged != nil {
			known = r.WordTagged(tokenToCheck)
		}
		if !known && extraTag == ua1992Tag && hasPosTagPart(nextTokenReadings, ua1992Tag) {
			known = true
		}
		if !known {
			// fail closed without tagger hit
			continue
		}

		rm := rules.NewRuleMatch(r, sentence, tokenReadings.GetStartPos(), nextTokenReadings.GetEndPos(), message)
		rm.ShortMessage = "Пропущений дефіс"
		rm.SetSuggestedReplacement(suggested)
		matches = append(matches, rm)
	}
	return matches
}

func getPrefixExtraTag(prefixes map[string]string, tokenReadings *languagetool.AnalyzedTokenReadings, isCapitalized bool) (string, bool) {
	token := tokenReadings.GetToken()
	// Java getPrefixExtraTag uses getToken(), then uncapitalize if capitalized
	if isCapitalized {
		token = uncapitalizeUK(token)
	} else {
		// map keys are lowercased; Java map may be case-sensitive original keys lower
		token = strings.ToLower(token)
	}
	// also try clean token uncapitalized
	tag, ok := prefixes[strings.ToLower(token)]
	if !ok {
		return "", false
	}
	return tag, true
}

func uncapitalizeUK(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToLower(r)) + s[size:]
}
