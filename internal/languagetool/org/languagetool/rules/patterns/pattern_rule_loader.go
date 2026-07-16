package patterns

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// PatternRuleLoader ports org.languagetool.rules.patterns.PatternRuleLoader
// for a simplified rules XML subset (full PatternRuleHandler deferred).
type PatternRuleLoader struct {
	RelaxedMode bool
}

func NewPatternRuleLoader() *PatternRuleLoader { return &PatternRuleLoader{} }

func (l *PatternRuleLoader) SetRelaxedMode(v bool) { l.RelaxedMode = v }

// GetRulesFromReader parses a simplified pattern-rules XML stream.
// filename is used only in error messages.
func (l *PatternRuleLoader) GetRulesFromReader(r io.Reader, filename, languageCode string) ([]*AbstractPatternRule, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Cannot load or parse input stream of '%s': %w", filename, err)
	}
	rules, err := l.parseRulesXML(data, languageCode)
	if err != nil {
		return nil, fmt.Errorf("Cannot load or parse input stream of '%s': %w", filename, err)
	}
	return rules, nil
}

// GetRulesFromString is a convenience wrapper.
func (l *PatternRuleLoader) GetRulesFromString(xmlStr, filename, languageCode string) ([]*AbstractPatternRule, error) {
	return l.GetRulesFromReader(strings.NewReader(xmlStr), filename, languageCode)
}

type xmlRulesRoot struct {
	XMLName    xml.Name      `xml:"rules"`
	Categories []xmlCategory `xml:"category"`
	Rules      []xmlRule     `xml:"rule"` // allow top-level rules
}

type xmlCategory struct {
	Rules      []xmlRule      `xml:"rule"`
	RuleGroups []xmlRuleGroup `xml:"rulegroup"`
}

type xmlRuleGroup struct {
	ID    string    `xml:"id,attr"`
	Name  string    `xml:"name,attr"`
	Rules []xmlRule `xml:"rule"`
}

type xmlRule struct {
	ID      string     `xml:"id,attr"`
	Name    string     `xml:"name,attr"`
	Pattern xmlPattern `xml:"pattern"`
	Message string     `xml:"message"`
	Short   string     `xml:"short"`
}

type xmlPattern struct {
	Tokens []xmlToken `xml:"token"`
}

type xmlToken struct {
	Regexp        string `xml:"regexp,attr"`
	CaseSensitive string `xml:"case_sensitive,attr"`
	Negate        string `xml:"negate,attr"`
	Min           string `xml:"min,attr"`
	Max           string `xml:"max,attr"`
	Skip          string `xml:"skip,attr"`
	Postag        string `xml:"postag,attr"`
	PostagRegexp  string `xml:"postag_regexp,attr"`
	Content       string `xml:",chardata"`
}

func (l *PatternRuleLoader) parseRulesXML(data []byte, languageCode string) ([]*AbstractPatternRule, error) {
	var root xmlRulesRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	var out []*AbstractPatternRule
	add := func(xr xmlRule, defaultID string) error {
		id := xr.ID
		if id == "" {
			id = defaultID
		}
		if id == "" && !l.RelaxedMode {
			return fmt.Errorf("rule id not set")
		}
		name := xr.Name
		if name == "" && !l.RelaxedMode {
			// name optional in some files; only fail if both missing and not relaxed
			if id == "" {
				return fmt.Errorf("rule id/name not set")
			}
		}
		var tokens []*PatternToken
		for _, xt := range xr.Pattern.Tokens {
			pt := tokenFromXML(xt)
			tokens = append(tokens, pt)
		}
		r := NewAbstractPatternRule(id, name, languageCode, tokens, false)
		r.Message = strings.TrimSpace(xr.Message)
		r.ShortMessage = strings.TrimSpace(xr.Short)
		out = append(out, r)
		return nil
	}
	for _, cat := range root.Categories {
		for _, xr := range cat.Rules {
			if err := add(xr, ""); err != nil {
				return nil, err
			}
		}
		for _, g := range cat.RuleGroups {
			for i, xr := range g.Rules {
				id := xr.ID
				if id == "" {
					id = g.ID
				}
				if err := add(xr, id); err != nil {
					return nil, err
				}
				// sub id 1-based
				if len(out) > 0 {
					out[len(out)-1].SubID = fmt.Sprintf("%d", i+1)
					if out[len(out)-1].ID == "" {
						out[len(out)-1].ID = g.ID
					}
				}
			}
		}
	}
	for _, xr := range root.Rules {
		if err := add(xr, ""); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func tokenFromXML(xt xmlToken) *PatternToken {
	content := strings.TrimSpace(xt.Content)
	cs := strings.EqualFold(xt.CaseSensitive, "yes")
	re := strings.EqualFold(xt.Regexp, "yes")
	pt := NewPatternToken(content, cs, re, false)
	if strings.EqualFold(xt.Negate, "yes") {
		pt.SetNegation(true)
	}
	if xt.Min != "" {
		var n int
		fmt.Sscanf(xt.Min, "%d", &n)
		pt.SetMinOccurrence(n)
	}
	if xt.Max != "" {
		var n int
		fmt.Sscanf(xt.Max, "%d", &n)
		pt.SetMaxOccurrence(n)
	}
	if xt.Skip != "" {
		var n int
		fmt.Sscanf(xt.Skip, "%d", &n)
		pt.SetSkipNext(n)
	}
	if xt.Postag != "" {
		pt.SetPosToken(PosToken{
			PosTag: xt.Postag,
			Regexp: strings.EqualFold(xt.PostagRegexp, "yes"),
		})
	}
	return pt
}
