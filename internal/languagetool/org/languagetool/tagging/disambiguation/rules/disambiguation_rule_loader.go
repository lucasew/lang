package rules

import (
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// DisambiguationRuleLoader ports
// org.languagetool.tagging.disambiguation.rules.DisambiguationRuleLoader
// Loads official disambiguation.xml (rulegroups, <and>, antipatterns, unifications).
type DisambiguationRuleLoader struct{}

func NewDisambiguationRuleLoader() *DisambiguationRuleLoader {
	return &DisambiguationRuleLoader{}
}

// GetRulesFromReader parses simplified disambiguation rules XML.
func (l *DisambiguationRuleLoader) GetRulesFromReader(r io.Reader, languageCode, xmlPath string) ([]*DisambiguationPatternRule, error) {
	rules, _, err := l.GetRulesAndUnifierFromReader(r, languageCode, xmlPath)
	return rules, err
}

// GetRulesAndUnifierFromReader parses rules plus <unification> tables.
func (l *DisambiguationRuleLoader) GetRulesAndUnifierFromReader(r io.Reader, languageCode, xmlPath string) ([]*DisambiguationPatternRule, *patterns.UnifierConfiguration, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}
	// Official LT disambiguation.xml uses custom DTD + SYSTEM .ent includes
	// (e.g. pt entities/misc.ent for &numero_por_extenso_*). Java SAX resolves
	// SYSTEM paths against the document base; without baseDir unknown entities
	// expand to "" → empty regexp tokens match every word (including SENT_START).
	baseDir := ""
	if xmlPath != "" {
		baseDir = filepath.Dir(xmlPath)
	}
	data = patterns.ExpandLTXMLEntitiesWithBase(baseDir, data)
	return l.parse(data, languageCode, xmlPath)
}

func (l *DisambiguationRuleLoader) GetRulesFromString(xmlStr, languageCode, xmlPath string) ([]*DisambiguationPatternRule, error) {
	rules, _, err := l.GetRulesAndUnifierFromString(xmlStr, languageCode, xmlPath)
	return rules, err
}

// GetRulesAndUnifierFromString loads with UnifierConfiguration.
func (l *DisambiguationRuleLoader) GetRulesAndUnifierFromString(xmlStr, languageCode, xmlPath string) ([]*DisambiguationPatternRule, *patterns.UnifierConfiguration, error) {
	return l.GetRulesAndUnifierFromReader(strings.NewReader(xmlStr), languageCode, xmlPath)
}

type disambigRoot struct {
	XMLName      xml.Name              `xml:"rules"`
	Unifications []disambigUnification `xml:"unification"`
	Rules        []disambigRule        `xml:"rule"`
	// RuleGroups: Java nests many rules under <rulegroup> (not visible as top-level <rule>).
	RuleGroups []disambigRuleGroup `xml:"rulegroup"`
}

// disambigRuleGroup ports <rulegroup id="…" name="…"> containing nested <rule>.
// Rulegroup-level <antipattern> is shared by every rule in the group (Java
// DisambiguationRuleHandler.rulegroupAntiPatterns + setAntiPatterns).
type disambigRuleGroup struct {
	ID           string            `xml:"id,attr"`
	Name         string            `xml:"name,attr"`
	AntiPatterns []disambigPattern `xml:"antipattern"`
	Rules        []disambigRule    `xml:"rule"`
}

// disambigUnification ports Java <unification feature="…">.
type disambigUnification struct {
	Feature      string                `xml:"feature,attr"`
	Equivalences []disambigEquivalence `xml:"equivalence"`
}

type disambigEquivalence struct {
	Type  string        `xml:"type,attr"`
	Token disambigToken `xml:"token"`
}

// disambigFilter ports Java <filter class="…" args="…"/> on a disambiguation rule.
// Wired via setRuleFilter → keepDespiteFilter (RuleFilter.matches).
type disambigFilter struct {
	Class string `xml:"class,attr"`
	Args  string `xml:"args,attr"`
}

type disambigRule struct {
	ID           string            `xml:"id,attr"`
	Name         string            `xml:"name,attr"`
	AntiPatterns []disambigPattern `xml:"antipattern"`
	Pattern      disambigPattern   `xml:"pattern"`
	// Filter ports Java DisambiguationRuleHandler case "filter".
	Filter   *disambigFilter `xml:"filter"`
	Disambig disambigElem    `xml:"disambig"`
	// Examples ports Java <example type="untouched|ambiguous" …/> under <rule>.
	Examples []disambigExample `xml:"example"`
}

// disambigExample ports DisambiguationRuleHandler EXAMPLE.
type disambigExample struct {
	Type       string `xml:"type,attr"`
	InputForm  string `xml:"inputform,attr"`
	OutputForm string `xml:"outputform,attr"`
	// Inner keeps <marker> markup (Java example.toString() body).
	Inner string `xml:",innerxml"`
}

// disambigPattern holds pattern children in document order: <token>, <and>, <unify>, <marker>.
// Java PatternRuleHandler walks elements; empty tokens from skipped <and> must not load.
type disambigPattern struct {
	// CaseSensitive ports pattern case_sensitive="yes" (inherits to tokens without own attr).
	CaseSensitive string
	// Tokens is filled by UnmarshalXML (ordered, includes synthetic AND-group tokens).
	Tokens []disambigToken
}

// UnmarshalXML ports pattern content: sequence of <token>, <marker>…</marker>, <and>, <unify>.
// Java PatternRuleHandler walks children; skipping <marker> drops tokens and invents
// empty/exception-only patterns (e.g. EXCEPT_NOT_VERB matched every word).
// Skipping <unify> drops unified tokens entirely (UNIFY match / number agreement broken).
func (p *disambigPattern) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	p.Tokens = nil
	p.CaseSensitive = ""
	for _, a := range start.Attr {
		if a.Name.Local == "case_sensitive" {
			p.CaseSensitive = a.Value
		}
	}
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return nil
			}
		case xml.StartElement:
			switch t.Name.Local {
			case "token":
				xt, err := decodeDisambigToken(d, t)
				if err != nil {
					return err
				}
				p.Tokens = append(p.Tokens, xt)
			case "marker":
				// Java <marker> wraps tokens that REPLACE/FILTER target (InsideMarker).
				if err := p.decodeMarkerContents(d); err != nil {
					return err
				}
			case "and":
				base, err := decodeDisambigAnd(d, t)
				if err != nil {
					return err
				}
				if base != nil {
					p.Tokens = append(p.Tokens, *base)
				}
			case "unify":
				// Java <unify feature/type tokens> — must not Skip (would drop tokens).
				if err := p.decodeDisambigUnify(d, t); err != nil {
					return err
				}
			default:
				if err := d.Skip(); err != nil {
					return err
				}
			}
		}
	}
}

func (p *disambigPattern) decodeMarkerContents(d *xml.Decoder) error {
	for {
		inner, err := d.Token()
		if err != nil {
			return err
		}
		switch it := inner.(type) {
		case xml.EndElement:
			if it.Name.Local == "marker" {
				return nil
			}
		case xml.StartElement:
			switch it.Name.Local {
			case "token":
				xt, err := decodeDisambigToken(d, it)
				if err != nil {
					return err
				}
				xt.Marker = "yes"
				p.Tokens = append(p.Tokens, xt)
			case "and":
				base, err := decodeDisambigAnd(d, it)
				if err != nil {
					return err
				}
				if base != nil {
					base.Marker = "yes"
					p.Tokens = append(p.Tokens, *base)
				}
			case "unify":
				// unify inside marker: tokens are InsideMarker
				startN := len(p.Tokens)
				if err := p.decodeDisambigUnify(d, it); err != nil {
					return err
				}
				for i := startN; i < len(p.Tokens); i++ {
					p.Tokens[i].Marker = "yes"
				}
			default:
				if err := d.Skip(); err != nil {
					return err
				}
			}
		}
	}
}

// decodeDisambigUnify ports PatternRuleLoader.decodeXMLUnify for disambiguation patterns.
// Sets UniFeatures / LastInUnification / UniNegated / UnificationNeutral on tokens.
func (p *disambigPattern) decodeDisambigUnify(d *xml.Decoder, start xml.StartElement) error {
	uniNeg := false
	for _, a := range start.Attr {
		if a.Name.Local == "negate" && strings.EqualFold(a.Value, "yes") {
			uniNeg = true
		}
	}
	// feature id → type ids (empty slice = all types for feature, Java).
	features := map[string][]string{}
	var collected []disambigToken

	appendTok := func(xt disambigToken, neutral bool) {
		xt.UniFeatures = copyDisambigFeatureMap(features)
		xt.UnificationNeutral = neutral
		collected = append(collected, xt)
	}

	parseIgnore := func() error {
		for {
			inner, err := d.Token()
			if err != nil {
				return err
			}
			switch it := inner.(type) {
			case xml.EndElement:
				if it.Name.Local == "unify-ignore" {
					return nil
				}
			case xml.StartElement:
				switch it.Name.Local {
				case "token":
					xt, err := decodeDisambigToken(d, it)
					if err != nil {
						return err
					}
					appendTok(xt, true)
				case "and":
					base, err := decodeDisambigAnd(d, it)
					if err != nil {
						return err
					}
					if base != nil {
						appendTok(*base, true)
					}
				case "marker":
					if err := decodeDisambigUnifyMarker(d, true, appendTok); err != nil {
						return err
					}
				default:
					if err := d.Skip(); err != nil {
						return err
					}
				}
			}
		}
	}

	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				if len(collected) > 0 {
					last := &collected[len(collected)-1]
					last.LastInUnification = true
					if uniNeg {
						last.UniNegated = true
					}
				}
				p.Tokens = append(p.Tokens, collected...)
				return nil
			}
		case xml.StartElement:
			switch t.Name.Local {
			case "feature":
				featID, types, err := decodeDisambigUnifyFeature(d, t)
				if err != nil {
					return err
				}
				if featID != "" {
					features[featID] = types
				}
			case "token":
				xt, err := decodeDisambigToken(d, t)
				if err != nil {
					return err
				}
				appendTok(xt, false)
			case "and":
				base, err := decodeDisambigAnd(d, t)
				if err != nil {
					return err
				}
				if base != nil {
					appendTok(*base, false)
				}
			case "unify-ignore":
				if err := parseIgnore(); err != nil {
					return err
				}
			case "marker":
				if err := decodeDisambigUnifyMarker(d, false, appendTok); err != nil {
					return err
				}
			default:
				if err := d.Skip(); err != nil {
					return err
				}
			}
		}
	}
}

func copyDisambigFeatureMap(in map[string][]string) map[string][]string {
	if len(in) == 0 {
		return map[string][]string{}
	}
	out := make(map[string][]string, len(in))
	for k, v := range in {
		out[k] = append([]string(nil), v...)
	}
	return out
}

func decodeDisambigUnifyFeature(d *xml.Decoder, start xml.StartElement) (id string, types []string, err error) {
	for _, a := range start.Attr {
		if a.Name.Local == "id" {
			id = a.Value
		}
	}
	types = []string{}
	for {
		tok, e := d.Token()
		if e != nil {
			return id, types, e
		}
		switch t := tok.(type) {
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return id, types, nil
			}
		case xml.StartElement:
			if t.Name.Local == "type" {
				typeID := ""
				for _, a := range t.Attr {
					if a.Name.Local == "id" {
						typeID = a.Value
					}
				}
				if err := drainDisambigElement(d, t.Name.Local); err != nil {
					return id, types, err
				}
				if typeID != "" {
					types = append(types, typeID)
				}
			} else if err := d.Skip(); err != nil {
				return id, types, err
			}
		}
	}
}

func drainDisambigElement(d *xml.Decoder, name string) error {
	depth := 1
	for depth > 0 {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == name {
				depth++
			}
		case xml.EndElement:
			if t.Name.Local == name {
				depth--
			}
		}
	}
	return nil
}

func decodeDisambigUnifyMarker(d *xml.Decoder, neutral bool, appendTok func(disambigToken, bool)) error {
	for {
		inner, err := d.Token()
		if err != nil {
			return err
		}
		switch it := inner.(type) {
		case xml.EndElement:
			if it.Name.Local == "marker" {
				return nil
			}
		case xml.StartElement:
			switch it.Name.Local {
			case "token":
				xt, err := decodeDisambigToken(d, it)
				if err != nil {
					return err
				}
				xt.Marker = "yes"
				appendTok(xt, neutral)
			case "and":
				base, err := decodeDisambigAnd(d, it)
				if err != nil {
					return err
				}
				if base != nil {
					base.Marker = "yes"
					appendTok(*base, neutral)
				}
			default:
				if err := d.Skip(); err != nil {
					return err
				}
			}
		}
	}
}

func decodeDisambigToken(d *xml.Decoder, start xml.StartElement) (disambigToken, error) {
	var xt disambigToken
	// Track whether case_sensitive was set on the token element (for pattern inherit).
	for _, a := range start.Attr {
		if a.Name.Local == "case_sensitive" {
			xt.caseSensitiveSet = true
			break
		}
	}
	err := d.DecodeElement(&xt, &start)
	// DecodeElement may clear unexported fields depending on version; re-check attrs after.
	if !xt.caseSensitiveSet {
		for _, a := range start.Attr {
			if a.Name.Local == "case_sensitive" {
				xt.caseSensitiveSet = true
				break
			}
		}
	}
	return xt, err
}

// decodeDisambigAnd reads Java <and><token>…</token></and> as one position with AndGroup.
func decodeDisambigAnd(d *xml.Decoder, start xml.StartElement) (*disambigToken, error) {
	var andToks []disambigToken
	for {
		inner, err := d.Token()
		if err != nil {
			return nil, err
		}
		switch it := inner.(type) {
		case xml.EndElement:
			if it.Name.Local == start.Name.Local {
				if len(andToks) == 0 {
					return nil, nil
				}
				base := andToks[0]
				base.AndTokens = append(base.AndTokens, andToks[1:]...)
				return &base, nil
			}
		case xml.StartElement:
			if it.Name.Local == "token" {
				xt, err := decodeDisambigToken(d, it)
				if err != nil {
					return nil, err
				}
				andToks = append(andToks, xt)
			} else if err := d.Skip(); err != nil {
				return nil, err
			}
		}
	}
}

type disambigToken struct {
	Regexp        string `xml:"regexp,attr"`
	CaseSensitive string `xml:"case_sensitive,attr"`
	Inflected     string `xml:"inflected,attr"`
	Negate        string `xml:"negate,attr"`
	// NegatePos ports negate_pos="yes" on the token POS itself (not only exception).
	NegatePos    string `xml:"negate_pos,attr"`
	Postag       string `xml:"postag,attr"`
	PostagRegexp string `xml:"postag_regexp,attr"`
	// Marker is soft extract for Java <marker>…</marker> (which tokens REPLACE/FILTER target).
	Marker string `xml:"marker,attr"`
	// SpaceBefore ports spacebefore="yes|no".
	SpaceBefore string `xml:"spacebefore,attr"`
	// Min/Max/Skip port Java PatternToken min/max/skip (PatternRuleMatcher).
	Min  string `xml:"min,attr"`
	Max  string `xml:"max,attr"`
	Skip string `xml:"skip,attr"`
	// Chunk / ChunkRe port Java PatternToken chunk / chunk_re (ChunkTag).
	Chunk   string `xml:"chunk,attr"`
	ChunkRe string `xml:"chunk_re,attr"`
	// Exceptions ports Java <exception> under <token> (first positive used).
	Exceptions []disambigException `xml:"exception"`
	// AndTokens ports soft <and_token> = Java <and> group members.
	AndTokens []disambigToken `xml:"and_token"`
	// Match ports Java pattern-token <match no="…" setpos=…/> (tokenReference).
	Match   *disambigMatch `xml:"match"`
	Content string         `xml:",chardata"`
	// Unification fields set by decodeDisambigUnify (not XML attrs on <token>).
	UniFeatures        map[string][]string
	LastInUnification  bool
	UnificationNeutral bool
	UniNegated         bool
	// caseSensitiveSet is true when the token XML set case_sensitive= (for pattern inherit).
	caseSensitiveSet bool
}

// disambigException ports Java pattern-token <exception> attributes used by
// DisambiguationRuleHandler / XMLRuleHandler (not invent soft attrs).
type disambigException struct {
	Regexp        string `xml:"regexp,attr"`
	CaseSensitive string `xml:"case_sensitive,attr"`
	Negate        string `xml:"negate,attr"`
	// NegatePos ports negate_pos="yes" (POS exception polarity).
	NegatePos string `xml:"negate_pos,attr"`
	// Inflected ports inflected="yes" on exception (lemma match path).
	Inflected string `xml:"inflected,attr"`
	// SpaceBefore ports spacebefore="yes|no" on exception scope.
	SpaceBefore  string `xml:"spacebefore,attr"`
	Scope        string `xml:"scope,attr"` // previous|next|empty=current
	Postag       string `xml:"postag,attr"`
	PostagRegexp string `xml:"postag_regexp,attr"`
	Content      string `xml:",chardata"`
}

type disambigElem struct {
	Action   string       `xml:"action,attr"`
	Postag   string       `xml:"postag,attr"`
	Features string       `xml:"features,attr"` // UNIFY: comma-separated feature ids
	Wds      []disambigWd `xml:"wd"`
	// Match ports Java <disambig><match no="…" postag=…/></disambig> (posSelector).
	// At most one match child is used (Java posSelector).
	Match *disambigMatch `xml:"match"`
}

// disambigMatch ports Match attributes under disambiguation <disambig>.
type disambigMatch struct {
	No                 string `xml:"no,attr"`
	Postag             string `xml:"postag,attr"`
	PostagReplace      string `xml:"postag_replace,attr"`
	PostagRegexp       string `xml:"postag_regexp,attr"`
	RegexpMatch        string `xml:"regexp_match,attr"`
	RegexpReplace      string `xml:"regexp_replace,attr"`
	CaseConversion     string `xml:"case_conversion,attr"`
	IncludeSkipped     string `xml:"include_skipped,attr"`
	SetPos             string `xml:"setpos,attr"`
	SuppressMisspelled string `xml:"suppress_mispelled,attr"` // Java spelling of attr
	// Content is lemma string body: <match no="1">рада</match>
	Content string `xml:",chardata"`
}

// disambigWd ports <wd pos="…" lemma="…"/> under <disambig action="add">.
type disambigWd struct {
	Pos     string `xml:"pos,attr"`
	Lemma   string `xml:"lemma,attr"`
	Content string `xml:",chardata"`
}

func (l *DisambiguationRuleLoader) parse(data []byte, languageCode, xmlPath string) ([]*DisambiguationPatternRule, *patterns.UnifierConfiguration, error) {
	var root disambigRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, nil, fmt.Errorf("parse disambiguation %s: %w", xmlPath, err)
	}
	cfg := patterns.NewUnifierConfiguration()
	for _, u := range root.Unifications {
		feat := tools.JavaStringTrim(u.Feature)
		if feat == "" {
			continue
		}
		for _, eq := range u.Equivalences {
			typ := tools.JavaStringTrim(eq.Type)
			if typ == "" {
				continue
			}
			pt := disambigTokenFromXML(eq.Token, false, false)
			if pt != nil {
				cfg.SetEquivalence(feat, typ, pt)
			}
		}
	}
	// Flatten top-level rules + rulegroup nested rules (Java DisambiguationRuleHandler).
	// Rulegroup: inherit id/name when missing; share rulegroup-level antipatterns.
	var out []*DisambiguationPatternRule
	for _, xr := range root.Rules {
		rule := buildDisambiguationPatternRule(xr, languageCode, cfg, nil)
		if rule == nil {
			continue
		}
		out = append(out, rule)
	}
	for _, g := range root.RuleGroups {
		groupAPs := g.AntiPatterns
		for i, xr := range g.Rules {
			// Java: if inRuleGroup && id/name == null → use ruleGroupId/Name.
			if tools.JavaStringTrim(xr.ID) == "" {
				xr.ID = g.ID
			}
			if tools.JavaStringTrim(xr.Name) == "" {
				xr.Name = g.Name
			}
			// subId is 1-based within the group (Java subId++ on each rule start).
			rule := buildDisambiguationPatternRule(xr, languageCode, cfg, groupAPs)
			if rule == nil {
				continue
			}
			if rule.PatternRule != nil {
				// Java: setSubId(inRuleGroup ? Integer.toString(subId) : "1")
				rule.SubID = fmt.Sprintf("%d", i+1)
			}
			out = append(out, rule)
		}
	}
	return out, cfg, nil
}

// buildDisambiguationPatternRule converts one XML rule. Skips rules with empty patterns
// (would match everything — not Java-faithful; usually a parse bug).
// groupAntiPatterns are rulegroup-shared antipatterns (Java rulegroupAntiPatterns); may be nil.
func buildDisambiguationPatternRule(xr disambigRule, languageCode string, cfg *patterns.UnifierConfiguration, groupAntiPatterns []disambigPattern) *DisambiguationPatternRule {
	var tokens []*patterns.PatternToken
	anyMarker := false
	for _, xt := range xr.Pattern.Tokens {
		if strings.EqualFold(xt.Marker, "yes") {
			anyMarker = true
		}
	}
	// Java: pattern-level case_sensitive inherits when token has no own case_sensitive.
	patternCS := strings.EqualFold(xr.Pattern.CaseSensitive, "yes")
	for _, xt := range xr.Pattern.Tokens {
		tokens = append(tokens, disambigTokenFromXML(xt, anyMarker, patternCS))
	}
	if len(tokens) == 0 {
		return nil
	}
	// Java XMLRuleHandler.setRuleFilter: both class and args non-null to attach.
	// Unknown filter class → skip rule (fail-closed; never disambiguate without the filter gate).
	// Same policy as PatternRuleLoader for unsupported filters.
	if xr.Filter != nil && tools.JavaStringTrim(xr.Filter.Class) != "" {
		if tools.JavaStringTrim(xr.Filter.Args) == "" {
			// Java: filterArgs null → setRuleFilter no-op; still load rule without filter.
		} else if !patterns.GlobalRuleFilterCreator.HasFilter(tools.JavaStringTrim(xr.Filter.Class)) {
			return nil
		}
	}
	action := ActionReplace
	if xr.Disambig.Action != "" {
		action = DisambiguatorAction(strings.ToUpper(xr.Disambig.Action))
	}
	// Java DisambiguationRuleHandler: <match> under <disambig> → posSelector
	posSelect := matchFromDisambigXML(xr.Disambig.Match)
	// default Java: REPLACE when only postag set (or match element)
	rule := NewDisambiguationPatternRule(xr.ID, xr.Name, languageCode, tokens, xr.Disambig.Postag, posSelect, action)
	rule.UnifierConfig = cfg
	// Matching (PatternRuleMatcher.testUnification) reads PatternRule.UnifierConfig.
	// DisambiguationPatternRule.UnifierConfig is a separate field used by applyUnify fallback.
	if rule.PatternRule != nil {
		rule.PatternRule.UnifierConfig = cfg
	}
	// Java prepareRule: start/end position corrections from <marker>
	if rule.PatternRule != nil {
		startCorr, endCorr := patterns.PositionCorrectionsFromTokens(tokens)
		rule.StartPositionCorrection = startCorr
		rule.EndPositionCorrection = endCorr
		// Java: setSubId(inRuleGroup ? … : "1"); rulegroup path overwrites SubID after build.
		if rule.SubID == "" {
			rule.SubID = "1"
		}
	}
	// Java setRuleFilter(filterClassName, filterArgs, rule) for DisambiguationPatternRule.
	if xr.Filter != nil {
		class := tools.JavaStringTrim(xr.Filter.Class)
		args := tools.JavaStringTrim(xr.Filter.Args)
		if class != "" && args != "" {
			if f, ok := patterns.GlobalRuleFilterCreator.TryGetFilter(class); ok {
				rule.Filter = f
				rule.FilterArgs = args
			}
		}
	}
	if action == ActionUnify {
		for _, f := range strings.Split(xr.Disambig.Features, ",") {
			f = tools.JavaStringTrim(f)
			if f != "" {
				rule.UnifyFeatures = append(rule.UnifyFeatures, f)
			}
		}
	}
	// Java ADD/REMOVE with <wd pos="…" lemma="…"/>.
	if len(xr.Disambig.Wds) > 0 {
		var readings []*languagetool.AnalyzedToken
		for _, wd := range xr.Disambig.Wds {
			surf := tools.JavaStringTrim(wd.Content)
			var posPtr, lemmaPtr *string
			if wd.Pos != "" {
				p := wd.Pos
				posPtr = &p
			}
			if wd.Lemma != "" {
				lm := wd.Lemma
				lemmaPtr = &lm
			}
			readings = append(readings, languagetool.NewAnalyzedToken(surf, posPtr, lemmaPtr))
		}
		rule.SetNewInterpretations(readings)
	}
	// Java: rulegroup antipatterns first, then rule antipatterns (setAntiPatterns appends).
	if len(groupAntiPatterns) > 0 {
		if aps := antiPatternsFromDisambigXML(xr.ID, languageCode, groupAntiPatterns, "group_anti"); len(aps) > 0 {
			rule.SetAntiPatterns(aps)
		}
	}
	if len(xr.AntiPatterns) > 0 {
		if aps := antiPatternsFromDisambigXML(xr.ID, languageCode, xr.AntiPatterns, "anti"); len(aps) > 0 {
			rule.SetAntiPatterns(aps)
		}
	}
	// Java DisambiguationRuleHandler EXAMPLE: untouched vs ambiguous (inputform/outputform).
	if len(xr.Examples) > 0 {
		var disEx []DisambiguatedExample
		var untouched []string
		for _, ex := range xr.Examples {
			body := tools.JavaStringTrim(ex.Inner)
			// Normalize whitespace like typical SAX char accumulation (collapse pure indent).
			switch strings.ToLower(tools.JavaStringTrim(ex.Type)) {
			case "untouched":
				if body != "" {
					untouched = append(untouched, body)
				}
			case "ambiguous":
				disEx = append(disEx, NewDisambiguatedExampleFull(body, ex.InputForm, ex.OutputForm))
			default:
				// Java only branches untouched vs else (ambiguous); other types go to disambExamples.
				if body != "" || ex.InputForm != "" || ex.OutputForm != "" {
					disEx = append(disEx, NewDisambiguatedExampleFull(body, ex.InputForm, ex.OutputForm))
				}
			}
		}
		if len(disEx) > 0 {
			rule.SetExamples(disEx)
		}
		if len(untouched) > 0 {
			rule.SetUntouchedExamples(untouched)
		}
	}
	return rule
}

// antiPatternsFromDisambigXML builds antipattern token rules (shared by rule + rulegroup).
func antiPatternsFromDisambigXML(ruleID, languageCode string, patternsXML []disambigPattern, idSuffix string) []*patterns.AbstractTokenBasedRule {
	var aps []*patterns.AbstractTokenBasedRule
	for i, ap := range patternsXML {
		var apToks []*patterns.PatternToken
		anyMarker := false
		for _, xt := range ap.Tokens {
			if strings.EqualFold(xt.Marker, "yes") {
				anyMarker = true
			}
		}
		patternCS := strings.EqualFold(ap.CaseSensitive, "yes")
		for _, xt := range ap.Tokens {
			apToks = append(apToks, disambigTokenFromXML(xt, anyMarker, patternCS))
		}
		if len(apToks) == 0 {
			continue
		}
		aps = append(aps, patterns.NewAbstractTokenBasedRule(
			fmt.Sprintf("%s_%s_%d", ruleID, idSuffix, i),
			"antipattern",
			languageCode,
			apToks,
		))
	}
	return aps
}

// matchFromDisambigXML ports DisambiguationRuleHandler MATCH under DISAMBIG → posSelector.
// Returns nil when no match child or missing no= attribute (Java only sets posSelector with no).
func matchFromDisambigXML(xm *disambigMatch) *patterns.Match {
	if xm == nil {
		return nil
	}
	noStr := tools.JavaStringTrim(xm.No)
	if noStr == "" {
		return nil
	}
	ref, err := strconv.Atoi(noStr)
	if err != nil {
		return nil
	}
	caseConv := patterns.CaseNone
	if v := tools.JavaStringTrim(xm.CaseConversion); v != "" {
		if c, ok := patterns.ParseCaseConversion(v); ok {
			caseConv = c
		}
	}
	include := patterns.IncludeNone
	if v := tools.JavaStringTrim(xm.IncludeSkipped); v != "" {
		if ir, ok := patterns.ParseIncludeRange(v); ok {
			include = ir
		}
	}
	postagRE := strings.EqualFold(xm.PostagRegexp, "yes")
	setPos := strings.EqualFold(xm.SetPos, "yes")
	// Java attribute is suppress_mispelled (one 's')
	suppress := strings.EqualFold(xm.SuppressMisspelled, "yes")
	m := patterns.NewMatch(
		tools.JavaStringTrim(xm.Postag),
		tools.JavaStringTrim(xm.PostagReplace),
		postagRE,
		tools.JavaStringTrim(xm.RegexpMatch),
		tools.JavaStringTrim(xm.RegexpReplace),
		caseConv,
		setPos,
		suppress,
		include,
	)
	m.SetTokenRef(ref)
	if body := tools.JavaStringTrim(xm.Content); body != "" {
		// Java: posSelector.setLemmaString(match.toString()) on endElement MATCH
		m.SetLemmaString(body)
	}
	return m
}

func disambigTokenFromXML(xt disambigToken, patternHasMarker bool, patternCaseSensitive bool) *patterns.PatternToken {
	content := tools.JavaStringTrim(xt.Content)
	// Java: token case_sensitive attr, else pattern-level case_sensitive inherit.
	cs := strings.EqualFold(xt.CaseSensitive, "yes")
	if !xt.caseSensitiveSet && patternCaseSensitive {
		cs = true
	}
	re := strings.EqualFold(xt.Regexp, "yes")
	inflected := strings.EqualFold(xt.Inflected, "yes")
	pt := patterns.NewPatternToken(content, cs, re, inflected)
	if strings.EqualFold(xt.Negate, "yes") {
		pt.SetNegation(true)
	}
	if xt.UniFeatures != nil {
		pt.SetUnification(xt.UniFeatures)
	}
	if xt.UniNegated {
		pt.SetUniNegation()
	}
	if xt.LastInUnification {
		pt.SetLastInUnification()
	}
	if xt.UnificationNeutral {
		pt.SetUnificationNeutral()
	}
	if xt.Postag != "" {
		pt.SetPosToken(patterns.PosToken{
			PosTag: xt.Postag,
			Regexp: strings.EqualFold(xt.PostagRegexp, "yes"),
			Negate: strings.EqualFold(xt.NegatePos, "yes"),
		})
	}
	// Java: default InsideMarker=true when the pattern has no <marker>.
	// When markers exist, only tokens inside <marker> are InsideMarker.
	if patternHasMarker {
		pt.InsideMarker = strings.EqualFold(xt.Marker, "yes")
	} else {
		pt.InsideMarker = true
	}
	if sb := tools.JavaStringTrim(xt.SpaceBefore); sb != "" {
		pt.SetWhitespaceBefore(strings.EqualFold(sb, "yes"))
	}
	if ch := tools.JavaStringTrim(xt.ChunkRe); ch != "" {
		pt.SetChunkTag(ch, true)
	} else if ch := tools.JavaStringTrim(xt.Chunk); ch != "" {
		pt.SetChunkTag(ch, false)
	}
	if xt.Min != "" {
		var n int
		if _, err := fmt.Sscanf(xt.Min, "%d", &n); err == nil {
			pt.SetMinOccurrence(n)
		}
	}
	if xt.Max != "" {
		var n int
		if _, err := fmt.Sscanf(xt.Max, "%d", &n); err == nil {
			pt.SetMaxOccurrence(n)
		}
	}
	if xt.Skip != "" {
		var n int
		if _, err := fmt.Sscanf(xt.Skip, "%d", &n); err == nil {
			pt.SetSkipNext(n)
		}
	}
	// Java PatternToken exceptions: setStringPosException → addException by scope (multi).
	for _, ex := range xt.Exceptions {
		exc := tools.JavaStringTrim(ex.Content)
		posTag := tools.JavaStringTrim(ex.Postag)
		if exc == "" && posTag == "" {
			continue
		}
		scope := strings.ToLower(tools.JavaStringTrim(ex.Scope))
		re := strings.EqualFold(ex.Regexp, "yes")
		cs := strings.EqualFold(ex.CaseSensitive, "yes")
		posRE := strings.EqualFold(ex.PostagRegexp, "yes")
		neg := strings.EqualFold(ex.Negate, "yes")
		posNeg := strings.EqualFold(ex.NegatePos, "yes")
		infl := strings.EqualFold(ex.Inflected, "yes")
		if sb := tools.JavaStringTrim(ex.SpaceBefore); sb != "" && scope == "" {
			// exception-level spacebefore applies to current exception token context
			pt.SetWhitespaceBefore(strings.EqualFold(sb, "yes"))
		}
		exTok := patterns.NewPatternToken(exc, cs, re, infl)
		exTok.SetNegation(neg)
		if posTag != "" {
			exTok.SetPosToken(patterns.PosToken{PosTag: posTag, Regexp: posRE, Negate: posNeg})
		}
		// Java setExceptionSpaceBefore → exception.setWhitespaceBefore
		if sb := tools.JavaStringTrim(ex.SpaceBefore); sb != "" && !strings.EqualFold(sb, "ignore") {
			exTok.SetWhitespaceBefore(strings.EqualFold(sb, "yes"))
		}
		switch scope {
		case "previous":
			pt.AddPreviousException(exTok)
		case "next":
			pt.AddNextException(exTok)
		default:
			pt.AddCurrentException(exTok)
		}
	}
	// Java <and> group members (also soft <and_token> attribute path): each must match some reading.
	for _, at := range xt.AndTokens {
		pt.AddAndGroupElement(disambigTokenFromXML(at, false, patternCaseSensitive))
	}
	// Java MATCH inside TOKEN (not under DISAMBIG): tokenReference / setpos.
	if m := matchFromTokenMatchXML(xt.Match); m != nil {
		pt.SetMatch(m)
		// Java: appends \\N into token string for reference elements.
		if content == "" && m.GetTokenRef() >= 0 {
			pt.Token = fmt.Sprintf("\\%d", m.GetTokenRef())
		}
	}
	return pt
}

// matchFromTokenMatchXML ports pattern-token <match> (same attrs as disambig match).
// Unlike matchFromDisambigXML, no= is optional for setpos-only; TokenRef still set when present.
// Java XMLRuleHandler: ref uses raw no= as offset from firstMatchToken.
func matchFromTokenMatchXML(xm *disambigMatch) *patterns.Match {
	if xm == nil {
		return nil
	}
	caseConv := patterns.CaseNone
	if v := tools.JavaStringTrim(xm.CaseConversion); v != "" {
		if c, ok := patterns.ParseCaseConversion(v); ok {
			caseConv = c
		}
	}
	include := patterns.IncludeNone
	if v := tools.JavaStringTrim(xm.IncludeSkipped); v != "" {
		if ir, ok := patterns.ParseIncludeRange(v); ok {
			include = ir
		}
	}
	postagRE := strings.EqualFold(xm.PostagRegexp, "yes")
	setPos := strings.EqualFold(xm.SetPos, "yes")
	suppress := strings.EqualFold(xm.SuppressMisspelled, "yes")
	m := patterns.NewMatch(
		tools.JavaStringTrim(xm.Postag),
		tools.JavaStringTrim(xm.PostagReplace),
		postagRE,
		tools.JavaStringTrim(xm.RegexpMatch),
		tools.JavaStringTrim(xm.RegexpReplace),
		caseConv,
		setPos,
		suppress,
		include,
	)
	if noStr := tools.JavaStringTrim(xm.No); noStr != "" {
		if ref, err := strconv.Atoi(noStr); err == nil {
			m.SetTokenRef(ref)
		}
	}
	if body := tools.JavaStringTrim(xm.Content); body != "" {
		m.SetLemmaString(body)
	}
	return m
}
