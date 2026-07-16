package pattern

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/finding"
	"github.com/lucasew/lang/internal/pipeline"
	"github.com/lucasew/lang/internal/tagger"
)

// MatchContext holds analyzed tokens for matching.
type MatchContext struct {
	File        string
	Lang        string
	Text        string
	BaseOffset  int
	All         []pipeline.Token
	NonWS       []pipeline.Token
	SpaceBefore []bool
}

// NewMatchContext builds context from a text span.
// If tg is non-nil, non-whitespace tokens receive POS readings.
// Callers may run disambiguation on NonWS afterwards.
func NewMatchContext(file, lang, text string, baseOffset int, tg *tagger.Tagger) MatchContext {
	all := pipeline.WordTokenize(text)
	nonWS := make([]pipeline.Token, 0, len(all)+1)
	nonWS = append(nonWS, pipeline.Token{
		Text:     "SENT_START",
		Start:    0,
		End:      0,
		Readings: []pipeline.Reading{{Lemma: "SENT_START", POS: "SENT_START"}},
	})
	spaceBefore := []bool{false}
	prevWS := false
	for _, t := range all {
		if t.Whitespace || isOnlySpace(t.Text) {
			prevWS = true
			continue
		}
		if tg != nil {
			for _, r := range tg.TagWord(t.Text) {
				t.Readings = append(t.Readings, pipeline.Reading{Lemma: r.Lemma, POS: r.POS})
			}
		}
		nonWS = append(nonWS, t)
		spaceBefore = append(spaceBefore, prevWS)
		prevWS = false
	}
	return MatchContext{
		File: file, Lang: lang, Text: text, BaseOffset: baseOffset,
		All: all, NonWS: nonWS, SpaceBefore: spaceBefore,
	}
}

func isOnlySpace(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsSpace(r) && r != '\u00A0' {
			return false
		}
	}
	return true
}

// MatchRule runs one pattern rule against the context.
func MatchRule(r *Rule, ctx MatchContext) []finding.Finding {
	if len(r.Tokens) == 0 {
		return nil
	}
	if r.Default == "off" || r.Default == "temp_off" {
		return nil
	}
	if r.Incomplete {
		return nil
	}
	// Premium/AI rule packs need filters/models we do not run yet.
	if strings.HasPrefix(r.ID, "AI_") || strings.HasPrefix(r.ID, "QB_") {
		return nil
	}
	// Chunk-based rules still unsupported.
	if ruleNeedsChunk(r) {
		return nil
	}

	var findings []finding.Finding
	tokens := ctx.NonWS
	n := len(tokens)
	for start := 0; start < n; start++ {
		ok, end, markerFrom, markerTo := matchTokens(r.Tokens, ctx, start)
		if !ok {
			continue
		}
		if antipatternBlocks(r, ctx, start, end) {
			continue
		}
		from, to := markerFrom, markerTo
		if from < 0 {
			from = tokens[start].Start
			if tokens[start].Text == "SENT_START" && start+1 < end {
				from = tokens[start+1].Start
			}
			if end > 0 && end <= len(tokens) {
				to = tokens[end-1].End
			} else {
				to = from
			}
		}
		absFrom := ctx.BaseOffset + from
		absTo := ctx.BaseOffset + to
		line, col := runeOffsetToLineCol(ctx.Text, from)
		sev := r.IssueType
		if sev == "" {
			sev = "other"
		}
		msg := r.Message
		if msg == "" {
			msg = r.Name
		}
		findings = append(findings, finding.Finding{
			File:        ctx.File,
			Line:        line,
			Column:      col,
			Offset:      absFrom,
			EndOffset:   absTo,
			Rule:        r.FullID(),
			Severity:    sev,
			Message:     msg,
			Suggestions: filterSimpleSuggestions(r.Suggestions),
			Language:    ctx.Lang,
		})
	}
	return findings
}

func ruleNeedsChunk(r *Rule) bool {
	var check func([]PatToken) bool
	check = func(ts []PatToken) bool {
		for _, t := range ts {
			if t.Chunk != "" {
				return true
			}
			if check(t.Exceptions) || check(t.And) || check(t.Or) {
				return true
			}
		}
		return false
	}
	if check(r.Tokens) {
		return true
	}
	for _, ap := range r.Anti {
		if check(ap) {
			return true
		}
	}
	return false
}

func filterSimpleSuggestions(sugs []string) []string {
	var out []string
	for _, s := range sugs {
		if strings.Contains(s, "<") {
			continue
		}
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func antipatternBlocks(r *Rule, ctx MatchContext, start, end int) bool {
	for _, ap := range r.Anti {
		for s := max(0, start-5); s < end && s < len(ctx.NonWS); s++ {
			ok, _, _, _ := matchTokens(ap, ctx, s)
			if ok {
				return true
			}
		}
	}
	return false
}

func matchTokens(pattern []PatToken, ctx MatchContext, start int) (ok bool, end int, markerFrom, markerTo int) {
	markerFrom, markerTo = -1, -1
	ti := start
	tokens := ctx.NonWS
	for pi := 0; pi < len(pattern); pi++ {
		pt := pattern[pi]
		min, maxOcc := pt.Min, pt.Max
		if maxOcc < min {
			maxOcc = min
		}
		matchedOcc := 0
		for matchedOcc < maxOcc {
			if ti >= len(tokens) {
				break
			}
			if tokenMatches(pt, ctx, ti) {
				if pt.InsideMarker {
					t := tokens[ti]
					if markerFrom < 0 {
						markerFrom = t.Start
					}
					markerTo = t.End
				}
				matchedOcc++
				ti++
			} else {
				break
			}
		}
		if matchedOcc < min {
			return false, start, -1, -1
		}
		if pt.Skip > 0 && pi+1 < len(pattern) {
			next := pattern[pi+1:]
			limit := pt.Skip
			for sk := 0; sk <= limit; sk++ {
				if ti+sk > len(tokens) {
					break
				}
				ok2, end2, mf, mt := matchTokens(next, ctx, ti+sk)
				if ok2 {
					if mf >= 0 {
						if markerFrom < 0 {
							markerFrom = mf
						}
						markerTo = mt
					}
					return true, end2, markerFrom, markerTo
				}
			}
			return false, start, -1, -1
		}
	}
	return true, ti, markerFrom, markerTo
}

func tokenMatches(pt PatToken, ctx MatchContext, ti int) bool {
	tok := ctx.NonWS[ti]
	if pt.SpaceBefore == "yes" && (ti >= len(ctx.SpaceBefore) || !ctx.SpaceBefore[ti]) {
		return false
	}
	if pt.SpaceBefore == "no" && ti < len(ctx.SpaceBefore) && ctx.SpaceBefore[ti] {
		return false
	}
	if pt.Chunk != "" {
		return false
	}

	// String / lemma match
	textOK := matchText(pt, tok)
	posOK := matchPOS(pt, tok)

	// LT: if TEST_STRING then text && pos; else pos only (with negations)
	// We approximate: both dimensions must hold; empty constraints pass.
	match := textOK && posOK
	if pt.Negate {
		// Negate applies to the string element in LT when present.
		match = (!textOK) && posOK
		if pt.Value == "" && pt.Re == nil {
			match = !posOK
		}
	}
	if !match {
		return false
	}
	for _, ex := range pt.Exceptions {
		if tokenMatches(ex, ctx, ti) {
			return false
		}
	}
	for _, a := range pt.And {
		if !tokenMatches(a, ctx, ti) {
			return false
		}
	}
	if len(pt.Or) > 0 {
		any := false
		for _, o := range pt.Or {
			if tokenMatches(o, ctx, ti) {
				any = true
				break
			}
		}
		if !any {
			return false
		}
	}
	return true
}

func matchText(pt PatToken, tok pipeline.Token) bool {
	if pt.Value == "" && pt.Re == nil && !pt.Inflected {
		return true
	}
	if pt.Inflected {
		// match against lemmas
		for _, lemma := range tok.Lemmas() {
			if matchString(pt, lemma) {
				return true
			}
		}
		// also try surface if no readings
		return matchString(pt, tok.Text)
	}
	return matchString(pt, tok.Text)
}

func matchPOS(pt PatToken, tok pipeline.Token) bool {
	if pt.Postag == "" {
		return true
	}
	want := pt.Postag
	posNeg := pt.NegatePos
	var hit bool
	if want == "UNKNOWN" {
		hit = len(tok.Readings) == 0
	} else if pt.PostagRegexp {
		re, err := regexp.Compile("^(?:" + want + ")$")
		if err != nil {
			return false
		}
		for _, r := range tok.Readings {
			if re.MatchString(r.POS) {
				hit = true
				break
			}
		}
	} else {
		for _, r := range tok.Readings {
			if r.POS == want {
				hit = true
				break
			}
		}
	}
	if posNeg {
		return !hit
	}
	return hit
}

func matchString(pt PatToken, text string) bool {
	if pt.Value == "" && pt.Re == nil {
		return true
	}
	if pt.Regexp && pt.Re != nil {
		return pt.Re.MatchString(text)
	}
	if pt.CaseSensitive {
		return text == pt.Value
	}
	return strings.EqualFold(text, pt.Value)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func runeOffsetToLineCol(text string, runeOffset int) (line, col int) {
	line, col = 1, 1
	i := 0
	for _, r := range text {
		if i >= runeOffset {
			break
		}
		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
		i++
	}
	return line, col
}
