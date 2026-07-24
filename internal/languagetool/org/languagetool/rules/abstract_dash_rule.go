package rules

import (
	"bufio"
	"io"
	"regexp"
	"sort"
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractDashRule ports org.languagetool.rules.AbstractDashRule.
// Java: TYPOGRAPHY, Tag.picky.
type AbstractDashRule struct {
	Messages         map[string]string
	ID               string
	CompoundPatterns []string // longest first
	Message          string
	Description      string
	// Category ports Rule.category (Java TYPOGRAPHY).
	Category *Category
	// Tags ports Rule.tags (Java picky).
	Tags []Tag
	// URL ports Rule.url (Java setUrl).
	URL string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
	// IsLetter reports whether r is a word character for boundary checks.
	// When nil, ASCII letters are used (EN/PL default). RU/PT override this.
	IsLetter func(r rune) bool
}

// InitDashRuleMeta applies Java AbstractDashRule constructor metadata.
func InitDashRuleMeta(r *AbstractDashRule, messages map[string]string) {
	if r == nil {
		return
	}
	r.Messages = messages
	if r.Category == nil {
		r.Category = CatTypography.GetCategory(messages)
	}
	if len(r.Tags) == 0 {
		r.Tags = []Tag{TagPicky}
	}
}

func (r *AbstractDashRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *AbstractDashRule) GetTags() []Tag {
	if r == nil {
		return nil
	}
	return r.Tags
}

func (r *AbstractDashRule) HasTag(tag Tag) bool {
	if r == nil {
		return false
	}
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// GetURL ports Rule.getUrl.
func (r *AbstractDashRule) GetURL() string {
	if r == nil {
		return ""
	}
	return r.URL
}

// SetURL ports Rule.setUrl.
func (r *AbstractDashRule) SetURL(u string) {
	if r != nil {
		r.URL = u
	}
}

// GetDescription ports Rule.getDescription.
func (r *AbstractDashRule) GetDescription() string {
	if r == nil {
		return ""
	}
	return r.Description
}

// GetID ports Rule.getId.
func (r *AbstractDashRule) GetID() string {
	if r == nil || r.ID == "" {
		return "DASH_RULE"
	}
	return r.ID
}

// EstimateContextForSureMatch ports AbstractDashRule.estimateContextForSureMatch (Java returns 2).
func (r *AbstractDashRule) EstimateContextForSureMatch() int { return 2 }

// AddExamplePair ports Rule.addExamplePair.
func (r *AbstractDashRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *AbstractDashRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *AbstractDashRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// LoadDashCompoundPatterns ports AbstractDashRule.loadCompoundFile variants.
func LoadDashCompoundPatterns(r io.Reader) ([]string, error) {
	var words []string
	sc := bufio.NewScanner(r)
	// allow long lines
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || line[0] == '#' {
			continue
		}
		if strings.HasSuffix(line, "+") || strings.HasSuffix(line, "?") {
			continue
		}
		if strings.HasSuffix(line, "*") || strings.HasSuffix(line, "$") {
			line = line[:len(line)-1]
		}
		if !strings.Contains(line, "-") {
			continue
		}
		words = append(words, line)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var patterns []string
	for _, word := range words {
		for _, v := range []string{
			strings.ReplaceAll(word, "-", "–"),
			strings.ReplaceAll(word, "-", "—"),
			strings.ReplaceAll(word, "-", " – "),
			strings.ReplaceAll(word, "-", " — "),
		} {
			if !seen[v] {
				seen[v] = true
				patterns = append(patterns, v)
			}
		}
	}
	sort.Slice(patterns, func(i, j int) bool {
		if len(patterns[i]) != len(patterns[j]) {
			return len(patterns[i]) > len(patterns[j])
		}
		return patterns[i] < patterns[j]
	})
	return patterns, nil
}

func (r *AbstractDashRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	text := sentence.GetText()
	var matches []*RuleMatch
	startPositions := map[int]bool{}
	msg := r.Message
	if msg == "" {
		msg = "A dash was used instead of a hyphen."
	}

	for _, pattern := range r.CompoundPatterns {
		from := 0
		for {
			idx := strings.Index(text[from:], pattern)
			if idx < 0 {
				break
			}
			begin := from + idx
			end := begin + len(pattern)
			from = begin + 1

			beginU16 := utf16Offset(text, begin)
			endU16 := utf16Offset(text, end)
			if startPositions[beginU16] {
				continue
			}
			// boundary: not a language letter before/after (Java isBoundary)
			isLetter := isASCIILetter
			if r.IsLetter != nil {
				isLetter = r.IsLetter
			}
			if begin > 0 {
				r0, _ := lastRune(text[:begin])
				if isLetter(r0) {
					continue
				}
			}
			if end < len(text) {
				r0, _ := firstRune(text[end:])
				if isLetter(r0) {
					continue
				}
			}

			covered := text[begin:end]
			rm := NewRuleMatch(r, sentence, beginU16, endU16, msg)
			rm.SetSuggestedReplacement(dashToHyphen(covered))
			matches = append(matches, rm)
			startPositions[beginU16] = true
		}
	}
	return matches
}

var spacesDashes = regexp.MustCompile(` ?[–—] ?`)

func dashToHyphen(s string) string {
	return spacesDashes.ReplaceAllString(s, "-")
}

func isASCIILetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func utf16Offset(text string, byteIdx int) int {
	n := 0
	for i, r := range text {
		if i >= byteIdx {
			break
		}
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}

func firstRune(s string) (rune, int) {
	for _, r := range s {
		return r, 1
	}
	return 0, 0
}

func lastRune(s string) (rune, int) {
	var last rune
	for _, r := range s {
		last = r
	}
	return last, 1
}

// silence unused unicode import if any
