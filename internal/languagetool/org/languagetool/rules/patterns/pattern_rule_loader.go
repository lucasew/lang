package patterns

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PatternRuleLoader ports org.languagetool.rules.patterns.PatternRuleLoader
// for a simplified rules XML subset (full PatternRuleHandler deferred).
type PatternRuleLoader struct {
	RelaxedMode bool
	// LastUnifierConfig is filled by the most recent successful parse
	// (language-level <unification> tables from the same file).
	LastUnifierConfig *UnifierConfiguration
}

func NewPatternRuleLoader() *PatternRuleLoader { return &PatternRuleLoader{} }

func (l *PatternRuleLoader) SetRelaxedMode(v bool) { l.RelaxedMode = v }

// GetRulesFromReader parses a simplified pattern-rules XML stream.
// filename is used in error messages and stored as Rule.sourceFile.
func (l *PatternRuleLoader) GetRulesFromReader(r io.Reader, filename, languageCode string) ([]*AbstractPatternRule, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Cannot load or parse input stream of '%s': %w", filename, err)
	}
	rules, err := l.parseRulesXML(data, languageCode, filename)
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
	XMLName xml.Name `xml:"rules"`
	// IdPrefix ports Java rules idprefix="…" (e.g. L2_ on grammar-l2-de.xml).
	IdPrefix string `xml:"idprefix,attr"`
	// Premium ports rules premium="yes|no" file-level default (Java premiumFileAttribute).
	Premium string `xml:"premium,attr"`
	// Unifications ports top-level <unification feature="…"> equivalence tables.
	Unifications []xmlUnification `xml:"unification"`
	// Phrases ports top-level <phrases><phrase id="…"> (PatternRuleHandler).
	Phrases    *xmlPhrasesBlock `xml:"phrases"`
	Categories []xmlCategory    `xml:"category"`
	Rules      []xmlRule        `xml:"rule"` // allow top-level rules
}

// xmlPhrasesBlock holds <phrases> definitions.
type xmlPhrasesBlock struct {
	Phrases []xmlPhraseDef `xml:"phrase"`
}

// xmlPhraseDef is one <phrase id="…"> with tokens / includephrases / phraseref.
type xmlPhraseDef struct {
	ID    string `xml:"id,attr"`
	Steps []xmlPatternStep
}

// UnmarshalXML parses phrase body in document order (tokens, includephrases, phraseref).
func (p *xmlPhraseDef) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	p.Steps = nil
	for _, a := range start.Attr {
		if a.Name.Local == "id" {
			p.ID = a.Value
		}
	}
	return decodePatternSteps(d, start.Name.Local, func(st xmlPatternStep) {
		p.Steps = append(p.Steps, st)
	})
}

// xmlPatternStep is one pattern child: a token group or a phraseref.
type xmlPatternStep struct {
	Token     *xmlToken // set for token/and/or/unify-expanded tokens
	PhraseRef string    // set for <phraseref idref="…"/>
	// InMarker is true when this step was parsed inside <marker> (Java inMarker).
	InMarker bool
}

// xmlUnification ports <unification feature="number">…</unification>.
type xmlUnification struct {
	Feature      string           `xml:"feature,attr"`
	Equivalences []xmlEquivalence `xml:"equivalence"`
}

// xmlEquivalence ports <equivalence type="sg"><token …/></equivalence>.
type xmlEquivalence struct {
	Type  string    `xml:"type,attr"`
	Token *xmlToken `xml:"token"`
}

type xmlCategory struct {
	ID         string         `xml:"id,attr"`
	Name       string         `xml:"name,attr"`
	// Default ports category default="off"|"on" (Java Category onByDefault).
	Default string `xml:"default,attr"`
	// Type ports category type="misspelling|style|…" → rule LocQualityIssueType when rule omits type.
	Type string `xml:"type,attr"`
	// Prio ports category prio="N" (Java prioCategoryAttribute; 0 = unset).
	Prio string `xml:"prio,attr"`
	// Premium ports category premium="yes|no" (Java premiumCategoryAttribute).
	Premium string `xml:"premium,attr"`
	Tags       string         `xml:"tags,attr"`
	ToneTags   string         `xml:"tone_tags,attr"`
	// GoalSpecific ports is_goal_specific on category (inherited when rule omits it).
	GoalSpecific string        `xml:"is_goal_specific,attr"`
	Rules        []xmlRule     `xml:"rule"`
	RuleGroups   []xmlRuleGroup `xml:"rulegroup"`
}

type xmlRuleGroup struct {
	ID    string    `xml:"id,attr"`
	Name  string    `xml:"name,attr"`
	// Default ports rulegroup default="off"|"temp_off" (inherited by child rules).
	Default string `xml:"default,attr"`
	// Type ports rulegroup type="grammar|typographical|…" (Java ruleGroupIssueType).
	Type string `xml:"type,attr"`
	// Prio ports rulegroup prio="N" (Java prioRuleGroupAttribute; non-zero overrides category).
	Prio string `xml:"prio,attr"`
	// Premium ports rulegroup premium="yes|no" (Java premiumRuleGroupAttribute).
	Premium string `xml:"premium,attr"`
	// MinPrevMatches ports rulegroup min_prev_matches (Java ruleGroupMinPrevMatches).
	MinPrevMatches string `xml:"min_prev_matches,attr"`
	// DistanceTokens ports rulegroup distance_tokens (Java ruleGroupDistanceTokens).
	DistanceTokens string `xml:"distance_tokens,attr"`
	Tags  string    `xml:"tags,attr"`
	ToneTags string `xml:"tone_tags,attr"`
	GoalSpecific string `xml:"is_goal_specific,attr"`
	// URL ports rulegroup <url> inherited when child omits url (Java urlForRuleGroup).
	URL string `xml:"url"`
	// AntiPatterns on the rulegroup apply to every child rule (Java PatternRuleHandler).
	AntiPatterns []xmlPattern `xml:"antipattern"`
	Rules        []xmlRule    `xml:"rule"`
}

type xmlRule struct {
	ID      string     `xml:"id,attr"`
	Name    string     `xml:"name,attr"`
	Default string     `xml:"default,attr"`
	// Type ports rule type="…" → LocQualityIssueType (overrides rulegroup/category).
	Type string `xml:"type,attr"`
	// Prio ports rule prio="N" (Java prioRuleAttribute; non-zero overrides group/category).
	Prio string `xml:"prio,attr"`
	// Premium ports rule premium="yes|no" (overrides group/category/file).
	Premium string `xml:"premium,attr"`
	// MinPrevMatches ports rule min_prev_matches (inherits rulegroup when unset).
	MinPrevMatches string `xml:"min_prev_matches,attr"`
	// DistanceTokens ports rule distance_tokens (inherits rulegroup when unset).
	DistanceTokens string `xml:"distance_tokens,attr"`
	Tags    string     `xml:"tags,attr"`
	ToneTags string    `xml:"tone_tags,attr"`
	GoalSpecific string `xml:"is_goal_specific,attr"`
	Pattern xmlPattern `xml:"pattern"`
	// Message keeps inner XML so <suggestion>…</suggestion> and soft \N backrefs survive.
	Message xmlMessage `xml:"message"`
	Short   string     `xml:"short"`
	// URL ports rule <url> element (Java setUrl).
	URL string `xml:"url"`
	// Filter is Java <filter class="…"/> — not implemented for most classes.
	// Rules with an unsupported filter must not register (would match without filter = cheat).
	Filter *xmlFilter `xml:"filter"`
	// AntiPatterns ports Java <antipattern>; loaded and applied in PatternRule.Match
	// via keepByGrammarAntiPatterns (overlap suppress, same test as keepByDisambig).
	AntiPatterns []xmlPattern `xml:"antipattern"`
}

// xmlFilter ports <filter class="org.languagetool.…Filter" args="…"/>.
type xmlFilter struct {
	Class string `xml:"class,attr"`
	Args  string `xml:"args,attr"`
}

type xmlMessage struct {
	Inner string `xml:",innerxml"`
}

// xmlPattern holds ordered pattern children: <token>, <marker>, <and>, <unify>, <phraseref>.
type xmlPattern struct {
	CaseSensitive string `xml:"case_sensitive,attr"`
	// RawPos ports pattern raw_pos="yes" (match against pre-disambiguation tags).
	RawPos string `xml:"raw_pos,attr"`
	// Tokens filled by UnmarshalXML (document order) for non-phrase patterns.
	Tokens []xmlToken
	// Steps preserve phraseref order for expansion (Java preparePhrase / createRules).
	Steps []xmlPatternStep
	// HasUnify is true when the pattern (or antipattern) contains <unify>.
	HasUnify bool
	// HasMarker is true when the pattern contains <marker> (Java startPos tracking).
	// When true, only tokens with InMarker get PatternToken.InsideMarker.
	HasMarker bool
}

// UnmarshalXML ports Java pattern children so <marker> / <and> / <phraseref> are not dropped.
func (p *xmlPattern) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	p.Tokens = nil
	p.Steps = nil
	p.HasUnify = false
	for _, a := range start.Attr {
		switch a.Name.Local {
		case "case_sensitive":
			p.CaseSensitive = a.Value
		case "raw_pos":
			p.RawPos = a.Value
		}
	}
	return decodePatternSteps(d, start.Name.Local, func(st xmlPatternStep) {
		if st.InMarker {
			p.HasMarker = true
		}
		p.Steps = append(p.Steps, st)
		if st.Token != nil {
			if st.InMarker {
				st.Token.InMarker = true
			}
			p.Tokens = append(p.Tokens, *st.Token)
			// unify tokens set HasUnify via decodeXMLUnify on pattern — handled below
		}
		if st.PhraseRef != "" {
			// phrase steps are not plain tokens
		}
	})
}

// decodePatternSteps reads children until endLocal, calling emit for each step.
func decodePatternSteps(d *xml.Decoder, endLocal string, emit func(xmlPatternStep)) error {
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.EndElement:
			if t.Name.Local == endLocal {
				return nil
			}
		case xml.StartElement:
			switch t.Name.Local {
			case "token":
				xt, err := decodeXMLToken(d, t)
				if err != nil {
					return err
				}
				emit(xmlPatternStep{Token: &xt})
			case "marker":
				// Java XMLRuleHandler MARKER: inMarker=true for nested tokens.
				if err := decodePatternSteps(d, "marker", func(st xmlPatternStep) {
					st.InMarker = true
					if st.Token != nil {
						t := *st.Token
						t.InMarker = true
						st.Token = &t
					}
					emit(st)
				}); err != nil {
					return err
				}
			case "and":
				base, err := decodeXMLAnd(d, t)
				if err != nil {
					return err
				}
				if base != nil {
					emit(xmlPatternStep{Token: base})
				}
			case "or":
				base, err := decodeXMLOr(d, t)
				if err != nil {
					return err
				}
				if base != nil {
					emit(xmlPatternStep{Token: base})
				}
			case "unify":
				// decode into temporary pattern to reuse unify decoder
				var tmp xmlPattern
				if err := tmp.decodeXMLUnify(d, t); err != nil {
					return err
				}
				for i := range tmp.Tokens {
					xt := tmp.Tokens[i]
					emit(xmlPatternStep{Token: &xt})
				}
			case "phraseref":
				idref := ""
				for _, a := range t.Attr {
					if a.Name.Local == "idref" {
						idref = a.Value
					}
				}
				if err := d.Skip(); err != nil {
					return err
				}
				if idref != "" {
					emit(xmlPatternStep{PhraseRef: idref})
				}
			case "includephrases":
				// Java: container for phraseref only; children emitted as phraseref steps.
				if err := decodePatternSteps(d, "includephrases", emit); err != nil {
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

func decodeMarkerSteps(d *xml.Decoder, emit func(xmlPatternStep)) error {
	return decodePatternSteps(d, "marker", func(st xmlPatternStep) {
		st.InMarker = true
		if st.Token != nil {
			t := *st.Token
			t.InMarker = true
			st.Token = &t
		}
		emit(st)
	})
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
				// Java setInsideMarker(inMarker) while inside <marker>.
				xt.InMarker = true
				p.HasMarker = true
				p.Tokens = append(p.Tokens, xt)
			case "and":
				base, err := decodeXMLAnd(d, it)
				if err != nil {
					return err
				}
				if base != nil {
					p.Tokens = append(p.Tokens, *base)
				}
			case "or":
				base, err := decodeXMLOr(d, it)
				if err != nil {
					return err
				}
				if base != nil {
					p.Tokens = append(p.Tokens, *base)
				}
			case "unify":
				if err := p.decodeXMLUnify(d, it); err != nil {
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

// decodeXMLUnify ports Java <unify> / <unify-ignore> / feature / type handling
// (PatternRuleHandler + XMLRuleHandler.finalizeTokens setUnification).
func (p *xmlPattern) decodeXMLUnify(d *xml.Decoder, start xml.StartElement) error {
	p.HasUnify = true
	uniNeg := false
	for _, a := range start.Attr {
		if a.Name.Local == "negate" && strings.EqualFold(a.Value, "yes") {
			uniNeg = true
		}
	}
	// feature id → type ids (empty slice = all types for feature, Java).
	features := map[string][]string{}
	var collected []xmlToken

	appendTok := func(xt xmlToken, neutral bool) {
		// Snapshot features present when the token is closed (Java finalizeTokens).
		xt.UniFeatures = copyFeatureMap(features)
		xt.UnificationNeutral = neutral
		collected = append(collected, xt)
	}

	var parseIgnore func() error
	parseIgnore = func() error {
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
					xt, err := decodeXMLToken(d, it)
					if err != nil {
						return err
					}
					appendTok(xt, true)
				case "and":
					base, err := decodeXMLAnd(d, it)
					if err != nil {
						return err
					}
					if base != nil {
						appendTok(*base, true)
					}
				case "or":
					base, err := decodeXMLOr(d, it)
					if err != nil {
						return err
					}
					if base != nil {
						appendTok(*base, true)
					}
				case "marker":
					// Tokens inside marker within unify-ignore.
					if err := decodeUnifyMarker(d, true, appendTok); err != nil {
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
				// Java: last token in unify gets LastInUnification (+ optional uniNegation).
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
				featID, types, err := decodeUnifyFeature(d, t)
				if err != nil {
					return err
				}
				if featID != "" {
					features[featID] = types
				}
			case "token":
				xt, err := decodeXMLToken(d, t)
				if err != nil {
					return err
				}
				appendTok(xt, false)
			case "and":
				base, err := decodeXMLAnd(d, t)
				if err != nil {
					return err
				}
				if base != nil {
					appendTok(*base, false)
				}
			case "or":
				base, err := decodeXMLOr(d, t)
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
				if err := decodeUnifyMarker(d, false, appendTok); err != nil {
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

// decodeUnifyFeature ports <feature id="…"> optional <type id="…"/> children.
// Empty type list means all registered types for the feature (Java).
func decodeUnifyFeature(d *xml.Decoder, start xml.StartElement) (id string, types []string, err error) {
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
				// Drain type element body to its end tag.
				if err := drainElement(d, t.Name.Local); err != nil {
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

// drainElement consumes tokens until the matching end element (start already consumed).
func drainElement(d *xml.Decoder, name string) error {
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

// decodeUnifyMarker appends tokens inside <marker> within a unify block.
func decodeUnifyMarker(d *xml.Decoder, neutral bool, appendTok func(xmlToken, bool)) error {
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
				appendTok(xt, neutral)
			case "and":
				base, err := decodeXMLAnd(d, it)
				if err != nil {
					return err
				}
				if base != nil {
					appendTok(*base, neutral)
				}
			case "or":
				base, err := decodeXMLOr(d, it)
				if err != nil {
					return err
				}
				if base != nil {
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

func copyFeatureMap(in map[string][]string) map[string][]string {
	if in == nil {
		return map[string][]string{}
	}
	out := make(map[string][]string, len(in))
	for k, v := range in {
		out[k] = append([]string(nil), v...)
	}
	return out
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

// decodeXMLOr ports Java <or>: first token is the base PatternToken, later tokens
// become or-group alternatives (XMLRuleHandler.finalizeTokens + setOrGroupElement).
func decodeXMLOr(d *xml.Decoder, start xml.StartElement) (*xmlToken, error) {
	var orToks []xmlToken
	for {
		inner, err := d.Token()
		if err != nil {
			return nil, err
		}
		switch it := inner.(type) {
		case xml.EndElement:
			if it.Name.Local == start.Name.Local {
				if len(orToks) == 0 {
					return nil, nil
				}
				base := orToks[0]
				base.OrTokens = append([]xmlToken(nil), orToks[1:]...)
				return &base, nil
			}
		case xml.StartElement:
			if it.Name.Local == "token" {
				xt, err := decodeXMLToken(d, it)
				if err != nil {
					return nil, err
				}
				orToks = append(orToks, xt)
			} else if err := d.Skip(); err != nil {
				return nil, err
			}
		}
	}
}

type xmlToken struct {
	Regexp        string `xml:"regexp,attr"`
	CaseSensitive string `xml:"case_sensitive,attr"`
	Negate        string `xml:"negate,attr"`
	Inflected     string `xml:"inflected,attr"`
	Min           string `xml:"min,attr"`
	Max           string `xml:"max,attr"`
	Skip          string `xml:"skip,attr"`
	Postag        string `xml:"postag,attr"`
	PostagRegexp  string `xml:"postag_regexp,attr"`
	// SpaceBefore ports spacebefore="yes|no" (Java PatternToken.setWhitespaceBefore).
	SpaceBefore string `xml:"spacebefore,attr"`
	// Chunk / ChunkRe port Java PatternToken chunk / chunk_re.
	Chunk      string         `xml:"chunk,attr"`
	ChunkRe    string         `xml:"chunk_re,attr"`
	NegatePos  string         `xml:"negate_pos,attr"`
	Content    string         `xml:",chardata"`
	Exceptions []xmlException `xml:"exception"`
	// Match ports <match no="…" setpos="yes" …/> inside a pattern token.
	Match *xmlTokenMatch `xml:"match"`
	// AndTokens ports soft <and_token> = Java <and> group members.
	AndTokens []xmlToken `xml:"and_token"`
	// OrTokens ports Java <or> group members after the first token (decodeXMLOr).
	OrTokens []xmlToken `xml:"-"`
	// Uni* filled by decodeXMLUnify (not XML attributes).
	UniFeatures        map[string][]string
	UniNegated         bool
	LastInUnification  bool
	UnificationNeutral bool
	// InMarker ports Java PatternToken.setInsideMarker(inMarker) for tokens under <marker>.
	InMarker bool
}

// xmlTokenMatch ports pattern-token <match> (backward reference / setpos).
type xmlTokenMatch struct {
	No             string `xml:"no,attr"`
	Postag         string `xml:"postag,attr"`
	PostagRegexp   string `xml:"postag_regexp,attr"`
	PostagReplace  string `xml:"postag_replace,attr"`
	RegexpMatch    string `xml:"regexp_match,attr"`
	RegexpReplace  string `xml:"regexp_replace,attr"`
	CaseConversion string `xml:"case_conversion,attr"`
	SetPos         string `xml:"setpos,attr"`
	IncludeSkipped string `xml:"include_skipped,attr"`
	Content        string `xml:",chardata"`
}

type xmlException struct {
	Regexp        string `xml:"regexp,attr"`
	Negate        string `xml:"negate,attr"`
	NegatePos     string `xml:"negate_pos,attr"`
	CaseSensitive string `xml:"case_sensitive,attr"`
	Scope         string `xml:"scope,attr"` // previous|next|empty=current
	Postag        string `xml:"postag,attr"`
	PostagRegexp  string `xml:"postag_regexp,attr"`
	Content       string `xml:",chardata"`
}

func (l *PatternRuleLoader) parseRulesXML(data []byte, languageCode, filename string) ([]*AbstractPatternRule, error) {
	var root xmlRulesRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	cfg := NewUnifierConfiguration()
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
			var pt *PatternToken
			if eq.Token != nil {
				pt = tokenFromXML(*eq.Token)
			} else {
				pt = NewPatternToken("", false, false, false)
			}
			cfg.SetEquivalence(feat, typ, pt)
		}
	}
	l.LastUnifierConfig = cfg
	// phraseMap: id → alternatives (each alternative is a token sequence).
	phraseMap := buildPhraseMap(root.Phrases)
	var out []*AbstractPatternRule
	idPrefix := strings.TrimSpace(root.IdPrefix)
	add := func(xr xmlRule, defaultID, catID, catName string, catTags, groupTags []rules.Tag, catTones, groupTones []languagetool.ToneTag, catGoal, groupGoal, groupDefault string, catDefaultOff bool, catType, groupType, groupURL, sourceFile string, catPrio, groupPrio int, filePremium, catPremium, groupPremium string, groupMinPrev, groupDistTok int) error {
		id := xr.ID
		if id == "" {
			id = defaultID
		}
		// Java PatternRuleHandler: id = idPrefix + id when idprefix is set on <rules>.
		if id != "" && idPrefix != "" && !strings.HasPrefix(id, idPrefix) {
			id = idPrefix + id
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
		// Java RuleFilter: decide match acceptance (e.g. MultitokenSpellerFilter).
		// Unknown filter class → skip rule (fail-closed; never match without filter).
		var resolvedFilter RuleFilter
		var filterArgs string
		if xr.Filter != nil && strings.TrimSpace(xr.Filter.Class) != "" {
			class := strings.TrimSpace(xr.Filter.Class)
			f, ok := GlobalRuleFilterCreator.TryGetFilter(class)
			if !ok {
				return nil
			}
			resolvedFilter = f
			filterArgs = strings.TrimSpace(xr.Filter.Args)
		}
		// Java: pattern-level case_sensitive inherits to tokens/exceptions
		// when the child does not set its own case_sensitive attribute.
		patternCS := strings.EqualFold(xr.Pattern.CaseSensitive, "yes")
		// Expand phraseref (Java preparePhrase + createRules) then <or> groups.
		phraseExpanded := expandPatternSteps(xr.Pattern, phraseMap, patternCS)
		if len(phraseExpanded) == 0 {
			return nil
		}
		antis := antiPatternsFromXML(id, languageCode, xr.AntiPatterns, cfg, phraseMap)
		rawMsg := strings.TrimSpace(xr.Message.Inner)
		msg, sugMatches := ProcessRuleMessage(rawMsg)
		short := strings.TrimSpace(xr.Short)
		// Java: rulegroup default off/temp_off overrides per-rule default=…
		defaultOff, defaultTempOff := resolveRuleDefaultOff(xr.Default, groupDefault)
		// tags / tone_tags: rule + group + category (Java addTags/addToneTags order).
		tags := mergeRuleTags(parseRuleTagsAttr(xr.Tags), groupTags, catTags)
		tones := mergeToneTags(parseToneTagsAttr(xr.ToneTags), groupTones, catTones)
		// is_goal_specific: rule overrides group overrides category (Java PatternRuleHandler).
		goalSpecific := resolveGoalSpecific(xr.GoalSpecific, groupGoal, catGoal)
		for _, tokens := range phraseExpanded {
			// Java PatternRuleHandler.createRules: expand <or> into multiple rules.
			for _, expToks := range expandOrGroups(tokens) {
				if len(expToks) == 0 {
					continue
				}
				r := NewAbstractPatternRule(id, name, languageCode, expToks, false)
				r.Message = msg
				r.ShortMessage = short
				r.UnifierConfig = cfg
				r.TestUnification = anyTokenUnified(expToks)
				r.InterpretPreDisambig = strings.EqualFold(xr.Pattern.RawPos, "yes")
				r.SuggestionMatches = append([]*Match(nil), sugMatches...)
				r.AntiPatterns = append([]*PatternRule(nil), antis...)
				if resolvedFilter != nil {
					r.Filter = resolvedFilter
					r.FilterArgs = filterArgs
				}
				r.CategoryID = catID
				r.CategoryName = catName
				r.CategoryDefaultOff = catDefaultOff
				r.CategoryType = catType
				// Java: rule type → rulegroup type → category type
				r.IssueType = resolveIssueType(xr.Type, groupType, catType)
				// Java: rule url else rulegroup url
				r.URL = resolveRuleURL(xr.URL, groupURL)
				r.SourceFile = sourceFile
				// Java finalize: cat prio then group prio then rule prio (non-zero overwrites).
				r.Priority = resolvePriority(catPrio, groupPrio, parsePrioAttr(xr.Prio))
				// Java prepareRule setPremium(isPremiumRule): rule > group > category > file.
				r.Premium = resolvePremium(xr.Premium, groupPremium, catPremium, filePremium)
				// Java: rule attr or inherit ruleGroupMinPrevMatches / ruleGroupDistanceTokens.
				r.MinPrevMatches = resolveIntAttr(xr.MinPrevMatches, groupMinPrev)
				r.DistanceTokens = resolveIntAttr(xr.DistanceTokens, groupDistTok)
				if defaultOff {
					r.DefaultOff = true
				}
				if defaultTempOff {
					r.DefaultTempOff = true
				}
				r.ToneTags = tones
				r.GoalSpecific = goalSpecific
				// Store tags on abstract for RegisterGrammarXML → PatternRule.SetTags
				if len(tags) > 0 {
					r.Tags = append([]rules.Tag(nil), tags...)
				}
				out = append(out, r)
			}
		}
		return nil
	}
	srcFile := strings.TrimSpace(filename)
	filePremium := strings.TrimSpace(root.Premium)
	for _, cat := range root.Categories {
		catTags := parseRuleTagsAttr(cat.Tags)
		catTones := parseToneTagsAttr(cat.ToneTags)
		// Java: onByDefault = !OFF.equals(attrs.getValue(DEFAULT))
		catDefaultOff := strings.EqualFold(strings.TrimSpace(cat.Default), XMLOff)
		catType := strings.TrimSpace(cat.Type)
		catPrio := parsePrioAttr(cat.Prio)
		catPremium := strings.TrimSpace(cat.Premium)
		for _, xr := range cat.Rules {
			if err := add(xr, "", cat.ID, cat.Name, catTags, nil, catTones, nil, cat.GoalSpecific, "", "", catDefaultOff, catType, "", "", srcFile, catPrio, 0, filePremium, catPremium, "", 0, 0); err != nil {
				return nil, err
			}
		}
		for _, g := range cat.RuleGroups {
			// Java: rulegroup antipatterns apply to every child rule (prepareRule).
			groupID := g.ID
			if groupID != "" && idPrefix != "" && !strings.HasPrefix(groupID, idPrefix) {
				groupID = idPrefix + groupID
			}
			groupAntis := antiPatternsFromXML(groupID, languageCode, g.AntiPatterns, cfg, phraseMap)
			groupTags := parseRuleTagsAttr(g.Tags)
			groupTones := parseToneTagsAttr(g.ToneTags)
			groupType := strings.TrimSpace(g.Type)
			groupURL := strings.TrimSpace(g.URL)
			groupPrio := parsePrioAttr(g.Prio)
			groupPremium := strings.TrimSpace(g.Premium)
			groupMinPrev := parsePrioAttr(g.MinPrevMatches) // same int parse as prio
			groupDistTok := parsePrioAttr(g.DistanceTokens)
			for i, xr := range g.Rules {
				id := xr.ID
				if id == "" {
					id = g.ID
				}
				start := len(out)
				if err := add(xr, id, cat.ID, cat.Name, catTags, groupTags, catTones, groupTones, cat.GoalSpecific, g.GoalSpecific, g.Default, catDefaultOff, catType, groupType, groupURL, srcFile, catPrio, groupPrio, filePremium, catPremium, groupPremium, groupMinPrev, groupDistTok); err != nil {
					return nil, err
				}
				// sub id 1-based per XML rule (shared by OR expansions of that rule)
				sub := fmt.Sprintf("%d", i+1)
				for j := start; j < len(out); j++ {
					last := out[j]
					last.SubID = sub
					if last.ID == "" {
						last.ID = groupID
					}
					// Java setAntiPatterns: rulegroup first, then rule-level (append).
					if len(groupAntis) > 0 {
						last.AntiPatterns = append(append([]*PatternRule(nil), groupAntis...), last.AntiPatterns...)
					}
				}
			}
		}
	}
	for _, xr := range root.Rules {
		if err := add(xr, "", "", "", nil, nil, nil, nil, "", "", "", false, "", "", "", srcFile, 0, 0, filePremium, "", "", 0, 0); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// resolveIntAttr ports rule attr with inheritance: if ruleAttr set use it, else inherit.
func resolveIntAttr(ruleAttr string, inherit int) int {
	s := strings.TrimSpace(ruleAttr)
	if s == "" {
		return inherit
	}
	return parsePrioAttr(s)
}

// resolvePremium ports PatternRuleHandler isPremiumRule:
// rule premium → rulegroup → category → file; yes/true → true; no/false → false; unset → false.
func resolvePremium(rulePrem, groupPrem, catPrem, filePrem string) bool {
	for _, a := range []string{rulePrem, groupPrem, catPrem, filePrem} {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		if strings.EqualFold(a, XMLYes) || strings.EqualFold(a, XMLTrue) {
			return true
		}
		if strings.EqualFold(a, XMLNo) || strings.EqualFold(a, XMLFalse) {
			return false
		}
	}
	return false
}

// parsePrioAttr ports Integer.parseInt on XML prio= (invalid/empty → 0).
func parsePrioAttr(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0
	}
	return n
}

// resolvePriority ports PatternRuleHandler finalize priority:
// start 0; if cat≠0 use cat; if group≠0 use group; if rule≠0 use rule.
func resolvePriority(catPrio, groupPrio, rulePrio int) int {
	prio := 0
	if catPrio != 0 {
		prio = catPrio
	}
	if groupPrio != 0 {
		prio = groupPrio
	}
	if rulePrio != 0 {
		prio = rulePrio
	}
	return prio
}

// resolveIssueType ports PatternRuleHandler type inheritance:
// rule type → rulegroup type → category type (Java setLocQualityIssueType).
func resolveIssueType(ruleType, groupType, catType string) string {
	for _, t := range []string{ruleType, groupType, catType} {
		t = strings.TrimSpace(t)
		if t != "" {
			return strings.ToLower(t)
		}
	}
	return ""
}

// resolveRuleURL ports rule <url> else rulegroup <url>.
func resolveRuleURL(ruleURL, groupURL string) string {
	if u := strings.TrimSpace(ruleURL); u != "" {
		return u
	}
	return strings.TrimSpace(groupURL)
}

// resolveRuleDefaultOff ports PatternRuleHandler default inheritance:
// when rulegroup is default=off or temp_off, all children inherit;
// otherwise the rule's own default= attribute is used.
// temp_off implies defaultOff (Java setDefaultTempOff).
func resolveRuleDefaultOff(ruleDefault, groupDefault string) (defaultOff, defaultTempOff bool) {
	gdef := strings.ToLower(strings.TrimSpace(groupDefault))
	switch gdef {
	case XMLOff:
		return true, false
	case XMLTempOff:
		return true, true
	}
	rdef := strings.ToLower(strings.TrimSpace(ruleDefault))
	switch rdef {
	case XMLTempOff:
		return true, true
	case XMLOff:
		return true, false
	}
	return false, false
}

// antiPatternsFromXML builds PatternRule antipatterns from XML <antipattern> blocks.
// Java: DisambiguationPatternRule with IMMUNIZE; Go Match uses overlap suppress equivalent.
// OR groups and phraserefs are expanded (same as createRules).
func antiPatternsFromXML(ruleID, languageCode string, aps []xmlPattern, cfg *UnifierConfiguration, phraseMap map[string][][]*PatternToken) []*PatternRule {
	if len(aps) == 0 {
		return nil
	}
	var out []*PatternRule
	for i, ap := range aps {
		patternCS := strings.EqualFold(ap.CaseSensitive, "yes")
		apID := fmt.Sprintf("%s_anti_%d", ruleID, i)
		if ruleID == "" {
			apID = fmt.Sprintf("anti_%d", i)
		}
		n := 0
		for _, apToks := range expandPatternSteps(ap, phraseMap, patternCS) {
			if len(apToks) == 0 {
				continue
			}
			// Java without <marker>: mark all tokens so immunize spans the full antipattern.
			for _, t := range apToks {
				if t != nil {
					t.InsideMarker = true
				}
			}
			for ei, exp := range expandOrGroups(apToks) {
				id := apID
				if n > 0 || ei > 0 {
					id = fmt.Sprintf("%s_x%d", apID, n)
				}
				n++
				pr := NewPatternRule(id, languageCode, exp, "antipattern", "", "")
				pr.UnifierConfig = cfg
				out = append(out, pr)
			}
		}
	}
	return out
}

// buildPhraseMap ports PatternRuleHandler phrase definitions (finalizePhrase / preparePhrase).
// Map: phrase id → list of alternatives (each a token sequence).
func buildPhraseMap(block *xmlPhrasesBlock) map[string][][]*PatternToken {
	m := map[string][][]*PatternToken{}
	if block == nil {
		return m
	}
	for _, def := range block.Phrases {
		id := strings.TrimSpace(def.ID)
		if id == "" {
			continue
		}
		// Expand steps against already-defined phrases (includephrases order).
		alts := expandSteps(def.Steps, m, false)
		if len(alts) == 0 {
			continue
		}
		// Deep-copy so later mutations of PatternToken don't alias phrase map entries.
		copied := make([][]*PatternToken, 0, len(alts))
		for _, alt := range alts {
			cp := make([]*PatternToken, len(alt))
			for i, t := range alt {
				cp[i] = clonePatternTokenNoOr(t)
			}
			copied = append(copied, cp)
		}
		m[id] = copied
	}
	return m
}

// expandPatternSteps converts a pattern to one or more token sequences (phraseref expansion).
func expandPatternSteps(p xmlPattern, phraseMap map[string][][]*PatternToken, patternCS bool) [][]*PatternToken {
	hasMarker := p.HasMarker || patternHasMarker(p)
	steps := p.Steps
	if len(steps) == 0 && len(p.Tokens) > 0 {
		// Fallback: tokens-only pattern (no custom steps recorded).
		var seq []*PatternToken
		for _, xt := range p.Tokens {
			xt = applyPatternCaseSensitive(xt, patternCS)
			seq = append(seq, tokenFromXMLWithMarker(xt, hasMarker))
		}
		if len(seq) == 0 {
			return nil
		}
		return [][]*PatternToken{seq}
	}
	return expandSteps(steps, phraseMap, patternCS)
}

// patternHasMarker reports whether any step/token was inside <marker>.
func patternHasMarker(p xmlPattern) bool {
	if p.HasMarker {
		return true
	}
	for _, st := range p.Steps {
		if st.InMarker || (st.Token != nil && st.Token.InMarker) {
			return true
		}
	}
	for _, xt := range p.Tokens {
		if xt.InMarker {
			return true
		}
	}
	return false
}

// expandSteps ports XMLRuleHandler.preparePhrase / finalizePhrase / createRules:
//
//   - phrasePatternTokens holds alternatives
//   - patternTokens is the current prefix/suffix buffer
//   - phraseref with empty buffer: each phrase alt becomes a new alternative (union)
//   - phraseref with non-empty buffer: each phrase alt is buffer+phrase (fork)
//   - token after phraseref clears buffer (lastPhrase), then appends
//   - at end, non-empty buffer is appended to every alternative
func expandSteps(steps []xmlPatternStep, phraseMap map[string][][]*PatternToken, patternCS bool) [][]*PatternToken {
	var alternatives [][]*PatternToken // Java phrasePatternTokens
	var buffer []*PatternToken         // Java patternTokens
	lastPhrase := false
	hasMarker := false
	for _, st := range steps {
		if st.InMarker || (st.Token != nil && st.Token.InMarker) {
			hasMarker = true
			break
		}
	}

	for _, st := range steps {
		if st.PhraseRef != "" {
			alts, ok := phraseMap[st.PhraseRef]
			if !ok || len(alts) == 0 {
				// Unknown phrase: fail-closed (no invent).
				return nil
			}
			for _, alt := range alts {
				copyAlt := make([]*PatternToken, 0, len(buffer)+len(alt))
				if len(buffer) == 0 {
					for _, t := range alt {
						// Java preparePhrase: phrase tokens setInsideMarker(inMarker).
						ct := clonePatternTokenNoOr(t)
						if hasMarker {
							ct.SetInsideMarker(st.InMarker)
						}
						copyAlt = append(copyAlt, ct)
					}
					alternatives = append(alternatives, copyAlt)
				} else {
					// prefix buffer + phrase alt
					for _, t := range buffer {
						copyAlt = append(copyAlt, clonePatternToken(t))
					}
					for _, t := range alt {
						ct := clonePatternTokenNoOr(t)
						if hasMarker {
							ct.SetInsideMarker(st.InMarker)
						}
						copyAlt = append(copyAlt, ct)
					}
					alternatives = append(alternatives, copyAlt)
				}
			}
			lastPhrase = true
			continue
		}
		if st.Token == nil {
			continue
		}
		// Java setToken: if lastPhrase, patternTokens.clear()
		if lastPhrase {
			buffer = nil
			lastPhrase = false
		}
		xt := *st.Token
		if st.InMarker {
			xt.InMarker = true
		}
		xt = applyPatternCaseSensitive(xt, patternCS)
		pt := tokenFromXMLWithMarker(xt, hasMarker)
		buffer = append(buffer, clonePatternToken(pt))
	}

	// Java rule end: if !patternTokens.isEmpty() { for ph : phrasePatternTokens { ph.addAll(patternTokens) } }
	if len(alternatives) == 0 {
		if len(buffer) == 0 {
			return nil
		}
		return [][]*PatternToken{buffer}
	}
	if len(buffer) > 0 {
		for i := range alternatives {
			for _, t := range buffer {
				alternatives[i] = append(alternatives[i], clonePatternToken(t))
			}
		}
	}
	return alternatives
}

func applyPatternCaseSensitive(xt xmlToken, patternCS bool) xmlToken {
	if !patternCS {
		return xt
	}
	if xt.CaseSensitive == "" {
		xt.CaseSensitive = "yes"
	}
	for i := range xt.Exceptions {
		if xt.Exceptions[i].CaseSensitive == "" {
			xt.Exceptions[i].CaseSensitive = "yes"
		}
	}
	return xt
}

// matchFromTokenMatchXML builds a Match from pattern-token <match> attributes.
func matchFromTokenMatchXML(xm *xmlTokenMatch) *Match {
	if xm == nil {
		return nil
	}
	caseConv := CaseNone
	if v := strings.TrimSpace(xm.CaseConversion); v != "" {
		caseConv = CaseConversion(strings.ToUpper(v))
	}
	include := IncludeNone
	if v := strings.TrimSpace(xm.IncludeSkipped); v != "" {
		include = IncludeRange(strings.ToUpper(v))
	}
	m := NewMatch(
		xm.Postag,
		xm.PostagReplace,
		strings.EqualFold(xm.PostagRegexp, "yes"),
		xm.RegexpMatch,
		xm.RegexpReplace,
		caseConv,
		strings.EqualFold(xm.SetPos, "yes"),
		false,
		include,
	)
	// Java: TokenRef is the raw no= value (offset from firstMatchToken, not 1-based message index).
	if no := strings.TrimSpace(xm.No); no != "" {
		var n int
		if _, err := fmt.Sscanf(no, "%d", &n); err == nil {
			m.SetTokenRef(n)
		}
	}
	if body := strings.TrimSpace(xm.Content); body != "" {
		m.SetLemmaString(body)
	}
	return m
}

func anyTokenUnified(tokens []*PatternToken) bool {
	for _, t := range tokens {
		if t != nil && t.IsUnified() {
			return true
		}
	}
	return false
}

// tokenFromXMLWithMarker ports Java setInsideMarker(inMarker):
//   - no <marker> in pattern → InsideMarker true (default; match uses full span)
//   - has <marker> → only tokens with InMarker are InsideMarker true
func tokenFromXMLWithMarker(xt xmlToken, patternHasMarker bool) *PatternToken {
	pt := tokenFromXML(xt)
	if patternHasMarker {
		pt.SetInsideMarker(xt.InMarker)
	}
	return pt
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
			Negate: strings.EqualFold(xt.NegatePos, "yes"),
		})
	} else if strings.EqualFold(xt.PostagRegexp, "yes") && xt.Match != nil {
		// POS filled by setpos match at compile time; mark as regexp POS shell.
		pt.SetPosToken(PosToken{Regexp: true})
	}
	if sb := strings.TrimSpace(xt.SpaceBefore); sb != "" {
		pt.SetWhitespaceBefore(strings.EqualFold(sb, "yes"))
	}
	if xt.Match != nil {
		pt.SetMatch(matchFromTokenMatchXML(xt.Match))
	}
	if ch := strings.TrimSpace(xt.ChunkRe); ch != "" {
		pt.SetChunkTag(ch, true)
	} else if ch := strings.TrimSpace(xt.Chunk); ch != "" {
		pt.SetChunkTag(ch, false)
	}
	// Current exception (surface and/or postag) + scope previous/next.
	// Java: isExceptionMatchedCompletely after any reading matches the token;
	// exception negate / negate_pos use PatternToken.isMatched XOR semantics.
	for _, ex := range xt.Exceptions {
		exc := strings.TrimSpace(ex.Content)
		posTag := strings.TrimSpace(ex.Postag)
		if exc == "" && posTag == "" {
			continue
		}
		scope := strings.ToLower(strings.TrimSpace(ex.Scope))
		re := strings.EqualFold(ex.Regexp, "yes")
		cs := strings.EqualFold(ex.CaseSensitive, "yes")
		posRE := strings.EqualFold(ex.PostagRegexp, "yes")
		neg := strings.EqualFold(ex.Negate, "yes")
		posNeg := strings.EqualFold(ex.NegatePos, "yes")
		switch scope {
		case "previous":
			// previous/next: surface only for now; negation on previous not yet multi-exception
			if exc != "" && pt.PreviousException == "" && !neg && !posNeg {
				pt.SetPreviousException(exc, re, cs)
			}
		case "next":
			if exc != "" && pt.NextException == "" && !neg && !posNeg {
				pt.SetNextException(exc, re, cs)
			}
		default:
			if !pt.HasCurrentException() {
				pt.SetStringPosExceptionFullNeg(exc, re, cs, neg, posTag, posRE, posNeg)
			}
		}
	}
	// Java <and> group members (soft <and_token>).
	for _, at := range xt.AndTokens {
		pt.AddAndGroupElement(tokenFromXML(at))
	}
	// Java <or> group members (alternatives after the first token).
	for _, ot := range xt.OrTokens {
		pt.AddOrGroupElement(tokenFromXML(ot))
	}
	return pt
}

// expandOrGroups ports PatternRuleHandler.createRules OR expansion:
// for each token with OrGroup, emit one rule variant per alternative (including the base).
func expandOrGroups(tokens []*PatternToken) [][]*PatternToken {
	if len(tokens) == 0 {
		return nil
	}
	var out [][]*PatternToken
	var rec func(i int, prefix []*PatternToken)
	rec = func(i int, prefix []*PatternToken) {
		if i >= len(tokens) {
			cp := make([]*PatternToken, len(prefix))
			copy(cp, prefix)
			out = append(out, cp)
			return
		}
		t := tokens[i]
		if t != nil && t.HasOrGroup() {
			// Java: for each or-group member, then also the base token itself.
			for _, alt := range t.OrGroup {
				rec(i+1, append(prefix, clonePatternTokenNoOr(alt)))
			}
			rec(i+1, append(prefix, clonePatternTokenNoOr(t)))
			return
		}
		rec(i+1, append(prefix, t))
	}
	rec(0, nil)
	return out
}

// clonePatternToken shallow-copies a token (preserves OrGroup for expandOrGroups).
func clonePatternToken(p *PatternToken) *PatternToken {
	if p == nil {
		return nil
	}
	cp := *p
	if p.UniFeatures != nil {
		cp.UniFeatures = copyFeatureMap(p.UniFeatures)
	}
	if len(p.OrGroup) > 0 {
		cp.OrGroup = make([]*PatternToken, len(p.OrGroup))
		for i, o := range p.OrGroup {
			cp.OrGroup[i] = clonePatternToken(o)
		}
	}
	// TokenMatch is read-only config; share pointer.
	// AndGroup / exceptions are read-only after load; share slices.
	return &cp
}

// clonePatternTokenNoOr shallow-copies a token and clears OrGroup (post-expansion).
func clonePatternTokenNoOr(p *PatternToken) *PatternToken {
	cp := clonePatternToken(p)
	if cp != nil {
		cp.OrGroup = nil
	}
	return cp
}
