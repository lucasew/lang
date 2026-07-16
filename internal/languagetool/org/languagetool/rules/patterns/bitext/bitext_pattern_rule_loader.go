package bitext

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// BitextPatternRuleLoader ports patterns.bitext.BitextPatternRuleLoader for a simplified XML.
// Expected shape (simplified from Java SAX handler):
//
//	<bitextrules>
//	  <rule id="ID" name="...">
//	    <pattern lang="src"><token>...</token></pattern>
//	    <pattern lang="trg"><token>...</token></pattern>
//	    <message>...</message>
//	  </rule>
//	</bitextrules>
type BitextPatternRuleLoader struct{}

func NewBitextPatternRuleLoader() *BitextPatternRuleLoader {
	return &BitextPatternRuleLoader{}
}

func (l *BitextPatternRuleLoader) GetRules(r io.Reader, filename string) ([]*BitextPatternRule, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var root bprRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("bitext pattern rules %s: %w", filename, err)
	}
	var out []*BitextPatternRule
	for _, xr := range root.Rules {
		if len(xr.Patterns) < 2 {
			continue
		}
		src := patternRuleFrom(xr.ID, xr.Name, xr.Patterns[0], xr.Message)
		trg := patternRuleFrom(xr.ID, xr.Name, xr.Patterns[1], xr.Message)
		out = append(out, NewBitextPatternRule(src, trg))
	}
	return out, nil
}

func patternRuleFrom(id, name string, p bprPattern, message string) *patterns.PatternRule {
	lang := p.Lang
	if lang == "" {
		lang = "en"
	}
	var tokens []*patterns.PatternToken
	for _, t := range p.Tokens {
		content := strings.TrimSpace(t.Content)
		re := strings.EqualFold(t.Regexp, "yes")
		tokens = append(tokens, patterns.NewPatternToken(content, false, re, false))
	}
	if message == "" {
		message = name
	}
	return patterns.NewPatternRule(id, lang, tokens, name, message, "")
}

type bprRoot struct {
	XMLName xml.Name  `xml:"bitextrules"`
	Rules   []bprRule `xml:"rule"`
}

type bprRule struct {
	ID       string       `xml:"id,attr"`
	Name     string       `xml:"name,attr"`
	Patterns []bprPattern `xml:"pattern"`
	Message  string       `xml:"message"`
}

type bprPattern struct {
	Lang   string     `xml:"lang,attr"`
	Tokens []bprToken `xml:"token"`
}

type bprToken struct {
	Regexp  string `xml:"regexp,attr"`
	Content string `xml:",chardata"`
}
