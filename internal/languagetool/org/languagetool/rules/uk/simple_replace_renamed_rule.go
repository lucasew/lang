package uk

import (
	"bufio"
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/replace_renamed.txt
var renamedFS embed.FS

type renamedEntry struct {
	suggestions []string
	info        string
}

var (
	renamedOnce sync.Once
	renamedMap  map[string]renamedEntry // lowercase key
)

func loadRenamed() map[string]renamedEntry {
	renamedOnce.Do(func() {
		f, err := renamedFS.Open("data/replace_renamed.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m := map[string]renamedEntry{}
		sc := bufio.NewScanner(f)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1024*1024)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || line[0] == '#' {
				continue
			}
			if i := strings.IndexByte(line, '#'); i >= 0 {
				line = strings.TrimSpace(line[:i])
			}
			parts := splitRenamedLine(line)
			if len(parts) < 2 {
				continue
			}
			key := strings.ToLower(parts[0])
			info := ""
			suggs := parts[1:]
			// Metadata only when last field looks like a year or decommunization note.
			if len(suggs) > 1 && looksLikeRenamedMeta(suggs[len(suggs)-1]) {
				info = suggs[len(suggs)-1]
				suggs = suggs[:len(suggs)-1]
			}
			// Merge case-variant keys (Переяслав-Хмельницький vs переяслав-хмельницький).
			if prev, ok := m[key]; ok {
				seen := map[string]bool{}
				for _, s := range prev.suggestions {
					seen[s] = true
				}
				for _, s := range suggs {
					if !seen[s] {
						prev.suggestions = append(prev.suggestions, s)
						seen[s] = true
					}
				}
				if prev.info == "" {
					prev.info = info
				}
				m[key] = prev
			} else {
				m[key] = renamedEntry{suggestions: suggs, info: info}
			}
		}
		if err := sc.Err(); err != nil {
			panic(err)
		}
		renamedMap = m
	})
	return renamedMap
}

func splitRenamedLine(line string) []string {
	eq := strings.IndexByte(line, '=')
	if eq < 0 {
		return nil
	}
	left := strings.TrimSpace(line[:eq])
	right := strings.TrimSpace(line[eq+1:])
	parts := []string{left}
	for _, p := range strings.Split(right, "|") {
		parts = append(parts, strings.TrimSpace(p))
	}
	return parts
}

// SimpleReplaceRenamedRule ports org.languagetool.rules.uk.SimpleReplaceRenamedRule
// without POS filtering; surface + light inflection prefix match.
// Joins apostrophe-split tokens (Червонознам'янка).
type SimpleReplaceRenamedRule struct {
	messages map[string]string
}

func NewSimpleReplaceRenamedRule(messages map[string]string) *SimpleReplaceRenamedRule {
	_ = loadRenamed()
	return &SimpleReplaceRenamedRule{messages: messages}
}

func (r *SimpleReplaceRenamedRule) GetID() string { return "UK_SIMPLE_REPLACE_RENAMED" }

func (r *SimpleReplaceRenamedRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	m := loadRenamed()
	tokens := sentence.GetTokensWithoutWhitespace()
	var out []*rules.RuleMatch
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok.IsSentenceStart() || tok.IsImmunized() {
			continue
		}
		// Join non-whitespace-separated fragments (apostrophes, hyphens already one token).
		joined := tok.GetToken()
		endIdx := i
		for j := i + 1; j < len(tokens) && !tokens[j].IsWhitespaceBefore() && !tokens[j].IsSentenceStart(); j++ {
			// stop at sentence-ending punctuation alone
			if isOnlyPunct(tokens[j].GetToken()) {
				break
			}
			joined += tokens[j].GetToken()
			endIdx = j
		}
		// advance outer loop past joined pieces
		if endIdx > i {
			// will set i at end
		}
		clean := strings.TrimRight(joined, ".,;:!?\"'«»")
		if clean == "" {
			continue
		}
		entries := findRenamedEntries(m, clean)
		if len(entries) == 0 {
			if endIdx > i {
				i = endIdx
			}
			continue
		}
		var suggestions []string
		info := ""
		seen := map[string]bool{}
		for _, e := range entries {
			for _, s := range e.suggestions {
				if !seen[s] {
					seen[s] = true
					adj := s
					if tools.StartsWithUppercase(clean) && !tools.IsAllUppercase(clean) {
						adj = tools.UppercaseFirstChar(s)
					}
					suggestions = append(suggestions, adj)
				}
			}
			if info == "" && e.info != "" {
				info = e.info
			}
		}
		if len(suggestions) == 0 {
			if endIdx > i {
				i = endIdx
			}
			continue
		}
		msg := "«" + clean + "» було перейменовано"
		if info != "" {
			msg += " (" + info + ")"
		}
		from := tok.GetStartPos()
		to := tokens[endIdx].GetEndPos()
		// if trailing punct was not included in join, end at last word fragment
		if clean != joined {
			// recompute: end at clean length in utf16 from from
			to = from + utf16Len(clean)
		}
		rm := rules.NewRuleMatch(r, sentence, from, to, msg)
		rm.ShortMessage = "Перейменована назва"
		rm.SetSuggestedReplacements(suggestions)
		out = append(out, rm)
		i = endIdx
	}
	return out
}

func isOnlyPunct(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return false
		}
		// Ukrainian letters
		if r >= 0x0400 && r <= 0x04FF {
			return false
		}
	}
	// apostrophe alone is not "only punct" for joining purposes — handled by join loop
	return s != "'" && s != "’" && s != "`"
}

func findRenamedEntries(m map[string]renamedEntry, token string) []renamedEntry {
	lc := strings.ToLower(token)
	// Prefer exact key match only when present (avoids over-prefixing multiword toponyms).
	if e, ok := m[lc]; ok {
		return []renamedEntry{e}
	}
	var found []renamedEntry
	for key, e := range m {
		if renamedKeyMatches(lc, key) {
			found = append(found, e)
		}
	}
	return found
}

func renamedKeyMatches(token, key string) bool {
	if token == key {
		return true
	}
	if strings.HasPrefix(token, key) {
		return isShortInflection(token[len(key):])
	}
	for _, suf := range []string{"ий", "ій", "а", "я", "е", "є", "і", "ї", "ої", "ого", "ому", "им", "ими", "их"} {
		if strings.HasSuffix(key, suf) {
			stem := strings.TrimSuffix(key, suf)
			if stem != "" && strings.HasPrefix(token, stem) {
				return isShortInflection(token[len(stem):])
			}
		}
	}
	return false
}

func looksLikeRenamedMeta(s string) bool {
	if s == "" {
		return false
	}
	// years like 2016, 1993
	if len(s) == 4 {
		allDigit := true
		for _, r := range s {
			if r < '0' || r > '9' {
				allDigit = false
				break
			}
		}
		if allDigit {
			return true
		}
	}
	return strings.Contains(s, "декомуніз")
}

func isShortInflection(rest string) bool {
	if rest == "" {
		return true
	}
	// Multiword remainder (e.g. "-хмельницький") is not an inflection.
	if strings.Contains(rest, "-") {
		return false
	}
	n := 0
	for range rest {
		n++
		if n > 6 {
			return false
		}
	}
	return true
}
