package nl

import (
	"bufio"
	"embed"
	"strings"
	"sync"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/multipartcompounds.txt
var multipartFS embed.FS

var (
	spaceCompoundOnce sync.Once
	// spacedVariant → (glued form, message)
	spaceCompoundHits map[string]spaceHit
)

type spaceHit struct {
	glued   string
	message string
}

// GenerateVariants ports SpaceInCompoundRule.generateVariants.
func GenerateVariants(soFar string, l []string, result map[string]struct{}) {
	if len(l) == 1 {
		if strings.Contains(soFar, " ") {
			result[soFar+l[0]] = struct{}{}
		}
		result[soFar+" "+l[0]] = struct{}{}
		return
	}
	rest := l[1:]
	GenerateVariants(soFar+l[0], rest, result)
	if soFar != "" {
		GenerateVariants(soFar+" "+l[0], rest, result)
	}
}

func loadSpaceCompounds() map[string]spaceHit {
	spaceCompoundOnce.Do(func() {
		f, err := multipartFS.Open("data/multipartcompounds.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		out := map[string]spaceHit{}
		sc := bufio.NewScanner(f)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1024*1024)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || line[0] == '#' {
				continue
			}
			parts := strings.SplitN(line, "|", 2)
			wordParts := parts[0]
			if !strings.Contains(wordParts, " ") {
				continue
			}
			words := strings.Split(wordParts, " ")
			glued := strings.Join(words, "")
			msg := "Waarschijnlijk bedoelt u: " + glued
			if len(parts) == 2 {
				msg += " (" + parts[1] + ")"
			}
			variants := map[string]struct{}{}
			GenerateVariants("", words, variants)
			for v := range variants {
				// only keep variants that still contain a space (spaced mistakes)
				if strings.Contains(v, " ") {
					out[v] = spaceHit{glued: glued, message: msg}
				}
			}
		}
		if err := sc.Err(); err != nil {
			panic(err)
		}
		spaceCompoundHits = out
	})
	return spaceCompoundHits
}

// SpaceInCompoundRule ports org.languagetool.rules.nl.SpaceInCompoundRule.
type SpaceInCompoundRule struct {
	Messages map[string]string
}

func NewSpaceInCompoundRule(messages map[string]string) *SpaceInCompoundRule {
	_ = loadSpaceCompounds()
	return &SpaceInCompoundRule{Messages: messages}
}

func (r *SpaceInCompoundRule) GetID() string { return "NL_SPACE_IN_COMPOUND" }

// isNonLetterBoundary is true when s is not a letter (word edge).
func isNonLetterBoundary(s string) bool {
	if s == "" {
		return true
	}
	for _, r := range s {
		return !unicode.IsLetter(r)
	}
	return true
}

func (r *SpaceInCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	hits := loadSpaceCompounds()
	text := sentence.GetText()
	var matches []*rules.RuleMatch
	// Longest-first scan: check all variants present in text
	for variant, hit := range hits {
		start := 0
		for {
			idx := strings.Index(text[start:], variant)
			if idx < 0 {
				break
			}
			begin := start + idx
			end := begin + len(variant)
			// boundary check: not substring of larger letter word
			if begin > 0 && !isNonLetterBoundary(text[begin-1:begin]) {
				start = begin + 1
				continue
			}
			if end < len(text) && !isNonLetterBoundary(text[end:end+1]) {
				start = begin + 1
				continue
			}
			rm := rules.NewRuleMatch(r, sentence, begin, end, hit.message)
			rm.SetSuggestedReplacement(hit.glued)
			matches = append(matches, rm)
			start = end
		}
	}
	return matches
}
