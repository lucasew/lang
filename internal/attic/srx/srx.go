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
		before, err1 := compilePart(beforeBuf.String())
		after, err2 := compilePart(afterBuf.String())
		if err1 != nil || err2 != nil {
			curBreak = ""
			beforeBuf.Reset()
			afterBuf.Reset()
			return
		}
		curRules = append(curRules, Rule{
			Break:  strings.EqualFold(curBreak, "yes"),
			Before: before,
			After:  after,
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

func compilePart(pat string) (*regexp.Regexp, error) {
	pat = strings.TrimSpace(pat)
	if pat == "" {
		return nil, nil // empty = always match (for after) / handled by caller
	}
	pat = javaRegexToGo(pat)
	return regexp.Compile("(?m:" + pat + ")")
}

// javaRegexToGo converts common Java Pattern escapes to RE2.
//
// Java SRX uses Pattern.UNICODE_CHARACTER_CLASS (see SrxTools), so \b treats
// letters like ș/ä as word chars. RE2's \b is ASCII-only ([A-Za-z0-9_]), which
// breaks Romanian șamd, German abbreviations, etc. We expand \b to a Unicode
// approximation: zero-width at ^/$, otherwise one non-word rune. SRX break
// positions use the match end, so a leading consumed separator is safe.
func javaRegexToGo(pat string) string {
	const unicodeWordBoundary = `(?:^|$|[^\p{L}\p{N}_])`
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
			// \b → Unicode-aware word boundary (Java UNICODE_CHARACTER_CLASS)
			if pat[i+1] == 'b' {
				b.WriteString(unicodeWordBoundary)
				i++
				continue
			}
			// \B non-boundary: leave as-is (rare in segment.srx)
			// \x{...} / \p{...} already fine for RE2
		}
		b.WriteByte(pat[i])
	}
	return b.String()
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
				loc := rule.Before.FindStringIndex(text[bstart:])
				if loc == nil {
					break
				}
				abs0 := bstart + loc[0]
				abs1 := bstart + loc[1]
				pos := len([]rune(text[:abs1]))
				if pos > 0 && pos < rn && !decided[pos] {
					after := string(runes[pos:])
					if matchAfter(rule.After, after) {
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

func matchAfter(re *regexp.Regexp, after string) bool {
	if re == nil {
		return true
	}
	loc := re.FindStringIndex(after)
	return loc != nil && loc[0] == 0
}
