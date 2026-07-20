package commandline

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const defaultContextSize = 45

// TextChecker is a pluggable checker used by CommandLineTools (avoids hard-wiring JLanguageTool).
type TextChecker interface {
	Check(text string) ([]*rules.RuleMatch, error)
}

// TextText runs checker and writes plain-text match output to w.
// Returns the number of matches.
func CheckText(w io.Writer, contents string, checker TextChecker) (int, error) {
	return CheckTextOpts(w, contents, checker, CheckTextOptions{})
}

// CheckTextOptions controls formatting for CheckTextOpts.
type CheckTextOptions struct {
	JSON         bool
	Lint         bool // SPEC tabwriter columns
	Filename     string
	ContextSize  int // -1 or 0 → default
	LineOffset   int
	PrevMatches  int
	Verbose      bool
	ListUnknown  bool
	UnknownWords []string
	// JSONWriter optional custom JSON body; when nil and JSON is true, emits a minimal array.
	JSONSerializer func(matches []*rules.RuleMatch, contents string, contextSize int) string
}

// CheckTextOpts is the full check/print entry used by the CLI.
func CheckTextOpts(w io.Writer, contents string, checker TextChecker, opts CheckTextOptions) (int, error) {
	if checker == nil {
		return 0, fmt.Errorf("nil checker")
	}
	if w == nil {
		w = io.Discard
	}
	ctx := opts.ContextSize
	if ctx <= 0 {
		ctx = defaultContextSize
	}
	start := time.Now()
	matches, err := checker.Check(contents)
	if err != nil {
		return 0, err
	}
	if opts.JSON {
		if opts.JSONSerializer != nil {
			_, _ = io.WriteString(w, opts.JSONSerializer(matches, contents, ctx))
		} else {
			_, _ = io.WriteString(w, matchesToMinimalJSON(matches))
		}
		return len(matches), nil
	}
	if opts.Lint {
		_ = WriteLintMatches(w, matches, contents, opts.Filename)
		return len(matches), nil
	}
	PrintMatches(w, matches, opts.PrevMatches, contents, ctx, opts.LineOffset, opts.Verbose)
	// sentence count estimate: non-empty lines of punctuation splits
	sentCount := estimateSentences(contents)
	DisplayTimeStats(w, start, sentCount)
	if opts.ListUnknown && len(opts.UnknownWords) > 0 {
		_, _ = fmt.Fprintf(w, "Unknown words: %s\n", strings.Join(opts.UnknownWords, ", "))
	}
	return len(matches), nil
}

// PrintMatches ports CommandLineTools.printMatches (Java text format).
// RulePriorityFn optional: lang.getRulePriority(rule); 0 omits prio= suffix.
func PrintMatches(w io.Writer, ruleMatches []*rules.RuleMatch, prevMatches int, contents string, contextSize, lineOffset int, verbose bool) {
	PrintMatchesEx(w, ruleMatches, prevMatches, contents, contextSize, lineOffset, verbose, nil)
}

// PrintMatchesEx is PrintMatches with optional rule-priority lookup.
func PrintMatchesEx(w io.Writer, ruleMatches []*rules.RuleMatch, prevMatches int, contents string, contextSize, lineOffset int, verbose bool, rulePriority func(ruleID string) int) {
	if w == nil {
		return
	}
	if contextSize <= 0 {
		contextSize = defaultContextSize
	}
	ct := tools.NewContextTools()
	ct.SetContextSize(contextSize)
	ct.SetEscapeHtml(false)
	for i, match := range ruleMatches {
		if match == nil {
			continue
		}
		// Java uses match.getLine()+1 / getColumn() after check sets them; derive from offset + lineOffset.
		line, col := lineColumnAt(contents, match.FromPos)
		line += lineOffset
		ruleID := ruleIDOf(match)
		// Java: match.getSpecificRuleId()
		if sid := specificRuleIDOf(match); sid != "" {
			ruleID = sid
		}
		output := fmt.Sprintf("%d.) Line %d, column %d, Rule ID: %s", i+1+prevMatches, line, col, ruleID)
		if sub := ruleSubIDOf(match); sub != "" {
			output += "[" + sub + "]"
		}
		// Java: premium: Premium.get().isPremiumRule
		prem := false
		if languagetool.DefaultPremium != nil {
			prem = languagetool.DefaultPremium.IsPremiumRule(ruleID)
		}
		output += " premium: " + fmt.Sprint(prem)
		if rulePriority != nil {
			if p := rulePriority(ruleID); p != 0 {
				output += fmt.Sprintf(" prio=%d", p)
			}
		}
		if verbose {
			if xn := ruleXMLLineOf(match); xn > 0 {
				output += fmt.Sprintf(" (line %d)", xn)
			}
		}
		_, _ = fmt.Fprintln(w, output)
		_, _ = fmt.Fprintf(w, "Message: %s\n", match.GetMessage())
		reps := match.GetSuggestedReplacements()
		if len(reps) > 0 {
			if len(reps) > 5 {
				reps = reps[:5]
			}
			_, _ = fmt.Fprintf(w, "Suggestion: %s\n", strings.Join(reps, "; "))
		}
		_, _ = fmt.Fprintln(w, ct.GetPlainTextContext(match.FromPos, match.ToPos, contents))
		if u := matchURL(match); u != "" {
			_, _ = fmt.Fprintf(w, "More info: %s\n", u)
		}
		if tags := ruleTagsOf(match); len(tags) > 0 {
			_, _ = fmt.Fprintf(w, "Tags: %v\n", tags)
		}
		if i < len(ruleMatches)-1 {
			_, _ = fmt.Fprintln(w)
		}
	}
}

// DisplayTimeStats ports CommandLineTools.displayTimeStats.
func DisplayTimeStats(w io.Writer, start time.Time, sentCount int) {
	if w == nil {
		return
	}
	elapsed := time.Since(start)
	ms := elapsed.Milliseconds()
	sec := elapsed.Seconds()
	var rate float64
	if sec > 0 {
		rate = float64(sentCount) / sec
	}
	_, _ = fmt.Fprintf(w, "Time: %dms for %d sentences (%.1f sentences/sec)\n", ms, sentCount, rate)
}

// FormatTagLine formats one analyzed sentence for --taggeronly style output.
func FormatTagLine(sentenceText string, tokens []string) string {
	if len(tokens) == 0 {
		return sentenceText
	}
	return sentenceText + "\n" + strings.Join(tokens, " ")
}

// FormatTaggedToken formats surface/lemma/POS for tagger-only dumps.
// Multiple readings become lemma1|lemma2 / pos1|pos2.
func FormatTaggedToken(t *languagetool.AnalyzedTokenReadings) string {
	if t == nil {
		return ""
	}
	surface := t.GetToken()
	var lemmas, poses []string
	n := t.GetReadingsLength()
	if n <= 0 {
		// fallback: try primary slot
		n = 1
	}
	for i := 0; i < n; i++ {
		at := t.GetAnalyzedToken(i)
		if at == nil {
			continue
		}
		lem, pos := surface, "_"
		if l := at.GetLemma(); l != nil && *l != "" {
			lem = *l
		}
		if p := at.GetPOSTag(); p != nil && *p != "" {
			pos = *p
		}
		if pos == "_" && lem == surface && i == 0 && n == 1 {
			// untagged
			return surface
		}
		lemmas = append(lemmas, lem)
		poses = append(poses, pos)
	}
	if len(poses) == 0 {
		return surface
	}
	return surface + "/" + strings.Join(lemmas, "|") + "/" + strings.Join(poses, "|")
}

// TagTextWithAnalyzed ports tagText(contents, lt): print AnalyzedSentence.String() per sentence.
func TagTextWithAnalyzed(w io.Writer, contents string, sentenceTokenize func(string) []string, analyzeSentence func(string) string) {
	if w == nil {
		return
	}
	if sentenceTokenize == nil {
		sentenceTokenize = func(s string) []string { return []string{s} }
	}
	if analyzeSentence == nil {
		analyzeSentence = func(s string) string { return s }
	}
	for _, sentence := range sentenceTokenize(contents) {
		_, _ = fmt.Fprintln(w, analyzeSentence(sentence))
	}
}

// TagText writes simple token lines for each sentence (pluggable sentence split + token strings).
func TagText(w io.Writer, contents string, sentences []string, analyze func(sentence string) []string) {
	if w == nil {
		return
	}
	if analyze == nil {
		analyze = func(s string) []string { return strings.Fields(s) }
	}
	if len(sentences) == 0 {
		sentences = []string{contents}
	}
	for _, s := range sentences {
		_, _ = fmt.Fprintln(w, FormatTagLine(s, analyze(s)))
	}
}

// ProfileRulesOnText runs each rule function over sentences and prints timing table.
func ProfileRulesOnText(w io.Writer, sentences []string, ruleIDs []string, matchFn func(ruleID, sentence string) int) {
	if w == nil || matchFn == nil {
		return
	}
	const iterations = 3
	_, _ = fmt.Fprintf(w, "Testing %d rules\n", len(ruleIDs))
	_, _ = fmt.Fprintf(w, "%-40s%10s%10s%10s%15s\n", "Rule ID", "Time", "Sentences", "Matches", "Sentences per sec.")
	for _, id := range ruleIDs {
		times := make([]int64, iterations)
		matchCount := 0
		for k := 0; k < iterations; k++ {
			start := time.Now()
			mc := 0
			for _, s := range sentences {
				mc += matchFn(id, s)
			}
			times[k] = time.Since(start).Milliseconds()
			matchCount = mc // last iteration count (Java sums across iterations; we report last full pass * iterations-ish)
			if k == iterations-1 {
				matchCount = mc
			}
		}
		med := medianMS(times)
		var rate float64
		if med > 0 {
			rate = float64(len(sentences)) / (float64(med) / 1000.0)
		}
		_, _ = fmt.Fprintf(w, "%-40s%10d%10d%10d%15.1f\n", id, med, len(sentences), matchCount, rate)
	}
}

func medianMS(m []int64) int64 {
	cp := append([]int64(nil), m...)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	mid := len(cp) / 2
	if len(cp)%2 == 1 {
		return cp[mid]
	}
	return (cp[mid-1] + cp[mid]) / 2
}

func lineColumnAt(text string, pos int) (line, col int) {
	line, col = 1, 1
	if pos < 0 {
		return line, col
	}
	if pos > len(text) {
		pos = len(text)
	}
	for i := 0; i < pos; {
		r, size := utf8.DecodeRuneInString(text[i:])
		i += size
		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

func ruleIDOf(m *rules.RuleMatch) string {
	if m == nil || m.Rule == nil {
		return "?"
	}
	type idder interface{ GetID() string }
	if r, ok := m.Rule.(idder); ok {
		return r.GetID()
	}
	return fmt.Sprintf("%T", m.Rule)
}

func ruleSubIDOf(m *rules.RuleMatch) string {
	if m == nil || m.Rule == nil {
		return ""
	}
	type sub interface{ GetSubID() string }
	if r, ok := m.Rule.(sub); ok {
		return r.GetSubID()
	}
	return ""
}

func estimateSentences(text string) int {
	if strings.TrimSpace(text) == "" {
		return 0
	}
	n := 0
	for _, p := range strings.FieldsFunc(text, func(r rune) bool {
		return r == '.' || r == '!' || r == '?' || r == '\n'
	}) {
		if strings.TrimSpace(p) != "" {
			n++
		}
	}
	if n == 0 {
		return 1
	}
	return n
}

func matchesToMinimalJSON(matches []*rules.RuleMatch) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, m := range matches {
		if i > 0 {
			b.WriteByte(',')
		}
		if m == nil {
			b.WriteString("null")
			continue
		}
		b.WriteString(fmt.Sprintf(`{"offset":%d,"length":%d,"message":%q,"rule":{"id":%q}}`,
			m.FromPos, m.ToPos-m.FromPos, m.Message, ruleIDOf(m)))
	}
	b.WriteByte(']')
	return b.String()
}

func specificRuleIDOf(m *rules.RuleMatch) string {
	if m == nil {
		return ""
	}
	if id := m.GetSpecificRuleId(); id != "" {
		return id
	}
	return ""
}

func matchURL(m *rules.RuleMatch) string {
	if m == nil {
		return ""
	}
	if u := m.GetURL(); u != "" {
		return u
	}
	type urler interface{ GetURL() string }
	if r, ok := m.Rule.(urler); ok {
		return r.GetURL()
	}
	return ""
}

// ruleTagsOf ports Rule.getTags() as string names for JSON rule.tags / CLI "Tags:".
func ruleTagsOf(m *rules.RuleMatch) []string {
	if m == nil || m.Rule == nil {
		return nil
	}
	// Prefer []rules.Tag (FakeRule, PatternRule, SpecificIdRule, …).
	type tagger interface{ GetTags() []rules.Tag }
	if r, ok := m.Rule.(tagger); ok {
		tags := r.GetTags()
		if len(tags) == 0 {
			return nil
		}
		out := make([]string, len(tags))
		for i, t := range tags {
			out[i] = string(t)
		}
		return out
	}
	// Fallback string surface (e.g. older stubs).
	type stringTagger interface{ GetTags() []string }
	if r, ok := m.Rule.(stringTagger); ok {
		return r.GetTags()
	}
	return nil
}

func ruleXMLLineOf(m *rules.RuleMatch) int {
	if m == nil || m.Rule == nil {
		return 0
	}
	type xl interface{ GetXmlLineNumber() int }
	if r, ok := m.Rule.(xl); ok {
		return r.GetXmlLineNumber()
	}
	return 0
}
