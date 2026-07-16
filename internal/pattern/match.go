package pattern

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/finding"
	"github.com/lucasew/lang/internal/pipeline"
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
func NewMatchContext(file, lang, text string, baseOffset int) MatchContext {
	all := pipeline.WordTokenize(text)
	nonWS := make([]pipeline.Token, 0, len(all)+1)
	nonWS = append(nonWS, pipeline.Token{Text: "SENT_START", Start: 0, End: 0})
	spaceBefore := []bool{false}
	prevWS := false
	for _, t := range all {
		if t.Whitespace || isOnlySpace(t.Text) {
			prevWS = true
			continue
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
	if r.RequiresPOS || len(r.Tokens) == 0 {
		return nil
	}
	if r.Default == "off" || r.Default == "temp_off" {
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

// matchTokens matches pattern starting at NonWS[start].
// Returns end index exclusive and marker rune offsets relative to ctx.Text.
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
		if maxOcc < 1 && min == 0 {
			maxOcc = 0
		}
		// Optional with max default 1: already set in loader

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
		// skip: next pattern element may start after skipping up to Skip tokens
		if pt.Skip > 0 && pi+1 < len(pattern) {
			// Try to match remaining pattern with various skip amounts — recursive simple approach:
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
	if pt.Postag != "" || pt.Chunk != "" || pt.Inflected {
		return false
	}

	text := tok.Text
	match := matchString(pt, text)
	if pt.Negate {
		match = !match
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
