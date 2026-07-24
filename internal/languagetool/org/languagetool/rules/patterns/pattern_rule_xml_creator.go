package patterns

import (
	"fmt"
	"html"
	"strings"
)

// PatternRuleXmlCreator ports org.languagetool.rules.patterns.PatternRuleXmlCreator
// for in-memory rules (file/XPath lookup deferred).
type PatternRuleXmlCreator struct{}

func NewPatternRuleXmlCreator() *PatternRuleXmlCreator { return &PatternRuleXmlCreator{} }

// ToXMLFromRule serializes a PatternRule to an indented XML fragment.
func (c *PatternRuleXmlCreator) ToXMLFromRule(rule *PatternRule) string {
	if rule == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<rule id="%s" name="%s">`, xmlEsc(rule.ID), xmlEsc(rule.Description)))
	b.WriteString("\n  <pattern")
	b.WriteString(">\n")
	for _, tok := range rule.Tokens {
		b.WriteString("    ")
		b.WriteString(tokenToXML(tok))
		b.WriteByte('\n')
	}
	b.WriteString("  </pattern>")
	if rule.Message != "" {
		b.WriteString("\n  <message>")
		b.WriteString(rule.Message) // may contain <suggestion> markup
		b.WriteString("</message>")
	}
	if rule.ShortMessage != "" {
		b.WriteString("\n  <short>")
		b.WriteString(xmlEsc(rule.ShortMessage))
		b.WriteString("</short>")
	}
	b.WriteString("\n</rule>")
	return b.String()
}

// ToXMLFromAbstract serializes an AbstractPatternRule.
func (c *PatternRuleXmlCreator) ToXMLFromAbstract(rule *AbstractPatternRule) string {
	if rule == nil {
		return ""
	}
	pr := &PatternRule{
		ID:           rule.ID,
		Description:  rule.Description,
		Tokens:       rule.PatternTokens,
		Message:      rule.Message,
		ShortMessage: rule.ShortMessage,
	}
	return c.ToXMLFromRule(pr)
}

// IndentXML applies the same lightweight indentation heuristics as Java nodeToString.
func IndentXML(xml string) string {
	return strings.NewReplacer(
		"<token", "\n    <token",
		"<and", "\n    <and",
		"</and>", "\n    </and>",
		"<phraseref", "\n    <phraseref",
		"<antipattern", "\n  <antipattern",
		"<pattern", "\n  <pattern",
		"</pattern", "\n  </pattern",
		"</antipattern", "\n  </antipattern",
		"</rule>", "\n</rule>",
		"<filter", "\n  <filter",
		"<message", "\n  <message",
		"<short", "\n  <short",
		"<url", "\n  <url",
		"<example", "\n  <example",
		"</suggestion><suggestion>", "</suggestion>\n  <suggestion>",
		"</message><suggestion>", "</message>\n  <suggestion>",
	).Replace(xml)
}

func tokenToXML(pt *PatternToken) string {
	if pt == nil {
		return "<token/>"
	}
	attrs := ""
	if pt.Regexp {
		attrs += ` regexp="yes"`
	}
	if pt.CaseSensitive {
		attrs += ` case_sensitive="yes"`
	}
	if pt.Negation {
		attrs += ` negate="yes"`
	}
	if pt.MinOccurrence != 1 {
		attrs += fmt.Sprintf(` min="%d"`, pt.MinOccurrence)
	}
	if pt.MaxOccurrence != 1 {
		attrs += fmt.Sprintf(` max="%d"`, pt.MaxOccurrence)
	}
	if pt.SkipNext != 0 {
		attrs += fmt.Sprintf(` skip="%d"`, pt.SkipNext)
	}
	inner := xmlEsc(pt.Token)
	if pt.Pos != nil && pt.Pos.PosTag != "" {
		// pos-only token
		if pt.Token == "" {
			posAttr := ` postag="` + xmlEsc(pt.Pos.PosTag) + `"`
			if pt.Pos.Regexp {
				posAttr += ` postag_regexp="yes"`
			}
			return "<token" + attrs + posAttr + "/>"
		}
	}
	if pt.Token == "" {
		return "<token" + attrs + "/>"
	}
	return "<token" + attrs + ">" + inner + "</token>"
}

func xmlEsc(s string) string {
	return html.EscapeString(s)
}
