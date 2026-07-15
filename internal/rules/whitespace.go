package rules

import (
	"unicode/utf8"

	"github.com/lucasew/lang/internal/finding"
	"github.com/lucasew/lang/internal/messages"
	"github.com/lucasew/lang/internal/pipeline"
)

const (
	// RuleWhitespace is LanguageTool MultipleWhitespaceRule id.
	RuleWhitespace = "WHITESPACE_RULE"
	// SeverityWhitespace is ITSIssueType.Whitespace as LT serializes it.
	SeverityWhitespace = "whitespace"
)

// MultipleWhitespace ports org.languagetool.rules.MultipleWhitespaceRule.
// Offsets are Unicode code points (Java char index for BMP), matching LT tests.
func MultipleWhitespace(text, file, lang string, msg messages.Bundle) []finding.Finding {
	tokens := pipeline.TokenizeWhitespaceAware(text)
	if len(tokens) == 0 {
		return nil
	}

	message := msg.Get("whitespace_repetition")
	var out []finding.Finding

	// LT starts from token 1 when SENT_START is present; our tokenizer has no SENT_START,
	// so we start at 0 — equivalent for pure whitespace detection on raw text.
	for i := 0; i < len(tokens); i++ {
		if pipeline.IsFirstWhite(tokens[i]) {
			nFirst := i
			j := i + 1
			for j < len(tokens) && pipeline.IsRemovableWhite(tokens[j]) {
				j++
			}
			last := j - 1
			if last > nFirst {
				start := tokens[nFirst].Start
				end := tokens[last].End
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
					Rule:        RuleWhitespace,
					Severity:    SeverityWhitespace,
					Message:     message,
					Suggestions: []string{tokens[nFirst].Text},
					Language:    lang,
				})
			}
			i = last
			continue
		}
		if tokens[i].Linebreak {
			j := i + 1
			for j < len(tokens) && pipeline.IsRemovableWhite(tokens[j]) {
				j++
			}
			i = j - 1
		}
	}
	return out
}

func runeOffsetToLineCol(text string, runeOffset int) (line, col int) {
	line, col = 1, 1
	if runeOffset < 0 {
		return line, col
	}
	i := 0
	for i < len(text) && runeOffset > 0 {
		r, size := utf8.DecodeRuneInString(text[i:])
		i += size
		runeOffset--
		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}
