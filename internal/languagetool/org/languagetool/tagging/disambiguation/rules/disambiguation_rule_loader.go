package rules

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// DisambiguationRuleLoader ports
// org.languagetool.tagging.disambiguation.rules.DisambiguationRuleLoader
// for a simplified disambiguation.xml subset.
type DisambiguationRuleLoader struct{}

func NewDisambiguationRuleLoader() *DisambiguationRuleLoader {
	return &DisambiguationRuleLoader{}
}

// GetRulesFromReader parses simplified disambiguation rules XML.
func (l *DisambiguationRuleLoader) GetRulesFromReader(r io.Reader, languageCode, xmlPath string) ([]*DisambiguationPatternRule, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return l.parse(data, languageCode, xmlPath)
}

func (l *DisambiguationRuleLoader) GetRulesFromString(xmlStr, languageCode, xmlPath string) ([]*DisambiguationPatternRule, error) {
	return l.GetRulesFromReader(strings.NewReader(xmlStr), languageCode, xmlPath)
}

type disambigRoot struct {
	XMLName xml.Name       `xml:"rules"`
	Rules   []disambigRule `xml:"rule"`
}

type disambigRule struct {
	ID       string          `xml:"id,attr"`
	Name     string          `xml:"name,attr"`
	Pattern  disambigPattern `xml:"pattern"`
	Disambig disambigElem    `xml:"disambig"`
}

type disambigPattern struct {
	Tokens []disambigToken `xml:"token"`
}

type disambigToken struct {
	Regexp  string `xml:"regexp,attr"`
	Content string `xml:",chardata"`
}

type disambigElem struct {
	Action string `xml:"action,attr"`
	Postag string `xml:"postag,attr"`
}

func (l *DisambiguationRuleLoader) parse(data []byte, languageCode, xmlPath string) ([]*DisambiguationPatternRule, error) {
	var root disambigRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse disambiguation %s: %w", xmlPath, err)
	}
	var out []*DisambiguationPatternRule
	for _, xr := range root.Rules {
		var tokens []*patterns.PatternToken
		for _, xt := range xr.Pattern.Tokens {
			content := strings.TrimSpace(xt.Content)
			re := strings.EqualFold(xt.Regexp, "yes")
			tokens = append(tokens, patterns.NewPatternToken(content, false, re, false))
		}
		action := ActionReplace
		if xr.Disambig.Action != "" {
			action = DisambiguatorAction(strings.ToUpper(xr.Disambig.Action))
		}
		// default Java: REPLACE when only postag set
		rule := NewDisambiguationPatternRule(xr.ID, xr.Name, languageCode, tokens, xr.Disambig.Postag, nil, action)
		out = append(out, rule)
	}
	return out, nil
}
