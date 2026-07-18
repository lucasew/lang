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
	ID         string         `xml:"id,attr"`
	Name       string         `xml:"name,attr"`
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
	Default string     `xml:"default,attr"`
	Pattern xmlPattern `xml:"pattern"`
	// Message keeps inner XML so <suggestion>…</suggestion> and soft \N backrefs survive.
	Message xmlMessage `xml:"message"`
	Short   string     `xml:"short"`
}

type xmlMessage struct {
	Inner string `xml:",innerxml"`
}

// xmlPattern holds ordered pattern children: <token>, <marker>, <and>.
type xmlPattern struct {
	CaseSensitive string `xml:"case_sensitive,attr"`
	// Tokens filled by UnmarshalXML (document order).
	Tokens []xmlToken
}

// UnmarshalXML ports Java pattern children so <marker> / <and> are not dropped.
func (p *xmlPattern) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	p.Tokens = nil
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
				xt, err := decodeXMLToken(d, t)
				if err != nil {
					return err
				}
				p.Tokens = append(p.Tokens, xt)
			case "marker":
				if err := p.decodeXMLMarker(d); err != nil {
					return err
				}
			case "and":
				base, err := decodeXMLAnd(d, t)
				if err != nil {
					return err
				}
				if base != nil {
					p.Tokens = append(p.Tokens, *base)
				}
			default:
				if err := d.Skip(); err != nil {
					return err
				}
			}
		}
	}
}

func (p *xmlPattern) decodeXMLMarker(d *xml.Decoder) error {
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
				xt, err := decodeXMLToken(d, it)
				if err != nil {
					return err
				}
				// Marker attr used by disambig; grammar uses InsideMarker on PatternToken via loader.
				// Keep tokens; full InsideMarker for grammar replace is future work.
				p.Tokens = append(p.Tokens, xt)
			case "and":
				base, err := decodeXMLAnd(d, it)
				if err != nil {
					return err
				}
				if base != nil {
					p.Tokens = append(p.Tokens, *base)
				}
			default:
				if err := d.Skip(); err != nil {
					return err
				}
			}
		}
	}
}

func decodeXMLToken(d *xml.Decoder, start xml.StartElement) (xmlToken, error) {
	var xt xmlToken
	err := d.DecodeElement(&xt, &start)
	return xt, err
}

func decodeXMLAnd(d *xml.Decoder, start xml.StartElement) (*xmlToken, error) {
	var andToks []xmlToken
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
				xt, err := decodeXMLToken(d, it)
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

type xmlToken struct {
	Regexp        string         `xml:"regexp,attr"`
	CaseSensitive string         `xml:"case_sensitive,attr"`
	Negate        string         `xml:"negate,attr"`
	Inflected     string         `xml:"inflected,attr"`
	Min           string         `xml:"min,attr"`
	Max           string         `xml:"max,attr"`
	Skip          string         `xml:"skip,attr"`
	Postag        string         `xml:"postag,attr"`
	PostagRegexp  string         `xml:"postag_regexp,attr"`
	// SpaceBefore ports spacebefore="yes|no" (Java PatternToken.setWhitespaceBefore).
	SpaceBefore string `xml:"spacebefore,attr"`
	// Chunk / ChunkRe port Java PatternToken chunk / chunk_re.
	Chunk   string `xml:"chunk,attr"`
	ChunkRe string `xml:"chunk_re,attr"`
	Content    string         `xml:",chardata"`
	Exceptions []xmlException `xml:"exception"`
	// AndTokens ports soft <and_token> = Java <and> group members.
	AndTokens []xmlToken `xml:"and_token"`
}

type xmlException struct {
	Regexp        string `xml:"regexp,attr"`
	Negate        string `xml:"negate,attr"`
	CaseSensitive string `xml:"case_sensitive,attr"`
	Scope         string `xml:"scope,attr"` // previous|next|empty=current
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
	add := func(xr xmlRule, defaultID, catID, catName string) error {
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
		// Java: pattern-level case_sensitive inherits to tokens/exceptions
		// when the child does not set its own case_sensitive attribute.
		patternCS := strings.EqualFold(xr.Pattern.CaseSensitive, "yes")
		var tokens []*PatternToken
		for _, xt := range xr.Pattern.Tokens {
			if patternCS {
				if xt.CaseSensitive == "" {
					xt.CaseSensitive = "yes"
				}
				for i := range xt.Exceptions {
					if xt.Exceptions[i].CaseSensitive == "" {
						xt.Exceptions[i].CaseSensitive = "yes"
					}
				}
			}
			pt := tokenFromXML(xt)
			tokens = append(tokens, pt)
		}
		// Empty patterns would match everything — not Java-faithful; skip.
		if len(tokens) == 0 {
			return nil
		}
		r := NewAbstractPatternRule(id, name, languageCode, tokens, false)
		r.Message = strings.TrimSpace(xr.Message.Inner)
		r.ShortMessage = strings.TrimSpace(xr.Short)
		r.CategoryID = catID
		r.CategoryName = catName
		// soft: default="off" / default="temp_off" registers but starts disabled
		def := strings.ToLower(strings.TrimSpace(xr.Default))
		if def == "off" || def == "temp_off" {
			r.DefaultOff = true
		}
		out = append(out, r)
		return nil
	}
	for _, cat := range root.Categories {
		for _, xr := range cat.Rules {
			if err := add(xr, "", cat.ID, cat.Name); err != nil {
				return nil, err
			}
		}
		for _, g := range cat.RuleGroups {
			for i, xr := range g.Rules {
				id := xr.ID
				if id == "" {
					id = g.ID
				}
				if err := add(xr, id, cat.ID, cat.Name); err != nil {
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
		if err := add(xr, "", "", ""); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func tokenFromXML(xt xmlToken) *PatternToken {
	content := strings.TrimSpace(xt.Content)
	cs := strings.EqualFold(xt.CaseSensitive, "yes")
	re := strings.EqualFold(xt.Regexp, "yes")
	inflected := strings.EqualFold(xt.Inflected, "yes")
	pt := NewPatternToken(content, cs, re, inflected)
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
	if sb := strings.TrimSpace(xt.SpaceBefore); sb != "" {
		pt.SetWhitespaceBefore(strings.EqualFold(sb, "yes"))
	}
	if ch := strings.TrimSpace(xt.ChunkRe); ch != "" {
		pt.SetChunkTag(ch, true)
	} else if ch := strings.TrimSpace(xt.Chunk); ch != "" {
		pt.SetChunkTag(ch, false)
	}
	// Soft subset: current exception (surface and/or postag) + scope previous/next.
	// Java: isExceptionMatchedCompletely after any reading matches the token.
	for _, ex := range xt.Exceptions {
		exc := strings.TrimSpace(ex.Content)
		posTag := strings.TrimSpace(ex.Postag)
		if exc == "" && posTag == "" {
			continue
		}
		// LT negate="yes" on exception is inverted; soft only implements positive.
		if strings.EqualFold(ex.Negate, "yes") {
			continue
		}
		scope := strings.ToLower(strings.TrimSpace(ex.Scope))
		re := strings.EqualFold(ex.Regexp, "yes")
		cs := strings.EqualFold(ex.CaseSensitive, "yes")
		posRE := strings.EqualFold(ex.PostagRegexp, "yes")
		switch scope {
		case "previous":
			if exc != "" && pt.PreviousException == "" {
				pt.SetPreviousException(exc, re, cs)
			}
		case "next":
			if exc != "" && pt.NextException == "" {
				pt.SetNextException(exc, re, cs)
			}
		default:
			if !pt.HasCurrentException() {
				pt.SetStringPosExceptionFull(exc, re, cs, posTag, posRE)
			}
		}
	}
	// Java <and> group members (soft <and_token>).
	for _, at := range xt.AndTokens {
		pt.AddAndGroupElement(tokenFromXML(at))
	}
	return pt
}
