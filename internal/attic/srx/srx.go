// Package srx implements SRX 2.0 sentence segmentation as used by LanguageTool.
package srx

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Document is a parsed segment.srx.
type Document struct {
	LangRules map[string][]Rule
	Maps      []MapRule
}

// Rule is one SRX break/exception rule.
type Rule struct {
	Break  bool
	Before *regexp.Regexp
	After  *regexp.Regexp
	// beforeNegAlts holds fixed-string alternatives from a Java (?!a|b|c) in beforebreak.
	// Before is compiled as (prefix)suffix with group 1 = prefix; a match is rejected if
	// text after the prefix starts with any alternative (Java zero-width negative lookahead).
	beforeNegAlts []string
	// afterNegAlts is the same for afterbreak (rare; Ukrainian rules).
	afterNegAlts []string
	// beforeWBGroups / afterWBGroups are 1-based capturing-group indices for empty
	// groups that replaced Java \b (RE2 has no zero-width Unicode word boundary).
	// A match is kept only if isJavaWordBoundary holds at each group's start offset.
	beforeWBGroups []int
	afterWBGroups  []int
}

// MapRule maps a language code pattern to a language rule name.
type MapRule struct {
	Pattern *regexp.Regexp
	Name    string
}

// Load parses LanguageTool's segment.srx (namespace-agnostic).
func Load(path string) (*Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parse(f)
}

func parse(r io.Reader) (*Document, error) {
	dec := xml.NewDecoder(r)
	doc := &Document{LangRules: map[string][]Rule{}}
	var (
		inLangRule bool
		langName   string
		curBreak   string
		inBefore   bool
		inAfter    bool
		beforeBuf  strings.Builder
		afterBuf   strings.Builder
		curRules   []Rule
	)
	flushRule := func() {
		if curBreak == "" && beforeBuf.Len() == 0 && afterBuf.Len() == 0 {
			return
		}
		before, beforeNeg, beforeWB, err1 := compilePart(beforeBuf.String())
		after, afterNeg, afterWB, err2 := compilePart(afterBuf.String())
		if err1 != nil || err2 != nil {
			curBreak = ""
			beforeBuf.Reset()
			afterBuf.Reset()
			return
		}
		curRules = append(curRules, Rule{
			Break:          strings.EqualFold(curBreak, "yes"),
			Before:         before,
			After:          after,
			beforeNegAlts:  beforeNeg,
			afterNegAlts:   afterNeg,
			beforeWBGroups: beforeWB,
			afterWBGroups:  afterWB,
		})
		curBreak = ""
		beforeBuf.Reset()
		afterBuf.Reset()
	}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("srx xml: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "languagerule":
				inLangRule = true
				langName = attr(t, "languagerulename")
				curRules = nil
			case "rule":
				if inLangRule {
					curBreak = attr(t, "break")
					beforeBuf.Reset()
					afterBuf.Reset()
				}
			case "beforebreak":
				inBefore = true
				beforeBuf.Reset()
			case "afterbreak":
				inAfter = true
				afterBuf.Reset()
			case "languagemap":
				pat := javaRegexToGo(attr(t, "languagepattern"))
				name := attr(t, "languagerulename")
				re, err := regexp.Compile("^(?:" + pat + ")$")
				if err == nil {
					doc.Maps = append(doc.Maps, MapRule{Pattern: re, Name: name})
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "beforebreak":
				inBefore = false
			case "afterbreak":
				inAfter = false
			case "rule":
				if inLangRule {
					flushRule()
				}
			case "languagerule":
				doc.LangRules[langName] = curRules
				inLangRule = false
				langName = ""
			}
		case xml.CharData:
			if inBefore {
				beforeBuf.Write(t)
			} else if inAfter {
				afterBuf.Write(t)
			}
		}
	}
	return doc, nil
}

func attr(se xml.StartElement, name string) string {
	for _, a := range se.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

// compilePart compiles one beforebreak/afterbreak pattern.
// Returns optional fixed-string negative-lookahead alternatives when the Java
// pattern uses (?!a|b|c) that RE2 cannot express (English "p. 6" rule, etc.),
// and 1-based group indices for Java \b positions (empty captures + runtime WB).
//
// Do not TrimSpace: Java SRX keeps significant leading/trailing spaces in
// patterns (e.g. English "[...]\\*\\*\\* " for ellipsis + space).
func compilePart(pat string) (*regexp.Regexp, []string, []int, error) {
	if pat == "" {
		return nil, nil, nil, nil // empty = always match (for after) / handled by caller
	}
	// Java (?iu) → RE2 (?i); UNICODE_CASE is approximate via \p{L} elsewhere.
	pat = strings.ReplaceAll(pat, "(?iu)", "(?i)")
	pat = strings.ReplaceAll(pat, "(?ui)", "(?i)")
	// Simple fixed-string negative lookahead: PREFIX(?!a|b|c)SUFFIX
	// (no nested parens inside the lookahead). Used by English segment.srx:
	// [\.\s](?!(on|it|...))\p{L}{1,2}\.\s
	if re, neg, wb, err, ok := tryCompileSimpleNegLookahead(pat); ok {
		return re, neg, wb, err
	}
	// Rewrite Java \b to empty () captures before other transforms (group numbers).
	pat, wbGroups := rewriteWordBoundaries(pat)
	goPat := javaRegexToGo(pat)
	re, err := regexp.Compile("(?m:" + goPat + ")")
	return re, nil, wbGroups, err
}

// rewriteWordBoundaries replaces each Java \b with an empty capturing group ()
// so RE2 can match (zero-width) while Split verifies UNICODE_CHARACTER_CLASS
// word boundaries at those offsets. RE2's \b is ASCII-only and cannot express
// Java \b; a consuming expansion (old approach) misses letter→punctuation
// boundaries such as Spanish "tal…» " (ellipsis after a letter).
//
// Returns the rewritten pattern and 1-based group indices for each \b.
func rewriteWordBoundaries(pat string) (string, []int) {
	var b strings.Builder
	b.Grow(len(pat) + 8)
	var wbGroups []int
	groupNum := 0
	for i := 0; i < len(pat); i++ {
		if pat[i] == '\\' && i+1 < len(pat) {
			if pat[i+1] == 'b' {
				// Java word boundary → empty capture; verified in Split.
				groupNum++
				wbGroups = append(wbGroups, groupNum)
				b.WriteString("()")
				i++
				continue
			}
			// Copy backslash + next; javaRegexToGo handles \uXXXX on the result.
			b.WriteByte('\\')
			b.WriteByte(pat[i+1])
			i++
			continue
		}
		if pat[i] == '(' {
			// Capturing group if not (?… special form.
			if i+1 < len(pat) && pat[i+1] == '?' {
				b.WriteByte('(')
				continue
			}
			groupNum++
			b.WriteByte('(')
			continue
		}
		b.WriteByte(pat[i])
	}
	return b.String(), wbGroups
}

// tryCompileSimpleNegLookahead rewrites A(?!alt1|alt2)B or A(?!(alt1|alt2))B
// into RE2 (A)B with group 1 = prefix, returning the fixed alternatives for
// match-time rejection (Java zero-width negative lookahead).
// Word-boundary groups inside A/B are rewritten; group 1 remains the prefix
// for neg-alt checks (wb groups are offset after that outer group).
func tryCompileSimpleNegLookahead(pat string) (*regexp.Regexp, []string, []int, error, bool) {
	const marker = "(?!"
	i := strings.Index(pat, marker)
	if i < 0 {
		return nil, nil, nil, nil, false
	}
	// Only one simple lookahead; nested or multiple → not handled here.
	if strings.Count(pat, marker) != 1 {
		return nil, nil, nil, nil, false
	}
	rest := pat[i+len(marker):]
	// Balance parentheses to find end of (?! ... ). English uses (?!(on|it|...)).
	depth := 1 // already inside the (?!  ... )
	end := -1
	for j := 0; j < len(rest); j++ {
		switch rest[j] {
		case '\\':
			j++ // skip escaped char
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				end = j
				j = len(rest) // break
			}
		}
	}
	if end < 0 {
		return nil, nil, nil, nil, false
	}
	altsRaw := rest[:end]
	// Unwrap a single outer non-capturing/capturing group: (on|it|...)
	if len(altsRaw) >= 2 && altsRaw[0] == '(' && altsRaw[len(altsRaw)-1] == ')' &&
		!strings.Contains(altsRaw[1:len(altsRaw)-1], "(") {
		altsRaw = altsRaw[1 : len(altsRaw)-1]
	}
	// Fixed-string alts only (English exclusion list: on|it|of|...).
	// Reject regex metacharacters so we never mis-handle (?!Так?) etc.
	if strings.ContainsAny(altsRaw, `+*?[]{}()^$\`) {
		return nil, nil, nil, nil, false
	}
	alts := strings.Split(altsRaw, "|")
	if len(alts) == 0 {
		return nil, nil, nil, nil, false
	}
	for _, a := range alts {
		if a == "" {
			return nil, nil, nil, nil, false
		}
	}
	prefix := pat[:i]
	suffix := rest[end+1:]
	prefix2, prefWB := rewriteWordBoundaries(prefix)
	suffix2, sufWB := rewriteWordBoundaries(suffix)
	// Outer (prefix) is group 1 for neg-alt; shift wb groups in prefix/suffix.
	var wb []int
	for _, g := range prefWB {
		wb = append(wb, g+1) // +1 for outer prefix group
	}
	prefGroups := countCapturingGroups(prefix2)
	for _, g := range sufWB {
		wb = append(wb, g+1+prefGroups)
	}
	// (prefix)suffix — group 1 ends where (?!...) was checked in Java.
	goPat := "(" + javaRegexToGo(prefix2) + ")" + javaRegexToGo(suffix2)
	re, err := regexp.Compile("(?m:" + goPat + ")")
	if err != nil {
		return nil, nil, nil, err, true
	}
	return re, alts, wb, nil, true
}

// countCapturingGroups counts capturing '(' in a pattern that has already had
// \b rewritten (no raw \b). Non-capturing (?… forms are skipped.
func countCapturingGroups(pat string) int {
	n := 0
	for i := 0; i < len(pat); i++ {
		if pat[i] == '\\' && i+1 < len(pat) {
			i++
			continue
		}
		if pat[i] == '(' {
			if i+1 < len(pat) && pat[i+1] == '?' {
				continue
			}
			n++
		}
	}
	return n
}

// javaRegexToGo converts common Java Pattern escapes to RE2.
// Word boundaries (\b) must already be rewritten via rewriteWordBoundaries;
// raw \b left here would be ASCII-only in RE2 and is not used for segment.srx.
func javaRegexToGo(pat string) string {
	var b strings.Builder
	b.Grow(len(pat) + 8)
	for i := 0; i < len(pat); i++ {
		if pat[i] == '\\' && i+1 < len(pat) {
			// \uXXXX → \x{XXXX}
			if pat[i+1] == 'u' && i+5 < len(pat) {
				hex := pat[i+2 : i+6]
				if _, err := strconv.ParseUint(hex, 16, 32); err == nil {
					b.WriteString(`\x{`)
					b.WriteString(hex)
					b.WriteByte('}')
					i += 5
					continue
				}
			}
		}
		b.WriteByte(pat[i])
	}
	return b.String()
}

// isJavaWordChar approximates Java Pattern.UNICODE_CHARACTER_CLASS \w:
// letters, digits, marks, and connector punctuation (includes '_').
func isJavaWordChar(r rune) bool {
	if r == utf8.RuneError {
		return false
	}
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return true
	}
	return unicode.In(r, unicode.Mn, unicode.Me, unicode.Mc, unicode.Pc)
}

// isJavaWordBoundary reports a Java \b at byte offset pos in s
// (boundary between word and non-word; start/end count as non-word).
func isJavaWordBoundary(s string, pos int) bool {
	if pos < 0 || pos > len(s) {
		return false
	}
	prevWord := false
	if pos > 0 {
		r, _ := utf8.DecodeLastRuneInString(s[:pos])
		prevWord = isJavaWordChar(r)
	}
	nextWord := false
	if pos < len(s) {
		r, _ := utf8.DecodeRuneInString(s[pos:])
		nextWord = isJavaWordChar(r)
	}
	return prevWord != nextWord
}

// wbGroupsOK reports whether every listed capturing group sits on a Java \b.
// groups are 1-based; sub is FindStringSubmatchIndex relative to absBase in text.
func wbGroupsOK(text string, absBase int, sub []int, groups []int) bool {
	if len(groups) == 0 {
		return true
	}
	for _, g := range groups {
		si := 2 * g
		if si+1 >= len(sub) || sub[si] < 0 {
			return false
		}
		if !isJavaWordBoundary(text, absBase+sub[si]) {
			return false
		}
	}
	return true
}

func (d *Document) ruleNames(code string) []string {
	var names []string
	seen := map[string]bool{}
	for _, m := range d.Maps {
		if m.Pattern.MatchString(code) {
			if !seen[m.Name] {
				names = append(names, m.Name)
				seen[m.Name] = true
			}
		}
	}
	return names
}

// Split splits text into sentences for LT short code (e.g. "en") with paragraph mode.
//
// Semantics match loomchild segment with segment.srx header cascade="yes":
// language maps are applied in document order; at each candidate boundary the
// first matching rule (before + after) decides break yes/no. Later rules that
// also match the same boundary are ignored (unlike last-write-wins).
func (d *Document) Split(text, shortCode, parCode string) []string {
	if text == "" {
		return nil
	}
	runes := []rune(text)
	rn := len(runes)
	names := d.ruleNames(shortCode + parCode)
	breakAt := make([]bool, rn+1)
	decided := make([]bool, rn+1)

	for _, name := range names {
		for _, rule := range d.LangRules[name] {
			// Empty before: skip (ill-defined for boundary search)
			if rule.Before == nil {
				continue
			}
			// Overlapping matches (loomchild/Java): advance one byte after each
			// match start so "d. h." gets no-break on both single-letter dots.
			// FindAllStringIndex is non-overlapping and would skip the second.
			for bstart := 0; bstart < len(text); {
				var abs0, abs1 int
				// Always use SubmatchIndex when \b groups or neg-alts need group offsets.
				needSub := len(rule.beforeNegAlts) > 0 || len(rule.beforeWBGroups) > 0
				if needSub {
					sub := rule.Before.FindStringSubmatchIndex(text[bstart:])
					if sub == nil {
						break
					}
					abs0 = bstart + sub[0]
					abs1 = bstart + sub[1]
					if !wbGroupsOK(text, bstart, sub, rule.beforeWBGroups) {
						bstart = abs0 + 1
						continue
					}
					// group 1 end: prefix before Java (?!alts); reject if alts match there.
					if len(rule.beforeNegAlts) > 0 && len(sub) >= 4 && sub[2] >= 0 {
						afterPrefix := text[bstart+sub[3]:]
						if hasFixedPrefix(afterPrefix, rule.beforeNegAlts) {
							bstart = abs0 + 1
							continue
						}
					}
				} else {
					loc := rule.Before.FindStringIndex(text[bstart:])
					if loc == nil {
						break
					}
					abs0 = bstart + loc[0]
					abs1 = bstart + loc[1]
				}
				pos := len([]rune(text[:abs1]))
				if pos > 0 && pos < rn && !decided[pos] {
					after := string(runes[pos:])
					if matchAfter(rule.After, after, rule.afterNegAlts, rule.afterWBGroups) {
						breakAt[pos] = rule.Break
						decided[pos] = true
					}
				}
				// next search starts one past this match's start (overlap)
				bstart = abs0 + 1
			}
		}
	}

	var parts []string
	start := 0
	for i := 1; i < rn; i++ {
		if breakAt[i] {
			parts = append(parts, string(runes[start:i]))
			start = i
		}
	}
	parts = append(parts, string(runes[start:]))
	out := parts[:0]
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{text}
	}
	return out
}

func matchAfter(re *regexp.Regexp, after string, negAlts []string, wbGroups []int) bool {
	if re == nil {
		return true
	}
	if len(negAlts) > 0 || len(wbGroups) > 0 {
		sub := re.FindStringSubmatchIndex(after)
		if sub == nil || sub[0] != 0 {
			return false
		}
		if !wbGroupsOK(after, 0, sub, wbGroups) {
			return false
		}
		if len(negAlts) > 0 && len(sub) >= 4 && sub[2] >= 0 {
			if hasFixedPrefix(after[sub[3]:], negAlts) {
				return false
			}
		}
		return true
	}
	loc := re.FindStringIndex(after)
	return loc != nil && loc[0] == 0
}

// hasFixedPrefix reports whether s starts with any of the fixed alternatives
// (Java (?!a|b|c) zero-width check at the current position).
func hasFixedPrefix(s string, alts []string) bool {
	for _, a := range alts {
		if strings.HasPrefix(s, a) {
			return true
		}
	}
	return false
}
