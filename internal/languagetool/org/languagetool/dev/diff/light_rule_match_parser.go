package diff

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var (
	startPattern = regexp.MustCompile(`^(?:\d+\.\) )?Line (\d+), column (\d+), Rule ID: (.*) premium: (false|true)`)
	coverPattern = regexp.MustCompile(`^([ ^]+)$`)
)

// ParseResult ports JsonParseResult for text parse path.
type ParseResult struct {
	Matches []*LightRuleMatch
}

// LightRuleMatchParser ports text-output parsing (JSON deferred).
type LightRuleMatchParser struct{}

func NewLightRuleMatchParser() *LightRuleMatchParser { return &LightRuleMatchParser{} }

// ParseOutput parses CLI-style match dump from r.
func (p *LightRuleMatchParser) ParseOutput(r io.Reader) ParseResult {
	var result []*LightRuleMatch
	lineNum, columnNum := -1, -1
	var ruleID, message, context, source, title string
	var suggestion *string // nil = no Suggestion line (Java null)
	var isPremium bool
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "Message: ") {
			message = strings.TrimPrefix(line, "Message: ")
			continue
		}
		if strings.HasPrefix(line, "Suggestion: ") {
			s := strings.TrimPrefix(line, "Suggestion: ")
			suggestion = &s
			continue
		}
		if strings.HasPrefix(line, "Rule source: ") {
			source = strings.TrimPrefix(line, "Rule source: ")
			continue
		}
		if strings.HasPrefix(line, "Title: ") {
			title = strings.TrimPrefix(line, "Title: ")
			continue
		}
		if m := startPattern.FindStringSubmatch(line); m != nil {
			lineNum, _ = strconv.Atoi(m[1])
			columnNum, _ = strconv.Atoi(m[2])
			ruleID = m[3]
			isPremium = m[4] == "true"
			continue
		}
		if (suggestion != nil || message != "") && context == "" {
			if strings.HasPrefix(line, "Tags:") || strings.HasPrefix(line, "Time:") {
				continue
			}
			// blank line between matches — don't treat as context
			if strings.TrimSpace(line) == "" {
				continue
			}
			context = line
			continue
		}
		if coverPattern.MatchString(line) && strings.Contains(line, "^") {
			cover := coverPattern.FindStringSubmatch(line)[1]
			startMarkerPos := strings.IndexByte(cover, '^')
			endMarkerPos := strings.LastIndexByte(cover, '^') + 1
			coveredText := "???"
			ctx := context
			if startMarkerPos >= 0 && context != "" {
				maxEnd := endMarkerPos
				if maxEnd > len(context) {
					maxEnd = len(context)
				}
				if startMarkerPos <= maxEnd && startMarkerPos < len(context) {
					coveredText = context[startMarkerPos:maxEnd]
					ctx = contextWithSpan(context, startMarkerPos, maxEnd)
				}
			}
			cleanID := strings.ReplaceAll(strings.ReplaceAll(ruleID, "[off]", ""), "[temp_off]", "")
			var suggs []string
			if suggestion != nil {
				suggs = []string{*suggestion}
			} else {
				// Java Arrays.asList(null) → "[null]"
				suggs = []string{"null"}
			}
			st := StatusOn
			if strings.Contains(ruleID, "[temp_off]") {
				st = StatusTempOff
			}
			result = append(result, &LightRuleMatch{
				Line: lineNum, Column: columnNum, FullRuleID: cleanID,
				Message: message, Category: "", Context: ctx, CoveredText: coveredText,
				Suggestions: suggs, RuleSource: source, Title: title,
				Status: st, Tags: nil, Premium: isPremium,
			})
			lineNum, columnNum = -1, -1
			ruleID, message, context, source = "", "", "", ""
			suggestion = nil
			// title kept across matches
		}
	}
	return ParseResult{Matches: result}
}

func contextWithSpan(context string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(context) {
		end = len(context)
	}
	if start > end {
		start = end
	}
	return context[:start] + "<span class='marker'>" + context[start:end] + "</span>" + context[end:]
}
