package rules

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/finding"
	"github.com/lucasew/lang/internal/messages"
	"github.com/lucasew/lang/internal/pipeline"
)

const (
	RuleWordRepeat     = "WORD_REPEAT_RULE"
	SeverityDuplication = "duplication"
)

// ignoreList from WordRepeatRule.ignore
var wordRepeatIgnore = map[string]bool{
	"phi": true, "li": true, "xiao": true, "duran": true,
	"wagga": true, "abdullah": true, "nwe": true, "pago": true, "cao": true,
}

// WordRepeat ports org.languagetool.rules.WordRepeatRule.
func WordRepeat(text, file, lang string, msg messages.Bundle) []finding.Finding {
	all := pipeline.WordTokenize(text)
	// non-ws with SENT_START
	var tokens []pipeline.Token
	tokens = append(tokens, pipeline.Token{Text: "SENT_START"})
	for _, t := range all {
		if t.Whitespace || strings.TrimSpace(t.Text) == "" {
			continue
		}
		tokens = append(tokens, t)
	}

	message := msg.Get("repetition")
	var out []finding.Finding
	prev := ""
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i].Text
		if isWord(tok) && strings.EqualFold(prev, tok) && !wordRepeatIgnore[strings.ToLower(tok)] {
			// span from previous token start through current token end
			prevTok := tokens[i-1]
			start := prevTok.Start
			end := tokens[i].End
			line, col := runeOffsetToLineCol(text, start)
			endLine, endCol := runeOffsetToLineCol(text, end)
			out = append(out, finding.Finding{
				File:        file,
				Line:        line,
				Column:      col,
				EndLine:     endLine,
				EndColumn:   endCol,
				Offset:      start,
				EndOffset:   end,
				Rule:        RuleWordRepeat,
				Severity:    SeverityDuplication,
				Message:     message,
				Suggestions: []string{prevTok.Text},
				Language:    lang,
			})
		}
		prev = tok
	}
	return out
}

func isWord(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}
