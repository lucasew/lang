package wikipedia

import (
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

// SimpleWikipediaTextFilter is a lightweight SwebleWikipediaTextFilter stand-in
// covering common wiki markup for green tests (full Sweble deferred).
type SimpleWikipediaTextFilter struct{}

func NewSimpleWikipediaTextFilter() *SimpleWikipediaTextFilter {
	return &SimpleWikipediaTextFilter{}
}

var (
	reLinkPipe  = regexp.MustCompile(`\[\[([^|\]]+)\|([^\]]+)\]\]`)
	reLinkPlain = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	reExtLink   = regexp.MustCompile(`\[https?://[^ \]]+\s+([^\]]+)\]`)
	reRef       = regexp.MustCompile(`(?is)<ref[^>]*>.*?</ref>`)
	reCode      = regexp.MustCompile(`(?is)<code>(.*?)</code>`)
	reSource    = regexp.MustCompile(`(?is)<source[^>]*>.*?</source>`)
	reItalic    = regexp.MustCompile(`''+([^']+)''+`)
	reListLine  = regexp.MustCompile(`(?m)^[#*]+\s*(.*)$`)
)

// removeBalancedWiki removes [[prefix:...]] blocks with nested [[links]] balanced.
func removeBalancedWiki(s, prefix string) string {
	// prefix e.g. "[[Datei:" case-insensitive
	lower := strings.ToLower(s)
	pref := strings.ToLower(prefix)
	var b strings.Builder
	i := 0
	for i < len(s) {
		idx := strings.Index(lower[i:], pref)
		if idx < 0 {
			b.WriteString(s[i:])
			break
		}
		idx += i
		b.WriteString(s[i:idx])
		// scan balanced [[ ... ]]
		depth := 0
		j := idx
		for j < len(s) {
			if j+1 < len(s) && s[j] == '[' && s[j+1] == '[' {
				depth++
				j += 2
				continue
			}
			if j+1 < len(s) && s[j] == ']' && s[j+1] == ']' {
				depth--
				j += 2
				if depth == 0 {
					break
				}
				continue
			}
			_, size := utf8.DecodeRuneInString(s[j:])
			if size == 0 {
				j++
			} else {
				j += size
			}
		}
		i = j
	}
	return b.String()
}

// Filter returns plain text extracted from wiki markup.
func (f *SimpleWikipediaTextFilter) Filter(input string) string {
	s := input
	for _, p := range []string{"[[Datei:", "[[File:", "[[Image:", "[[datei:", "[[file:", "[[image:"} {
		s = removeBalancedWiki(s, p)
	}
	// also strip remaining case variants of media via lower scan once more
	s = removeBalancedWiki(s, "[[datei:")
	s = removeBalancedWiki(s, "[[file:")
	s = removeBalancedWiki(s, "[[image:")

	s = reRef.ReplaceAllString(s, "")
	s = reSource.ReplaceAllString(s, " ")
	s = reCode.ReplaceAllString(s, "$1")
	s = reExtLink.ReplaceAllString(s, "$1")
	// nested links: repeatedly expand pipe/plain until stable
	for i := 0; i < 8; i++ {
		n := reLinkPipe.ReplaceAllString(s, "$2")
		n = reLinkPlain.ReplaceAllString(n, "$1")
		if n == s {
			break
		}
		s = n
	}
	s = reItalic.ReplaceAllString(s, "$1")
	if reListLine.MatchString(s) {
		var parts []string
		for _, line := range strings.Split(s, "\n") {
			if m := reListLine.FindStringSubmatch(line); m != nil {
				if t := strings.TrimSpace(m[1]); t != "" {
					parts = append(parts, t)
				}
			} else if t := strings.TrimSpace(line); t != "" {
				parts = append(parts, t)
			}
		}
		s = strings.Join(parts, "\n\n")
	}
	s = html.UnescapeString(s)
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		prevSpace = false
		b.WriteRune(r)
	}
	s = strings.TrimSpace(b.String())
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}
