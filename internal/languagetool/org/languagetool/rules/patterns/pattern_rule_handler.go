package patterns

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
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
	ID       string         `xml:"id,attr"`
	Name     string         `xml:"name,attr"`
	Type     string         `xml:"type,attr"`
	// Default ports category default="off"|"on" (Java onByDefault).
	Default  string         `xml:"default,attr"`
	// Prio ports category prio="N".
	Prio     string         `xml:"prio,attr"`
	Tags     string         `xml:"tags,attr"`      // space-separated Tag names (e.g. picky)
	ToneTags string         `xml:"tone_tags,attr"` // space-separated ToneTag names
	Rules    []grammarRule  `xml:"rule"`
	Groups   []grammarGroup `xml:"rulegroup"`
}

type grammarGroup struct {
	ID       string        `xml:"id,attr"`
	Name     string        `xml:"name,attr"`
	// Default ports rulegroup default="off"|"temp_off" (Java ruleGroupDefaultOff/TempOff).
	Default  string        `xml:"default,attr"`
	// Prio ports rulegroup prio="N".
	Prio     string        `xml:"prio,attr"`
	// Type ports rulegroup type= (Java ruleGroupIssueType).
	Type     string        `xml:"type,attr"`
	Tags     string        `xml:"tags,attr"`
	ToneTags string        `xml:"tone_tags,attr"`
	URL      string        `xml:"url"`
	Rules    []grammarRule `xml:"rule"`
}

type grammarRule struct {
	ID           string          `xml:"id,attr"`
	Name         string          `xml:"name,attr"`
	Default      string          `xml:"default,attr"`
	// Type ports rule type= (overrides group/category).
	Type         string          `xml:"type,attr"`
	// Prio ports rule prio="N".
	Prio         string          `xml:"prio,attr"`
	Tags         string          `xml:"tags,attr"`
	ToneTags     string          `xml:"tone_tags,attr"`
	GoalSpecific string          `xml:"is_goal_specific,attr"` // "yes" / "true"
	Pattern      *grammarPattern `xml:"pattern"`
	Regexp       *grammarRegexp  `xml:"regexp"`
	Filter       *grammarFilter  `xml:"filter"`
	Message      string          `xml:"message"`
	Short        string          `xml:"short"`
	URL          string          `xml:"url"`
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
		// Java: onByDefault = !OFF.equals(default); Category stores isDefaultOff.
		onByDefault := !strings.EqualFold(strings.TrimSpace(cat.Default), XMLOff)
		if cat.ID != "" {
			h.Categories[cat.ID] = rules.NewCategoryFull(
				rules.NewCategoryId(cat.ID),
				orDefault(cat.Name, cat.ID),
				rules.CategoryInternal,
				onByDefault,
				"",
			)
		}
		catTones := parseToneTagsAttr(cat.ToneTags)
		catTags := parseRuleTagsAttr(cat.Tags)
		catPrio := parsePrioAttr(cat.Prio)
		catType := strings.TrimSpace(cat.Type)
		for _, xr := range cat.Rules {
			if err := h.addRule(xr, cat.ID, catTones, nil, catTags, nil, "", catPrio, 0, catType, ""); err != nil {
				return err
			}
		}
		for _, g := range cat.Groups {
			groupTones := parseToneTagsAttr(g.ToneTags)
			groupTags := parseRuleTagsAttr(g.Tags)
			groupPrio := parsePrioAttr(g.Prio)
			groupType := strings.TrimSpace(g.Type)
			groupURL := strings.TrimSpace(g.URL)
			for i, xr := range g.Rules {
				if xr.ID == "" {
					xr.ID = g.ID
				}
				if err := h.addRule(xr, cat.ID, catTones, groupTones, catTags, groupTags, g.Default, catPrio, groupPrio, catType, groupType); err != nil {
					return err
				}
				if len(h.LoadedPatternRules) > 0 {
					// last pattern rule SubID / URL group inherit for group children
					last := h.LoadedPatternRules[len(h.LoadedPatternRules)-1]
					last.SubID = fmt.Sprintf("%d", i+1)
					if last.URL == "" && groupURL != "" {
						last.URL = groupURL
					}
				}
				if len(h.XMLRuleHandler.Rules) > 0 {
					h.XMLRuleHandler.Rules[len(h.XMLRuleHandler.Rules)-1].SubID = fmt.Sprintf("%d", i+1)
				}
			}
		}
	}
	return nil
}

// parseToneTagsAttr ports PatternRuleHandler tone_tags="a b c" split + ToneTag.valueOf.
// Unknown names are skipped (Java valueOf would throw — we fail soft for incomplete enums).
func parseToneTagsAttr(s string) []languagetool.ToneTag {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out []languagetool.ToneTag
	for _, p := range strings.Fields(s) {
		if p == "" {
			continue
		}
		// Java ToneTag enum names are upper-ish; Go constants are lowercase values.
		tag := languagetool.ToneTag(strings.ToLower(p))
		out = append(out, tag)
	}
	return out
}

// parseRuleTagsAttr ports tags="picky …" → []rules.Tag (Java Tag.valueOf).
func parseRuleTagsAttr(s string) []rules.Tag {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out []rules.Tag
	for _, p := range strings.Fields(s) {
		if p == "" {
			continue
		}
		out = append(out, rules.Tag(strings.ToLower(p)))
	}
	return out
}

// mergeToneTags ports successive rule.addToneTags(category/group/rule) lists.
func mergeToneTags(parts ...[]languagetool.ToneTag) []languagetool.ToneTag {
	var out []languagetool.ToneTag
	seen := map[languagetool.ToneTag]struct{}{}
	for _, part := range parts {
		for _, t := range part {
			if _, ok := seen[t]; ok {
				continue
			}
			seen[t] = struct{}{}
			out = append(out, t)
		}
	}
	return out
}

// mergeRuleTags ports successive rule.addTags lists (rule + group + category).
func mergeRuleTags(parts ...[]rules.Tag) []rules.Tag {
	var out []rules.Tag
	seen := map[rules.Tag]struct{}{}
	for _, part := range parts {
		for _, t := range part {
			if _, ok := seen[t]; ok {
				continue
			}
			seen[t] = struct{}{}
			out = append(out, t)
		}
	}
	return out
}

// resolveGoalSpecific ports PatternRuleHandler is_goal_specific inheritance:
// rule attribute, else rulegroup, else category, else false.
// TRUE/yes → true; FALSE/no → false; empty → fall through.
func resolveGoalSpecific(ruleAttr, groupAttr, catAttr string) bool {
	for _, a := range []string{ruleAttr, groupAttr, catAttr} {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		if strings.EqualFold(a, "true") || strings.EqualFold(a, "yes") {
			return true
		}
		if strings.EqualFold(a, "false") || strings.EqualFold(a, "no") {
			return false
		}
	}
	return false
}

func (h *PatternRuleHandler) addRule(xr grammarRule, categoryID string, categoryTones, groupTones []languagetool.ToneTag, categoryTags, groupTags []rules.Tag, groupDefault string, catPrio, groupPrio int, catType, groupType string) error {
	if xr.ID == "" && !h.RelaxedMode {
		return fmt.Errorf("rule without id in %s", h.SourceFile)
	}
	lang := h.LanguageCode
	ruleTones := parseToneTagsAttr(xr.ToneTags)
	tones := mergeToneTags(ruleTones, groupTones, categoryTones)
	tags := mergeRuleTags(parseRuleTagsAttr(xr.Tags), groupTags, categoryTags)
	goalSpecific := strings.EqualFold(xr.GoalSpecific, "yes") || strings.EqualFold(xr.GoalSpecific, "true")
	// Java PatternRuleHandler: rulegroup default off/temp_off overrides rule-level default.
	defaultOff, defaultTempOff := resolveRuleDefaultOff(xr.Default, groupDefault)
	issueType := resolveIssueType(xr.Type, groupType, catType)
	prio := resolvePriority(catPrio, groupPrio, parsePrioAttr(xr.Prio))
	ruleURL := strings.TrimSpace(xr.URL)
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
		if rr.AbstractPatternRule != nil {
			rr.AbstractPatternRule.ToneTags = tones
			rr.AbstractPatternRule.GoalSpecific = goalSpecific
			rr.AbstractPatternRule.DefaultOff = defaultOff
			rr.AbstractPatternRule.DefaultTempOff = defaultTempOff
			rr.AbstractPatternRule.IssueType = issueType
			rr.AbstractPatternRule.Priority = prio
			rr.AbstractPatternRule.URL = ruleURL
			rr.AbstractPatternRule.SourceFile = h.SourceFile
		}
		h.LoadedRegexRules = append(h.LoadedRegexRules, rr)
		// also as abstract for listing
		abs := NewAbstractPatternRule(xr.ID, xr.Name, lang, nil, false)
		abs.Message = rr.Message
		abs.ShortMessage = rr.ShortMessage
		abs.SourceFile = h.SourceFile
		abs.ToneTags = tones
		abs.GoalSpecific = goalSpecific
		abs.DefaultOff = defaultOff
		abs.DefaultTempOff = defaultTempOff
		abs.IssueType = issueType
		abs.Priority = prio
		abs.URL = ruleURL
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
	pr.ToneTags = tones
	pr.GoalSpecific = goalSpecific
	pr.DefaultOff = defaultOff
	pr.DefaultTempOff = defaultTempOff
	pr.IssueType = issueType
	pr.Priority = prio
	pr.URL = ruleURL
	pr.SourceFile = h.SourceFile
	if len(tags) > 0 {
		pr.SetTags(tags)
	}
	h.LoadedPatternRules = append(h.LoadedPatternRules, pr)
	abs := NewAbstractPatternRule(xr.ID, xr.Name, lang, tokens, false)
	abs.Message = pr.Message
	abs.ShortMessage = pr.ShortMessage
	abs.SourceFile = h.SourceFile
	abs.ToneTags = tones
	abs.GoalSpecific = goalSpecific
	abs.DefaultOff = defaultOff
	abs.DefaultTempOff = defaultTempOff
	abs.IssueType = issueType
	abs.Priority = prio
	abs.URL = ruleURL
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
