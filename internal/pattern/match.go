package pattern

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/finding"
	"github.com/lucasew/lang/internal/pipeline"
)

// MatchContext holds analyzed tokens for matching.
type MatchContext struct {
	// File label for findings
	File string
	// Lang code
	Lang string
	// Full text (sentence or document)
	Text string
	// BaseOffset is rune offset of Text within the full document
	BaseOffset int
	// All tokens including whitespace
	All []pipeline.Token
	// NonWS tokens with SENT_START at [0]
	NonWS []pipeline.Token
	// SpaceBefore[i] for NonWS[i]
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
	for _, r := range s {
		if !unicode.IsSpace(r) && r != '\u00A0' {
			return false
		}
	}
	return s != ""
}

// MatchRule runs one pattern rule against the context.
func MatchRule(r *Rule, ctx MatchContext) []finding.Finding {
	if r.RequiresPOS || len(r.Tokens) == 0 {
		return nil
	}
	if r.Default == "off" || r.Default == "temp_off" {
		// still match unless disabled by default? LT default off rules don't run unless enabled.
		return nil
	}

	var findings []finding.Finding
	tokens := ctx.NonWS
	n := len(tokens)
	// SENT_START is index 0; patterns may start matching from 0 or 1
	for start := 0; start < n; start++ {
		ok, end, markerFrom, markerTo := matchTokens(r.Tokens, ctx, start)
		if !ok {
			continue
		}
		// Antipatterns: if any antipattern matches covering same area, skip
		if antipatternBlocks(r, ctx, start, end) {
			continue
		}
		from, to := markerFrom, markerTo
		if from < 0 {
			// whole match: first real token to last
			from = tokens[start].Start
			if tokens[start].Text == "SENT_START" && start+1 < end {
				from = tokens[start+1].Start
			}
			to = tokens[end-1].End
		}
		// Absolute offsets
		absFrom := ctx.BaseOffset + from
		absTo := ctx.BaseOffset + to
		line, col := runeOffsetToLineCol(ctx.Text, from)
		// Adjust line for base? For multi-sentence, baseOffset line calc needs full text — engine passes per-sentence with base; line is local then adjusted by engine.
		sev := r.IssueType
		if sev == "" {
			sev = "other"
		}
		msg := r.Message
		if msg == "" {
			msg = r.Name
		}
		sug := r.Suggestions
		// Simple suggestion: if single static suggestion without <match>
		findings = append(findings, finding.Finding{
			File:        ctx.File,
			Line:        line,
			Column:      col,
			Offset:      absFrom,
			EndOffset:   absTo,
			Rule:        r.FullID(),
			Severity:    sev,
			Message:     msg,
			Suggestions: filterSimpleSuggestions(sug),
			Language:    ctx.Lang,
		})
		// Don't advance past start+1 only — allow overlapping; LT uses greedy then filters
	}
	return findings
}

func filterSimpleSuggestions(sugs []string) []string {
	var out []string
	for _, s := range sugs {
		if strings.Contains(s, "<") {
			continue // skip complex match-based suggestions for now
		}
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func antipatternBlocks(r *Rule, ctx MatchContext, start, end int) bool {
	for _, ap := range r.Anti {
		// try match antipattern near the match window
		for s := max(0, start-3); s < end && s < len(ctx.NonWS); s++ {
			ok, _, _, _ := matchTokens(ap, ctx, s)
			if ok {
				return true
			}
		}
	}
	return false
}

// matchTokens tries to match pattern at NonWS[start]. Returns end index exclusive, marker rune offsets relative to ctx.Text.
func matchTokens(pattern []PatToken, ctx MatchContext, start int) (ok bool, end int, markerFrom, markerTo int) {
	markerFrom, markerTo = -1, -1
	ti := start
	tokens := ctx.NonWS
	for pi := 0; pi < len(pattern); pi++ {
		pt := pattern[pi]
		min := pt.Min
		max := pt.Max
		if max < min {
			max = min
		}
		matchedOcc := 0
		occStart := ti
		for matchedOcc < max {
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
				// skip attribute: optional skip after match handled after loop
			} else {
				break
			}
		}
		if matchedOcc < min {
			// try skip from previous?
			return false, start, -1, -1
		}
		// Handle skip: up to Skip following tokens can be ignored before next pattern element — LT skip is on the element meaning next tokens skipped after this element
		if pt.Skip > 0 && pi+1 < len(pattern) {
			// not consuming fixed; next element search with skip window — simplified: allow next match within skip
			// For full fidelity need more complex performer; implement simple skip as optional advances
		}
		_ = occStart
	}
	return true, ti, markerFrom, markerTo
}

func tokenMatches(pt PatToken, ctx MatchContext, ti int) bool {
	tok := ctx.NonWS[ti]
	// spacebefore
	if pt.SpaceBefore == "yes" && (ti >= len(ctx.SpaceBefore) || !ctx.SpaceBefore[ti]) {
		return false
	}
	if pt.SpaceBefore == "no" && ti < len(ctx.SpaceBefore) && ctx.SpaceBefore[ti] {
		return false
	}
	// POS required — fail closed
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
	// exceptions
	for _, ex := range pt.Exceptions {
		if tokenMatches(ex, ctx, ti) {
			return false
		}
	}
	// AND group: all must match
	for _, a := range pt.And {
		if !tokenMatches(a, ctx, ti) {
			return false
		}
	}
	// OR group: any
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
		// empty token matches any token (POS-only or bare)
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
