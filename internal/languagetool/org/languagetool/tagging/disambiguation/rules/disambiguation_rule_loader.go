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
	// Official LT disambiguation.xml uses custom DTD entities.
	data = patterns.ExpandLTXMLEntities(data)
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
	Content     string `xml:",chardata"`
}

type disambigElem struct {
	Action string      `xml:"action,attr"`
	Postag string      `xml:"postag,attr"`
	Wds    []disambigWd `xml:"wd"`
}

// disambigWd ports <wd pos="…" lemma="…"/> under <disambig action="add">.
type disambigWd struct {
	Pos    string `xml:"pos,attr"`
	Lemma  string `xml:"lemma,attr"`
	Content string `xml:",chardata"`
}

func (l *DisambiguationRuleLoader) parse(data []byte, languageCode, xmlPath string) ([]*DisambiguationPatternRule, error) {
	var root disambigRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse disambiguation %s: %w", xmlPath, err)
	}
	var out []*DisambiguationPatternRule
	for _, xr := range root.Rules {
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
		action := ActionReplace
		if xr.Disambig.Action != "" {
			action = DisambiguatorAction(strings.ToUpper(xr.Disambig.Action))
		}
		// default Java: REPLACE when only postag set
		rule := NewDisambiguationPatternRule(xr.ID, xr.Name, languageCode, tokens, xr.Disambig.Postag, nil, action)
		// Java ADD with <wd pos="PCT"/> etc. (UNKNOWN_PCT, COMMA_POSTAG)
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
				// empty surface: filled from matched token at apply time
				readings = append(readings, languagetool.NewAnalyzedToken(surf, posPtr, lemmaPtr))
			}
			rule.SetNewInterpretations(readings)
		}
		out = append(out, rule)
	}
	return out, nil
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
	return pt
}
