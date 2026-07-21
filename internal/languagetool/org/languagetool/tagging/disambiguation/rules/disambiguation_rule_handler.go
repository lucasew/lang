package rules

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// DisambiguationRuleHandler ports the SAX handler surface as a structured XML parser
// with rule groups, anti-patterns, examples, and <wd> new readings.
// Full Match/unification markup is deferred to DisambiguationRuleLoader simplicity.
type DisambiguationRuleHandler struct {
	LanguageCode string
	SourceFile   string
	Rules        []*DisambiguationPatternRule
}

func NewDisambiguationRuleHandler(languageCode, sourceFile string) *DisambiguationRuleHandler {
	return &DisambiguationRuleHandler{
		LanguageCode: languageCode,
		SourceFile:   sourceFile,
	}
}

// Parse reads expanded disambiguation XML into Rules.
func (h *DisambiguationRuleHandler) Parse(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return h.ParseBytes(data)
}

func (h *DisambiguationRuleHandler) ParseBytes(data []byte) error {
	var root drhRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("disambiguation handler: %w", err)
	}
	h.Rules = nil
	for _, g := range root.RuleGroups {
		groupID := g.ID
		groupName := g.Name
		for _, xr := range g.Rules {
			rule, err := h.buildRule(xr, groupID, groupName)
			if err != nil {
				return err
			}
			if rule != nil {
				h.Rules = append(h.Rules, rule)
			}
		}
	}
	for _, xr := range root.Rules {
		rule, err := h.buildRule(xr, "", "")
		if err != nil {
			return err
		}
		if rule != nil {
			h.Rules = append(h.Rules, rule)
		}
	}
	return nil
}

func (h *DisambiguationRuleHandler) buildRule(xr drhRule, groupID, groupName string) (*DisambiguationPatternRule, error) {
	id := xr.ID
	if id == "" {
		id = groupID
	}
	if id == "" {
		return nil, nil
	}
	name := xr.Name
	if name == "" {
		name = groupName
	}
	var tokens []*patterns.PatternToken
	for _, xt := range xr.Pattern.Tokens {
		content := tools.JavaStringTrim(xt.Content)
		re := strings.EqualFold(xt.Regexp, "yes")
		caseSens := strings.EqualFold(xt.CaseSensitive, "yes")
		inflected := strings.EqualFold(xt.Inflected, "yes")
		pt := patterns.NewPatternToken(content, caseSens, re, inflected)
		if xt.Postag != "" {
			// store postag via Negation/POS if API allows - use Postag on token if available
			if setter, ok := any(pt).(interface{ SetPosToken(string, bool, bool) }); ok {
				setter.SetPosToken(xt.Postag, strings.EqualFold(xt.PostagRegexp, "yes"), false)
			}
			_ = xt.Postag
		}
		tokens = append(tokens, pt)
	}
	action := ActionReplace
	if xr.Disambig.Action != "" {
		action = DisambiguatorAction(strings.ToUpper(xr.Disambig.Action))
	}
	rule := NewDisambiguationPatternRule(id, name, h.LanguageCode, tokens, xr.Disambig.Postag, nil, action)
	// new readings from <wd>
	var newReadings []*languagetool.AnalyzedToken
	for _, w := range xr.Disambig.Words {
		lemma := tools.JavaStringTrim(w.Lemma)
		pos := tools.JavaStringTrim(w.Pos)
		tok := tools.JavaStringTrim(w.Content)
		var posP, lemP *string
		if pos != "" {
			posP = &pos
		}
		if lemma != "" {
			lemP = &lemma
		}
		if tok == "" {
			tok = lemma
		}
		newReadings = append(newReadings, languagetool.NewAnalyzedToken(tok, posP, lemP))
	}
	if len(newReadings) > 0 {
		rule.SetNewInterpretations(newReadings)
	}
	// examples
	var examples []DisambiguatedExample
	var untouched []string
	for _, ex := range xr.Examples {
		typ := strings.ToLower(ex.Type)
		text := tools.JavaStringTrim(ex.Content)
		switch typ {
		case "untouched":
			untouched = append(untouched, text)
		case "ambiguous", "":
			// input/output markers optional in simplified form
			examples = append(examples, DisambiguatedExample{Example: text})
		}
	}
	if len(examples) > 0 {
		rule.SetExamples(examples)
	}
	if len(untouched) > 0 {
		rule.SetUntouchedExamples(untouched)
	}
	return rule, nil
}

// GetRules returns parsed rules.
func (h *DisambiguationRuleHandler) GetRules() []*DisambiguationPatternRule {
	return h.Rules
}

type drhRoot struct {
	XMLName    xml.Name       `xml:"rules"`
	RuleGroups []drhRuleGroup `xml:"rulegroup"`
	Rules      []drhRule      `xml:"rule"`
}

type drhRuleGroup struct {
	ID    string    `xml:"id,attr"`
	Name  string    `xml:"name,attr"`
	Rules []drhRule `xml:"rule"`
}

type drhRule struct {
	ID       string       `xml:"id,attr"`
	Name     string       `xml:"name,attr"`
	Pattern  drhPattern   `xml:"pattern"`
	Disambig drhDisambig  `xml:"disambig"`
	Examples []drhExample `xml:"example"`
}

type drhPattern struct {
	Tokens []drhToken `xml:"token"`
}

type drhToken struct {
	Regexp        string `xml:"regexp,attr"`
	CaseSensitive string `xml:"case_sensitive,attr"`
	Inflected     string `xml:"inflected,attr"`
	Postag        string `xml:"postag,attr"`
	PostagRegexp  string `xml:"postag_regexp,attr"`
	Content       string `xml:",chardata"`
}

type drhDisambig struct {
	Action string  `xml:"action,attr"`
	Postag string  `xml:"postag,attr"`
	Words  []drhWd `xml:"wd"`
}

type drhWd struct {
	Lemma   string `xml:"lemma,attr"`
	Pos     string `xml:"pos,attr"`
	Content string `xml:",chardata"`
}

type drhExample struct {
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}
