package commandline

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// MatchesAsMinimalXML ports API-style XML for rule matches (subset of LT HTTP API).
func MatchesAsMinimalXML(matches []*rules.RuleMatch, languageCode string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<matches software="LanguageTool" language="` + xmlEscape(languageCode) + `">` + "\n")
	for _, m := range matches {
		if m == nil {
			continue
		}
		id := ruleIDOfMatch(m)
		offset := m.FromPos
		length := m.ToPos - m.FromPos
		if length < 0 {
			length = 0
		}
		fmt.Fprintf(&b, `  <error fromx="%d" tox="%d" ruleId="%s" msg="%s"`,
			offset, offset+length, xmlEscape(id), xmlEscape(m.GetMessage()))
		reps := m.GetSuggestedReplacements()
		if len(reps) > 0 {
			fmt.Fprintf(&b, ` replacements="%s"`, xmlEscape(strings.Join(reps, "#")))
		}
		b.WriteString("/>\n")
	}
	b.WriteString("</matches>\n")
	return b.String()
}

func xmlEscape(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

// MatchesAsJSON uses tools.RuleMatchesAsJsonSerializer (falls back to minimal array).
func MatchesAsJSON(matches []*rules.RuleMatch, languageCode, text string) string {
	ser := tools.NewRuleMatchesAsJsonSerializer()
	ser.LanguageCode = languageCode
	mj := make([]tools.MatchForJSON, 0, len(matches))
	for _, m := range matches {
		if m == nil {
			continue
		}
		id := ruleIDOfMatch(m)
		catID, catName, issue, short := languagetool.RuleMeta(id)
		if m.IssueType != "" {
			issue = m.IssueType
		}
		if m.CategoryID != "" {
			catID = m.CategoryID
		}
		if m.CategoryName != "" {
			catName = m.CategoryName
		}
		desc := languagetool.RuleDescription(id)
		sm := m.GetShortMessage()
		if sm == "" {
			sm = short
		}
		mj = append(mj, tools.MatchForJSON{
			Message:               m.GetMessage(),
			ShortMessage:          sm,
			FromPos:               m.FromPos,
			ToPos:                 m.ToPos,
			SuggestedReplacements: m.GetSuggestedReplacements(),
			RuleID:                id,
			RuleDescription:       desc,
			IssueType:             issue,
			CategoryID:            catID,
			CategoryName:          catName,
			Severity:              languagetool.SeverityFromIssueType(issue),
			RuleURL:               languagetool.RuleURL(id, languageCode),
			Tags:                  ruleTagsOf(m),
		})
	}
	s, err := ser.RuleMatchesToJSON(mj, text, 45)
	if err != nil || s == "" {
		return matchesToMinimalJSON(matches)
	}
	return s
}

// WriteMatchesOutput writes matches according to OutputFormat.
func WriteMatchesOutput(w io.Writer, matches []*rules.RuleMatch, opts *CommandLineOptions) error {
	if w == nil {
		return nil
	}
	lang := "en"
	if opts != nil && opts.Language != "" {
		lang = opts.Language
	}
	format := OutputPlaintext
	if opts != nil {
		format = opts.OutputFormat
	}
	switch format {
	case OutputJSON:
		_, err := io.WriteString(w, MatchesAsJSON(matches, lang, ""))
		return err
	case OutputXML:
		_, err := io.WriteString(w, MatchesAsMinimalXML(matches, lang))
		return err
	case OutputSARIF:
		fn := ""
		if opts != nil {
			fn = opts.Filename
		}
		_, err := io.WriteString(w, MatchesAsSARIF(matches, "", fn, lang))
		return err
	case OutputLint:
		fn := ""
		if opts != nil {
			fn = opts.Filename
		}
		return WriteLintMatches(w, matches, "", fn)
	default:
		// plaintext handled by PrintMatches elsewhere
		return nil
	}
}

// LineColumnAt maps byte offset in text to 1-based line and column.
func LineColumnAt(text string, offset int) (line, col int) {
	line, col = 1, 1
	if offset < 0 {
		return line, col
	}
	runes := []rune(text)
	if offset > len(runes) {
		offset = len(runes)
	}
	for i := 0; i < offset; i++ {
		if runes[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}
