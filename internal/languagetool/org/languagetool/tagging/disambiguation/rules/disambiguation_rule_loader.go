package rules

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
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
	// Official LT disambiguation.xml uses custom DTD entities.
	data = patterns.ExpandLTXMLEntities(data)
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
	XMLName       xml.Name              `xml:"rules"`
	Unifications  []disambigUnification `xml:"unification"`
	Rules         []disambigRule        `xml:"rule"`
	// RuleGroups: Java nests many rules under <rulegroup> (not visible as top-level <rule>).
	RuleGroups []disambigRuleGroup `xml:"rulegroup"`
}

// disambigRuleGroup ports <rulegroup id="…" name="…"> containing nested <rule>.
type disambigRuleGroup struct {
	ID    string         `xml:"id,attr"`
	Name  string         `xml:"name,attr"`
	Rules []disambigRule `xml:"rule"`
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

type disambigRule struct {
	ID           string            `xml:"id,attr"`
	Name         string            `xml:"name,attr"`
	AntiPatterns []disambigPattern `xml:"antipattern"`
	Pattern      disambigPattern   `xml:"pattern"`
	Disambig     disambigElem      `xml:"disambig"`
}

// disambigPattern holds pattern children in document order: <token> and/or <and>.
// Java PatternRuleHandler walks elements; empty tokens from skipped <and> must not load.
type disambigPattern struct {
	// Tokens is filled by UnmarshalXML (ordered, includes synthetic AND-group tokens).
	Tokens []disambigToken
}

// UnmarshalXML ports pattern content: sequence of <token> and <and><token>…</and>.
func (p *disambigPattern) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	p.Tokens = nil
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
				var xt disambigToken
				if err := d.DecodeElement(&xt, &t); err != nil {
					return err
				}
				p.Tokens = append(p.Tokens, xt)
			case "and":
				// Java <and> is one pattern position: first token is base, rest AndGroup.
				var andToks []disambigToken
				for {
					inner, err := d.Token()
					if err != nil {
						return err
					}
					switch it := inner.(type) {
					case xml.EndElement:
						if it.Name.Local == "and" {
							goto andDone
						}
					case xml.StartElement:
						if it.Name.Local == "token" {
							var xt disambigToken
							if err := d.DecodeElement(&xt, &it); err != nil {
								return err
							}
							andToks = append(andToks, xt)
						} else if err := d.Skip(); err != nil {
							return err
						}
					}
				}
			andDone:
				if len(andToks) == 0 {
					continue
				}
				base := andToks[0]
				// Sibling <token>s inside <and> become AndTokens (Java and-group).
				base.AndTokens = append(base.AndTokens, andToks[1:]...)
				p.Tokens = append(p.Tokens, base)
			default:
				if err := d.Skip(); err != nil {
					return err
				}
			}
		}
	}
}

type disambigToken struct {
	Regexp        string `xml:"regexp,attr"`
	CaseSensitive string `xml:"case_sensitive,attr"`
	Inflected     string `xml:"inflected,attr"`
	Negate        string `xml:"negate,attr"`
	Postag        string `xml:"postag,attr"`
	PostagRegexp  string `xml:"postag_regexp,attr"`
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
	Content   string          `xml:",chardata"`
}

// disambigException is a soft subset of Java pattern-token <exception>.
type disambigException struct {
	Regexp        string `xml:"regexp,attr"`
	CaseSensitive string `xml:"case_sensitive,attr"`
	Negate        string `xml:"negate,attr"`
	Scope         string `xml:"scope,attr"` // previous|next|empty=current
	Postag        string `xml:"postag,attr"`
	PostagRegexp  string `xml:"postag_regexp,attr"`
	Content       string `xml:",chardata"`
}

type disambigElem struct {
	Action   string       `xml:"action,attr"`
	Postag   string       `xml:"postag,attr"`
	Features string       `xml:"features,attr"` // soft UNIFY: comma-separated feature ids
	Wds      []disambigWd `xml:"wd"`
}

// disambigWd ports <wd pos="…" lemma="…"/> under <disambig action="add">.
type disambigWd struct {
	Pos    string `xml:"pos,attr"`
	Lemma  string `xml:"lemma,attr"`
	Content string `xml:",chardata"`
}

func (l *DisambiguationRuleLoader) parse(data []byte, languageCode, xmlPath string) ([]*DisambiguationPatternRule, *patterns.UnifierConfiguration, error) {
	var root disambigRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, nil, fmt.Errorf("parse disambiguation %s: %w", xmlPath, err)
	}
	cfg := patterns.NewUnifierConfiguration()
	for _, u := range root.Unifications {
		feat := strings.TrimSpace(u.Feature)
		if feat == "" {
			continue
		}
		for _, eq := range u.Equivalences {
			typ := strings.TrimSpace(eq.Type)
			if typ == "" {
				continue
			}
			pt := disambigTokenFromXML(eq.Token, false)
			if pt != nil {
				cfg.SetEquivalence(feat, typ, pt)
			}
		}
	}
	// Flatten top-level rules + rulegroup nested rules (Java DisambiguationRuleHandler).
	var flat []disambigRule
	flat = append(flat, root.Rules...)
	for _, g := range root.RuleGroups {
		flat = append(flat, g.Rules...)
	}

	var out []*DisambiguationPatternRule
	for _, xr := range flat {
		rule := buildDisambiguationPatternRule(xr, languageCode, cfg)
		if rule == nil {
			continue
		}
		out = append(out, rule)
	}
	return out, cfg, nil
}

// buildDisambiguationPatternRule converts one XML rule. Skips rules with empty patterns
// (would match everything — not Java-faithful; usually a parse bug).
func buildDisambiguationPatternRule(xr disambigRule, languageCode string, cfg *patterns.UnifierConfiguration) *DisambiguationPatternRule {
	var tokens []*patterns.PatternToken
	anyMarker := false
	for _, xt := range xr.Pattern.Tokens {
		if strings.EqualFold(xt.Marker, "yes") {
			anyMarker = true
		}
	}
	for _, xt := range xr.Pattern.Tokens {
		tokens = append(tokens, disambigTokenFromXML(xt, anyMarker))
	}
	if len(tokens) == 0 {
		return nil
	}
	action := ActionReplace
	if xr.Disambig.Action != "" {
		action = DisambiguatorAction(strings.ToUpper(xr.Disambig.Action))
	}
	// default Java: REPLACE when only postag set
	rule := NewDisambiguationPatternRule(xr.ID, xr.Name, languageCode, tokens, xr.Disambig.Postag, nil, action)
	rule.UnifierConfig = cfg
	if action == ActionUnify {
		for _, f := range strings.Split(xr.Disambig.Features, ",") {
			f = strings.TrimSpace(f)
			if f != "" {
				rule.UnifyFeatures = append(rule.UnifyFeatures, f)
			}
		}
	}
	// Java ADD/REMOVE with <wd pos="…" lemma="…"/>.
	if len(xr.Disambig.Wds) > 0 {
		var readings []*languagetool.AnalyzedToken
		for _, wd := range xr.Disambig.Wds {
			surf := strings.TrimSpace(wd.Content)
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
	// Java <antipattern>.
	if len(xr.AntiPatterns) > 0 {
		var aps []*patterns.AbstractTokenBasedRule
		for i, ap := range xr.AntiPatterns {
			var apToks []*patterns.PatternToken
			anyMarker := false
			for _, xt := range ap.Tokens {
				if strings.EqualFold(xt.Marker, "yes") {
					anyMarker = true
				}
			}
			for _, xt := range ap.Tokens {
				apToks = append(apToks, disambigTokenFromXML(xt, anyMarker))
			}
			if len(apToks) == 0 {
				continue
			}
			aps = append(aps, patterns.NewAbstractTokenBasedRule(
				fmt.Sprintf("%s_anti_%d", xr.ID, i),
				"antipattern",
				languageCode,
				apToks,
			))
		}
		if len(aps) > 0 {
			rule.SetAntiPatterns(aps)
		}
	}
	return rule
}

func disambigTokenFromXML(xt disambigToken, patternHasMarker bool) *patterns.PatternToken {
	content := strings.TrimSpace(xt.Content)
	cs := strings.EqualFold(xt.CaseSensitive, "yes")
	re := strings.EqualFold(xt.Regexp, "yes")
	inflected := strings.EqualFold(xt.Inflected, "yes")
	pt := patterns.NewPatternToken(content, cs, re, inflected)
	if strings.EqualFold(xt.Negate, "yes") {
		pt.SetNegation(true)
	}
	if xt.Postag != "" {
		pt.SetPosToken(patterns.PosToken{
			PosTag: xt.Postag,
			Regexp: strings.EqualFold(xt.PostagRegexp, "yes"),
		})
	}
	// Java: default InsideMarker=true when the pattern has no <marker>.
	// When markers exist, only tokens inside <marker> are InsideMarker.
	if patternHasMarker {
		pt.InsideMarker = strings.EqualFold(xt.Marker, "yes")
	} else {
		pt.InsideMarker = true
	}
	if sb := strings.TrimSpace(xt.SpaceBefore); sb != "" {
		pt.SetWhitespaceBefore(strings.EqualFold(sb, "yes"))
	}
	if ch := strings.TrimSpace(xt.ChunkRe); ch != "" {
		pt.SetChunkTag(ch, true)
	} else if ch := strings.TrimSpace(xt.Chunk); ch != "" {
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
	// Java PatternToken exceptions: positive surface/regexp and/or postag blocks
	// the token after any reading matches (isExceptionMatchedCompletely).
	// Soft: first non-negated current exception; first scope=previous|next.
	for _, ex := range xt.Exceptions {
		exc := strings.TrimSpace(ex.Content)
		posTag := strings.TrimSpace(ex.Postag)
		if exc == "" && posTag == "" {
			continue
		}
		if strings.EqualFold(ex.Negate, "yes") {
			continue
		}
		scope := strings.ToLower(strings.TrimSpace(ex.Scope))
		re := strings.EqualFold(ex.Regexp, "yes")
		cs := strings.EqualFold(ex.CaseSensitive, "yes")
		posRE := strings.EqualFold(ex.PostagRegexp, "yes")
		if scope == "previous" {
			// scope=previous: surface only (soft subset; POS not on previous yet)
			if exc != "" && pt.PreviousException == "" {
				pt.SetPreviousException(exc, re, cs)
			}
			continue
		}
		if scope == "next" {
			if exc != "" && pt.NextException == "" {
				pt.SetNextException(exc, re, cs)
			}
			continue
		}
		if !pt.HasCurrentException() {
			pt.SetStringPosExceptionFull(exc, re, cs, posTag, posRE)
		}
	}
	// Java <and> group members (also soft <and_token> attribute path): each must match some reading.
	for _, at := range xt.AndTokens {
		pt.AddAndGroupElement(disambigTokenFromXML(at, false))
	}
	return pt
}
