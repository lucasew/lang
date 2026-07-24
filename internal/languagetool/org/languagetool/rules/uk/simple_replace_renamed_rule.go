package uk

import (
	"bufio"
	"embed"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/replace_renamed.txt
var renamedFS embed.FS

var (
	renamedOnce sync.Once
	// renamedMap keys as in Java ExtraDictionaryLoader.loadLists (original case).
	renamedMap map[string][]string

	// geoPostagPattern ports SimpleReplaceRenamedRule.GEO_POSTAG_PATTERN:
	// noun:inanim.*?:prop.*|adj.*  (full match, Java Pattern.matches).
	geoPostagPattern = regexp.MustCompile(`^(?:noun:inanim.*?:prop.*|adj.*)$`)

	decomunizationURL = mustParseURL("https://uk.wikipedia.org/wiki/%D0%A1%D0%BF%D0%B8%D1%81%D0%BE%D0%BA_%D1%82%D0%BE%D0%BF%D0%BE%D0%BD%D1%96%D0%BC%D1%96%D0%B2_%D0%A3%D0%BA%D1%80%D0%B0%D1%97%D0%BD%D0%B8,_%D0%BF%D0%B5%D1%80%D0%B5%D0%B9%D0%BC%D0%B5%D0%BD%D0%BE%D0%B2%D0%B0%D0%BD%D0%B8%D1%85_%D0%B2%D0%BD%D0%B0%D1%81%D0%BB%D1%96%D0%B4%D0%BE%D0%BA_%D0%B4%D0%B5%D0%BA%D0%BE%D0%BC%D1%83%D0%BD%D1%96%D0%B7%D0%B0%D1%86%D1%96%D1%97")
)

func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func loadRenamed() map[string][]string {
	renamedOnce.Do(func() {
		f, err := renamedFS.Open("data/replace_renamed.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		// Java: line.split(" *= *|\\|") then put(split[0], subList(1, length))
		m := map[string][]string{}
		sc := bufio.NewScanner(f)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1024*1024)
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "#") || tools.JavaStringTrim(line) == "" {
				continue
			}
			// split on " *= *|" or "|"
			parts := splitRenamedJava(line)
			if len(parts) < 2 {
				continue
			}
			key := parts[0]
			list := parts[1:]
			m[key] = list
		}
		if err := sc.Err(); err != nil {
			panic(err)
		}
		renamedMap = m
	})
	return renamedMap
}

// splitRenamedJava ports Java line.split(" *= *|\\|").
func splitRenamedJava(line string) []string {
	// First field ends at " = " or "=" with optional spaces; rest split by |
	eq := regexp.MustCompile(` *= *`).FindStringIndex(line)
	if eq == nil {
		return nil
	}
	key := line[:eq[0]]
	rest := line[eq[1]:]
	out := []string{key}
	if rest == "" {
		return out
	}
	for _, p := range strings.Split(rest, "|") {
		out = append(out, p)
	}
	return out
}

// SimpleReplaceRenamedRule ports org.languagetool.rules.uk.SimpleReplaceRenamedRule.
// Match is lemma + GEO POS only (Java); without tags/lemmas fail closed (no surface invent).
// Java: setLocQualityIssueType(ITSIssueType.Style).
type SimpleReplaceRenamedRule struct {
	messages  map[string]string
	IssueType rules.ITSIssueType
}

func NewSimpleReplaceRenamedRule(messages map[string]string) *SimpleReplaceRenamedRule {
	_ = loadRenamed()
	return &SimpleReplaceRenamedRule{
		messages:  messages,
		IssueType: rules.ITSStyle,
	}
}

func (r *SimpleReplaceRenamedRule) GetID() string { return "UK_SIMPLE_REPLACE_RENAMED" }

func (r *SimpleReplaceRenamedRule) GetDescription() string {
	return "Пропозиція поточної назви для перейменованих власних назв"
}

// GetLocQualityIssueType ports Rule.getLocQualityIssueType (Java Style).
func (r *SimpleReplaceRenamedRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return rules.ITSStyle
	}
	return r.IssueType
}

func (r *SimpleReplaceRenamedRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	m := loadRenamed()
	tokens := sentence.GetTokensWithoutWhitespace()
	var out []*rules.RuleMatch
	for _, tokenReadings := range tokens {
		if tokenReadings == nil || tokenReadings.IsSentenceStart() {
			continue
		}
		// LinkedHashSet order of first-seen lemmas
		var renamedLemmas []string
		seen := map[string]bool{}
		clear := false
		for _, reading := range tokenReadings.GetReadings() {
			if reading == nil {
				continue
			}
			pos := reading.GetPOSTag()
			if pos != nil && *pos == languagetool.SentenceEndTagName {
				continue
			}
			lemmaPtr := reading.GetLemma()
			if lemmaPtr == nil {
				continue
			}
			lemma := *lemmaPtr
			repl, inList := m[lemma]
			_ = repl
			if inList && geoPOSMatch(pos) {
				if !seen[lemma] {
					seen[lemma] = true
					renamedLemmas = append(renamedLemmas, lemma)
				}
			} else {
				// overlaps with normal lemma
				renamedLemmas = nil
				clear = true
				break
			}
		}
		if clear || len(renamedLemmas) == 0 {
			continue
		}
		info := ""
		var replacements []string
		for _, lemma := range renamedLemmas {
			repl := m[lemma]
			if len(repl) == 0 {
				continue
			}
			replacements = append(replacements, repl[0])
			// Java: for i=1; i<repl.size()-1; i++
			for i := 1; i < len(repl)-1; i++ {
				replacements = append(replacements, repl[i])
			}
			if info == "" && len(repl) > 1 {
				info = repl[len(repl)-1]
			}
		}
		if len(replacements) == 0 {
			continue
		}
		// Java createRuleMatch: message uses first lemma as tokenStr
		msgLemma := renamedLemmas[0]
		msg := "«" + msgLemma + "» було перейменовано"
		if info != "" {
			msg += " (" + info + ")"
		}
		rm := rules.NewRuleMatch(r, sentence, tokenReadings.GetStartPos(), tokenReadings.GetEndPos(), msg)
		rm.ShortMessage = "Перейменована назва"
		rm.SetSuggestedReplacements(replacements)
		if strings.Contains(info, "декомуніз") {
			rm.URL = decomunizationURL.String()
		}
		out = append(out, rm)
	}
	return out
}

// geoPOSMatch ports PosTagHelper.hasPosTag(reading, GEO_POSTAG_PATTERN) for one reading.
func geoPOSMatch(pos *string) bool {
	if pos == nil {
		return false
	}
	return geoPostagPattern.MatchString(*pos)
}
