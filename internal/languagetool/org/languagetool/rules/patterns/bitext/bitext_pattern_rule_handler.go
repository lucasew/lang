package bitext

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// BitextPatternRuleHandler ports
// org.languagetool.rules.patterns.bitext.BitextPatternRuleHandler for structured XML:
//
//	<rules targetLang="de">
//	  <rule id="..." name="...">
//	    <source lang="en"><pattern>...</pattern></source>
//	    <target><pattern>...</pattern></target>
//	    <message>...</message>
//	  </rule>
//	</rules>
//
// Also accepts the simplified bitextrules format via BitextPatternRuleLoader.
type BitextPatternRuleHandler struct {
	TargetLang string
	Rules      []*BitextPatternRule
}

func NewBitextPatternRuleHandler() *BitextPatternRuleHandler {
	return &BitextPatternRuleHandler{}
}

func (h *BitextPatternRuleHandler) GetBitextRules() []*BitextPatternRule {
	return h.Rules
}

// Parse reads bitext rules XML.
func (h *BitextPatternRuleHandler) Parse(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	// try full handler format first
	var root bphRoot
	if err := xml.Unmarshal(data, &root); err == nil && (len(root.Rules) > 0 || root.TargetLang != "") {
		h.TargetLang = root.TargetLang
		h.Rules = nil
		for _, xr := range root.Rules {
			br, err := h.build(xr)
			if err != nil {
				return err
			}
			if br != nil {
				h.Rules = append(h.Rules, br)
			}
		}
		if len(h.Rules) > 0 {
			return nil
		}
	}
	// fallback simplified loader
	rules, err := NewBitextPatternRuleLoader().GetRules(strings.NewReader(string(data)), "bitext.xml")
	if err != nil {
		return fmt.Errorf("bitext handler: %w", err)
	}
	h.Rules = rules
	return nil
}

func (h *BitextPatternRuleHandler) build(xr bphRule) (*BitextPatternRule, error) {
	if xr.Source.Pattern.Lang == "" && xr.Source.Lang != "" {
		xr.Source.Pattern.Lang = xr.Source.Lang
	}
	srcLang := xr.Source.Pattern.Lang
	if srcLang == "" {
		srcLang = xr.Source.Lang
	}
	trgLang := h.TargetLang
	if trgLang == "" {
		trgLang = "en"
	}
	src := tokensToPatternRule(xr.ID, xr.Name, srcLang, xr.Source.Pattern.Tokens, xr.Message)
	trg := tokensToPatternRule(xr.ID, xr.Name, trgLang, xr.Target.Pattern.Tokens, xr.Message)
	if src == nil || trg == nil {
		return nil, nil
	}
	br := NewBitextPatternRule(src, trg)
	br.SetSourceLanguage(srcLang)
	return br, nil
}

func tokensToPatternRule(id, name, lang string, tokens []bphToken, message string) *patterns.PatternRule {
	if len(tokens) == 0 {
		return nil
	}
	var pts []*patterns.PatternToken
	for _, t := range tokens {
		pts = append(pts, patterns.NewPatternToken(strings.TrimSpace(t.Content), false, strings.EqualFold(t.Regexp, "yes"), false))
	}
	if message == "" {
		message = name
	}
	return patterns.NewPatternRule(id, lang, pts, name, message, "")
}

type bphRoot struct {
	XMLName    xml.Name  `xml:"rules"`
	TargetLang string    `xml:"targetLang,attr"`
	Rules      []bphRule `xml:"rule"`
}

type bphRule struct {
	ID      string    `xml:"id,attr"`
	Name    string    `xml:"name,attr"`
	Source  bphSource `xml:"source"`
	Target  bphTarget `xml:"target"`
	Message string    `xml:"message"`
}

type bphSource struct {
	Lang    string     `xml:"lang,attr"`
	Pattern bphPattern `xml:"pattern"`
}

type bphTarget struct {
	Pattern bphPattern `xml:"pattern"`
}

type bphPattern struct {
	Lang   string     `xml:"lang,attr"`
	Tokens []bphToken `xml:"token"`
}

type bphToken struct {
	Regexp  string `xml:"regexp,attr"`
	Content string `xml:",chardata"`
}
