package patterns

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PatternRuleHandler ports org.languagetool.rules.patterns.PatternRuleHandler
// for a practical grammar.xml subset (token patterns + regexp rules).
type PatternRuleHandler struct {
	*XMLRuleHandler
	SourceFile string

	// PLEASE_SPELL_ME is injected into suggestions that need spell-check (Java constant).
	// MarkerTag is the example marker.
	// Loaded regex and pattern rules.
	LoadedPatternRules []*PatternRule
	LoadedRegexRules   []*RegexPatternRule
	// Categories by id seen while parsing.
	Categories map[string]*rules.Category

	// UnifierConfiguration accumulates unification blocks (feature/type).
	UnifierConfiguration *UnifierConfiguration
}

// MarkerTag and PleaseSpellMe port PatternRuleHandler constants.
const (
	MarkerTag     = "<marker>"
	PleaseSpellMe = "<pleasespellme/>"
	RawTag        = "raw_pos"
	// spaceInRegex is the Java replacement for a bare space outside character classes.
	// Java: "(?:[\\s\u00A0\u202F]+)"
	spaceInRegex = "(?:[\\s\u00A0\u202F]+)"
)

// ReplaceSpacesInRegex ports PatternRuleHandler.replaceSpacesInRegex:
// spaces outside [] become a flexible whitespace class (incl. NBSP / NNBSP).
func ReplaceSpacesInRegex(s string) string {
	var b strings.Builder
	inBracket := false
	for _, c := range s {
		switch c {
		case '[':
			inBracket = true
			b.WriteRune(c)
		case ']':
			inBracket = false
			b.WriteRune(c)
		case ' ':
			if !inBracket {
				b.WriteString(spaceInRegex)
			} else {
				b.WriteRune(c)
			}
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}

// ReplaceSpacesInRegex method form for Java twin callers.
func (h *PatternRuleHandler) ReplaceSpacesInRegex(s string) string {
	return ReplaceSpacesInRegex(s)
}

func NewPatternRuleHandler(sourceFile, languageCode string) *PatternRuleHandler {
	return &PatternRuleHandler{
		XMLRuleHandler:       NewXMLRuleHandler(languageCode),
		SourceFile:           sourceFile,
		Categories:           map[string]*rules.Category{},
		UnifierConfiguration: NewUnifierConfiguration(),
	}
}

// GetRules returns AbstractPatternRule list (token-based only).
func (h *PatternRuleHandler) GetRules() []*AbstractPatternRule {
	return h.XMLRuleHandler.GetRules()
}

// GetAllMatchers returns both pattern and regex rules as RuleMatcher-compatible values.
func (h *PatternRuleHandler) GetPatternRules() []*PatternRule { return h.LoadedPatternRules }
func (h *PatternRuleHandler) GetRegexRules() []*RegexPatternRule {
	return h.LoadedRegexRules
}

// Parse reads grammar XML from r.
func (h *PatternRuleHandler) Parse(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return h.parseXML(data)
}

func (h *PatternRuleHandler) ParseString(xmlStr string) error {
	return h.Parse(strings.NewReader(xmlStr))
}

type grammarRoot struct {
	XMLName      xml.Name           `xml:"rules"`
	Lang         string             `xml:"lang,attr"`
	Unifications []grammarUnifyFeat `xml:"unification"`
	Categories   []grammarCategory  `xml:"category"`
}

type grammarUnifyFeat struct {
	Feature      string               `xml:"feature,attr"`
	Equivalences []grammarEquivalence `xml:"equivalence"`
}

type grammarEquivalence struct {
	Type  string        `xml:"type,attr"`
	Token *grammarToken `xml:"token"`
}

type grammarCategory struct {
	ID     string         `xml:"id,attr"`
	Name   string         `xml:"name,attr"`
	Type   string         `xml:"type,attr"`
	Rules  []grammarRule  `xml:"rule"`
	Groups []grammarGroup `xml:"rulegroup"`
}

type grammarGroup struct {
	ID    string        `xml:"id,attr"`
	Name  string        `xml:"name,attr"`
	Rules []grammarRule `xml:"rule"`
}

type grammarRule struct {
	ID      string          `xml:"id,attr"`
	Name    string          `xml:"name,attr"`
	Default string          `xml:"default,attr"`
	Pattern *grammarPattern `xml:"pattern"`
	Regexp  *grammarRegexp  `xml:"regexp"`
	Filter  *grammarFilter  `xml:"filter"`
	Message string          `xml:"message"`
	Short   string          `xml:"short"`
	URL     string          `xml:"url"`
}

type grammarFilter struct {
	Class string `xml:"class,attr"`
	Args  string `xml:"args,attr"`
}

type grammarPattern struct {
	CaseSensitive string         `xml:"case_sensitive,attr"`
	Tokens        []grammarToken `xml:"token"`
}

type grammarRegexp struct {
	Mark    string `xml:"mark,attr"`
	Content string `xml:",chardata"`
	// case_sensitive on regexp element not always present
}

type grammarToken struct {
	Regexp        string `xml:"regexp,attr"`
	Negate        string `xml:"negate,attr"`
	Inflected     string `xml:"inflected,attr"`
	CaseSensitive string `xml:"case_sensitive,attr"`
	Postag        string `xml:"postag,attr"`
	PostagRegexp  string `xml:"postag_regexp,attr"`
	Min           string `xml:"min,attr"`
	Max           string `xml:"max,attr"`
	Skip          string `xml:"skip,attr"`
	Content       string `xml:",chardata"`
}

func (h *PatternRuleHandler) parseXML(data []byte) error {
	var root grammarRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("pattern rules %s: %w", h.SourceFile, err)
	}
	if root.Lang != "" && h.LanguageCode == "" {
		h.LanguageCode = root.Lang
	}
	// unification
	for _, u := range root.Unifications {
		for _, eq := range u.Equivalences {
			var pt *PatternToken
			if eq.Token != nil {
				pt = tokenFromGrammar(*eq.Token, false)
			} else {
				pt = NewPatternToken("", false, false, false)
			}
			h.UnifierConfiguration.SetEquivalence(u.Feature, eq.Type, pt)
		}
	}
	for _, cat := range root.Categories {
		if cat.ID != "" {
			h.Categories[cat.ID] = rules.NewCategory(rules.NewCategoryId(cat.ID), orDefault(cat.Name, cat.ID))
		}
		for _, xr := range cat.Rules {
			if err := h.addRule(xr, cat.ID); err != nil {
				return err
			}
		}
		for _, g := range cat.Groups {
			for i, xr := range g.Rules {
				if xr.ID == "" {
					xr.ID = g.ID
				}
				if err := h.addRule(xr, cat.ID); err != nil {
					return err
				}
				if len(h.XMLRuleHandler.Rules) > 0 {
					h.XMLRuleHandler.Rules[len(h.XMLRuleHandler.Rules)-1].SubID = fmt.Sprintf("%d", i+1)
				}
			}
		}
	}
	return nil
}

func (h *PatternRuleHandler) addRule(xr grammarRule, categoryID string) error {
	if xr.ID == "" && !h.RelaxedMode {
		return fmt.Errorf("rule without id in %s", h.SourceFile)
	}
	lang := h.LanguageCode
	if xr.Regexp != nil {
		content := strings.TrimSpace(xr.Regexp.Content)
		re, err := regexp.Compile(content)
		if err != nil {
			return fmt.Errorf("rule %s regexp: %w", xr.ID, err)
		}
		mark := 0
		if xr.Regexp.Mark != "" {
			fmt.Sscanf(xr.Regexp.Mark, "%d", &mark)
		}
		rr := NewRegexPatternRule(xr.ID, xr.Name, strings.TrimSpace(xr.Message), strings.TrimSpace(xr.Short), "", lang, re, mark)
		if xr.Filter != nil {
			rr.FilterArgs = xr.Filter.Args
			if strings.Contains(xr.Filter.Class, "RegexAntiPatternFilter") || strings.Contains(xr.Filter.Args, "antipatterns:") {
				// applied at check time via FilterArgs
			}
		}
		h.LoadedRegexRules = append(h.LoadedRegexRules, rr)
		// also as abstract for listing
		abs := NewAbstractPatternRule(xr.ID, xr.Name, lang, nil, false)
		abs.Message = rr.Message
		abs.ShortMessage = rr.ShortMessage
		abs.SourceFile = h.SourceFile
		if xr.Default == "off" || xr.Default == "temp_off" {
			// mark via premium/off not modeled — skip default-off from active lists if needed
		}
		h.XMLRuleHandler.Rules = append(h.XMLRuleHandler.Rules, abs)
		return nil
	}
	if xr.Pattern == nil {
		if h.RelaxedMode {
			return nil
		}
		return fmt.Errorf("rule %s has neither pattern nor regexp", xr.ID)
	}
	caseSens := strings.EqualFold(xr.Pattern.CaseSensitive, "yes")
	var tokens []*PatternToken
	for _, xt := range xr.Pattern.Tokens {
		tokens = append(tokens, tokenFromGrammar(xt, caseSens))
	}
	pr := NewPatternRule(xr.ID, lang, tokens, xr.Name, strings.TrimSpace(xr.Message), strings.TrimSpace(xr.Short))
	h.LoadedPatternRules = append(h.LoadedPatternRules, pr)
	abs := NewAbstractPatternRule(xr.ID, xr.Name, lang, tokens, false)
	abs.Message = pr.Message
	abs.ShortMessage = pr.ShortMessage
	abs.SourceFile = h.SourceFile
	h.XMLRuleHandler.Rules = append(h.XMLRuleHandler.Rules, abs)
	_ = categoryID
	return nil
}

func tokenFromGrammar(xt grammarToken, patternCaseSensitive bool) *PatternToken {
	content := strings.TrimSpace(xt.Content)
	cs := patternCaseSensitive || strings.EqualFold(xt.CaseSensitive, "yes")
	re := strings.EqualFold(xt.Regexp, "yes")
	inf := strings.EqualFold(xt.Inflected, "yes")
	pt := NewPatternToken(content, cs, re, inf)
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

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}
