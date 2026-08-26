package uk

import (
	"embed"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// dashPrefixFS is the official dash_prefixes.txt (byte-identical to
// inspiration/.../resource/uk/dash_prefixes.txt). Used as fallback when the
// inspiration path is not discoverable; speller also reads this embed.
//
//go:embed data/dash_prefixes.txt
var dashPrefixFS embed.FS

// UA_1992_TAG_PART ports MissingHyphenRule.UA_1992_TAG_PART.
const ua1992TagPart = ":alt"

var (
	dashOnce     sync.Once
	dashPrefixes map[string]string // key → extra tag ("" or ":alt", …); Java static map after filters
	// ALL_LOWER ports MissingHyphenRule.ALL_LOWER = [а-яіїєґ'-]+ (Matcher.matches full string).
	allLowerUK = regexp.MustCompile(`^[а-яіїєґ'\-]+$`)
)

// loadDashPrefixes ports static dashPrefixes = ExtraDictionaryLoader.loadMap("/uk/dash_prefixes.txt")
// plus static filters (remove блок/рейтинг; ALL_LOWER keys only; strip values containing :bad).
func loadDashPrefixes() map[string]string {
	dashOnce.Do(func() {
		m := map[string]string{}
		if p := discoverOfficialDashPrefixes(); p != "" {
			if f, err := os.Open(p); err == nil {
				loaded, err := LoadMap(f)
				_ = f.Close()
				if err == nil && loaded != nil {
					m = loaded
				}
			}
		}
		if len(m) == 0 {
			// Fallback: embedded official resource (byte-identical).
			if f, err := dashPrefixFS.Open("data/dash_prefixes.txt"); err == nil {
				loaded, err := LoadMap(f)
				_ = f.Close()
				if err == nil && loaded != nil {
					m = loaded
				}
			}
		}
		// Java static:
		//   dashPrefixes.remove("блок");
		//   dashPrefixes.remove("рейтинг");
		//   removeIf !ALL_LOWER.matcher(key).matches() || value.contains(":bad")
		delete(m, "блок")
		delete(m, "рейтинг")
		for k, v := range m {
			if !allLowerUK.MatchString(k) || strings.Contains(v, ":bad") {
				delete(m, k)
			}
		}
		dashPrefixes = m
	})
	return dashPrefixes
}

// discoverOfficialDashPrefixes finds Java resource /uk/dash_prefixes.txt under the
// inspiration tree (same pattern as tagging/uk discoverUKResource).
func discoverOfficialDashPrefixes() string {
	const rel = "inspiration/languagetool/languagetool-language-modules/uk/src/main/resources/org/languagetool/resource/uk/dash_prefixes.txt"
	if _, file, _, ok := runtime.Caller(0); ok {
		root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../"))
		p := filepath.Join(root, rel)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		cand := filepath.Join(dir, rel)
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// MissingHyphenRule ports org.languagetool.rules.uk.MissingHyphenRule.
type MissingHyphenRule struct {
	Messages map[string]string
	// WordTagger ports the Java WordTagger (UkrainianTagger.getWordTagger()).
	// When nil, only the :alt + next-:alt branch can accept a match.
	WordTagger tagging.WordTagger
	// IssueType ports getLocQualityIssueType (Java Misspelling).
	IssueType rules.ITSIssueType
}

// NewMissingHyphenRule ports MissingHyphenRule(ResourceBundle, WordTagger).
func NewMissingHyphenRule(messages map[string]string, wordTagger tagging.WordTagger) *MissingHyphenRule {
	_ = loadDashPrefixes()
	return &MissingHyphenRule{
		Messages:   messages,
		WordTagger: wordTagger,
		IssueType:  rules.ITSMisspelling,
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

// Match ports MissingHyphenRule.match.
func (r *MissingHyphenRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	prefixes := loadDashPrefixes()
	tokens := sentence.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch

	// Java: for (int i = 1; i < tokens.length-1; i++)
	for i := 1; i < len(tokens)-1; i++ {
		tokenReadings := tokens[i]
		nextTokenReadings := tokens[i+1]
		if tokenReadings == nil || nextTokenReadings == nil {
			continue
		}

		// PosTagHelper.hasPosTagStart(next, "noun")
		// && !PosTagHelper.hasPosTagPart(next, "pron")
		// && ALL_LOWER.matcher(next.getCleanToken()).matches()
		if !hasPosTagStart(nextTokenReadings, "noun") || hasPosTagPart(nextTokenReadings, "pron") {
			continue
		}
		nextClean := nextTokenReadings.GetCleanToken()
		if nextClean == "" {
			// Java NPE if clean is null; GetCleanToken empty → treat as fail closed
			continue
		}
		if !allLowerUK.MatchString(nextClean) {
			continue
		}

		curClean := tokenReadings.GetCleanToken()
		isCapitalized := IsCapitalized(curClean)

		extraTag, hasPrefix := getPrefixExtraTag(prefixes, tokenReadings, isCapitalized)
		// Java: extraTag != null || (clean.toLowerCase().equals("тайм") && hasLemma(next, "аут"))
		if !hasPrefix {
			if strings.ToLower(curClean) == "тайм" && HasLemmaToken(nextTokenReadings, "аут") {
				hasPrefix = true
				extraTag = ""
			}
		}
		if !hasPrefix {
			continue
		}

		// всі медіа країни|півострова
		if strings.EqualFold(curClean, "медіа") &&
			(nextClean == "країни" || nextClean == "півострова") {
			continue
		}
		// до шоу поп-діви (ASCII hyphen in next clean token)
		if strings.EqualFold(curClean, "шоу") && strings.Contains(nextClean, "-") {
			continue
		}

		// suggestions use getToken() surfaces (not clean)
		curTok := tokenReadings.GetToken()
		nextTok := nextTokenReadings.GetToken()
		var suggested, message string
		if extraTag == ua1992TagPart {
			suggested = curTok + nextTok
			message = "Можливо, зайвий пробіл?"
		} else {
			suggested = curTok + "-" + nextTok
			message = "Можливо, пропущено дефіс?"
		}

		// tokenToCheck = isCapitalized ? uncapitalize(suggested) : suggested
		tokenToCheck := suggested
		if isCapitalized {
			tokenToCheck = uncapitalizeUK(suggested)
		}

		// wordTagger.tag(tokenToCheck).size() > 0
		//   || (:alt && next has :alt)
		known := false
		if r.WordTagger != nil && len(r.WordTagger.Tag(tokenToCheck)) > 0 {
			known = true
		}
		if !known && extraTag == ua1992TagPart && hasPosTagPart(nextTokenReadings, ua1992TagPart) {
			known = true
		}
		if !known {
			continue
		}

		// new RuleMatch(this, sentence, start, end, message, getDescription())
		rm := rules.NewRuleMatch(r, sentence, tokenReadings.GetStartPos(), nextTokenReadings.GetEndPos(), message)
		rm.ShortMessage = r.GetDescription()
		rm.SetSuggestedReplacement(suggested)
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}

// getPrefixExtraTag ports MissingHyphenRule.getPrefixExtraTag.
// Java: token = getToken(); if capitalized → StringUtils.uncapitalize(token); return map.get(token).
func getPrefixExtraTag(prefixes map[string]string, tokenReadings *languagetool.AnalyzedTokenReadings, isCapitalized bool) (string, bool) {
	if tokenReadings == nil || prefixes == nil {
		return "", false
	}
	token := tokenReadings.GetToken()
	if isCapitalized {
		token = uncapitalizeUK(token)
	}
	tag, ok := prefixes[token]
	return tag, ok
}

// uncapitalizeUK ports org.apache.commons.lang3.StringUtils.uncapitalize (first char only).
func uncapitalizeUK(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size == 1 {
		return s
	}
	return string(unicode.ToLower(r)) + s[size:]
}
