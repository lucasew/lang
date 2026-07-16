// Package disambig implements a subset of LanguageTool XML rule disambiguation.
package disambig

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/jregex"
	"github.com/lucasew/lang/internal/pattern"
	"github.com/lucasew/lang/internal/pipeline"
)

// Action kinds for disambiguation.
type Action int

const (
	ActionFilter Action = iota // default: keep only matching readings (disambig postag=)
	ActionRemove               // remove readings matching wd
	ActionAdd                  // add a reading
	ActionReplace              // replace with postag on all
)

// Rule is one disambiguation rule.
type Rule struct {
	ID      string
	Tokens  []pattern.PatToken
	Action  Action
	// For remove/add: target POS and optional lemma
	TargetPOS   string
	TargetLemma string
	// For filter/replace: assign this POS
	AssignPOS string
	// Marker range in pattern tokens
	MarkerFrom int // inclusive index in Tokens, -1 = all matched
	MarkerTo   int // exclusive
	Incomplete bool
}

// Engine runs disambiguation rules in order.
type Engine struct {
	Rules []*Rule
}

// LoadFile loads disambiguation.xml (with DTD entities).
func LoadFile(path string) (*Engine, error) {
	// Reuse pattern entity expander via reading as grammar-like file
	// Use simple expand by importing pattern's approach - duplicate minimal expand here
	r, err := openExpanded(path)
	if err != nil {
		return nil, err
	}
	rules, err := parse(r)
	if err != nil {
		return nil, err
	}
	return &Engine{Rules: rules}, nil
}

func openExpanded(path string) (io.Reader, error) {
	// Use pattern package's loader path by reading through a temp approach:
	// We call pattern.LoadFile's entity expand - not exported.
	// Duplicate light expansion using same files in pattern - export function.
	return pattern.OpenExpandedXML(path)
}

// Apply runs all rules on non-whitespace tokens (with SENT_START at 0).
func (e *Engine) Apply(tokens []pipeline.Token) []pipeline.Token {
	if e == nil {
		return tokens
	}
	for _, r := range e.Rules {
		if r.Incomplete || len(r.Tokens) == 0 {
			continue
		}
		tokens = applyRule(r, tokens)
	}
	return tokens
}

func applyRule(r *Rule, tokens []pipeline.Token) []pipeline.Token {
	// Build fake MatchContext-like walk using pattern matching internals is hard (unexported).
	// Reimplement light match for disambig patterns on tokens.
	n := len(tokens)
	for start := 0; start < n; start++ {
		ok, end := matchPat(r.Tokens, tokens, start)
		if !ok {
			continue
		}
		from, to := start, end
		if r.MarkerFrom >= 0 {
			from = start + r.MarkerFrom
			to = start + r.MarkerTo
			if to <= from {
				to = from + 1
			}
		}
		if from < 0 {
			from = start
		}
		if to > end {
			to = end
		}
		for i := from; i < to && i < len(tokens); i++ {
			tokens[i] = applyAction(r, tokens[i])
		}
	}
	return tokens
}

func applyAction(r *Rule, tok pipeline.Token) pipeline.Token {
	switch r.Action {
	case ActionRemove:
		var keep []pipeline.Reading
		for _, rd := range tok.Readings {
			if r.TargetPOS != "" && rd.POS == r.TargetPOS {
				if r.TargetLemma == "" || rd.Lemma == r.TargetLemma {
					continue // remove
				}
			}
			if r.TargetLemma != "" && r.TargetPOS == "" && rd.Lemma == r.TargetLemma {
				continue
			}
			keep = append(keep, rd)
		}
		tok.Readings = keep
	case ActionAdd:
		tok.Readings = append(tok.Readings, pipeline.Reading{Lemma: firstLemma(tok), POS: r.TargetPOS})
	case ActionFilter, ActionReplace:
		if r.AssignPOS != "" {
			lemma := firstLemma(tok)
			tok.Readings = []pipeline.Reading{{Lemma: lemma, POS: r.AssignPOS}}
		}
	}
	return tok
}

func firstLemma(tok pipeline.Token) string {
	if len(tok.Readings) > 0 && tok.Readings[0].Lemma != "" {
		return tok.Readings[0].Lemma
	}
	return tok.Text
}

func matchPat(pattern []pattern.PatToken, tokens []pipeline.Token, start int) (bool, int) {
	ti := start
	for _, pt := range pattern {
		min, max := pt.Min, pt.Max
		if max < min {
			max = min
		}
		matched := 0
		for matched < max {
			if ti >= len(tokens) {
				break
			}
			if matchTok(pt, tokens[ti]) {
				matched++
				ti++
			} else {
				break
			}
		}
		if matched < min {
			return false, start
		}
	}
	return true, ti
}

func matchTok(pt pattern.PatToken, tok pipeline.Token) bool {
	if pt.Chunk != "" {
		return false
	}
	// text
	if pt.Value != "" || pt.Re != nil {
		ok := false
		if pt.Inflected {
			for _, l := range tok.Lemmas() {
				if matchStr(pt, l) {
					ok = true
					break
				}
			}
		} else {
			ok = matchStr(pt, tok.Text)
		}
		if pt.Negate {
			ok = !ok
		}
		if !ok {
			return false
		}
	}
	if pt.Postag != "" {
		hit := false
		if pt.Postag == "UNKNOWN" {
			hit = len(tok.Readings) == 0
		} else if pt.PostagRegexp {
			re := pt.Re // may be string re for postag - use compiled Postag field
			// compile on the fly
			re2, err := regexp.Compile("^(?:" + pt.Postag + ")$")
			if err != nil {
				return false
			}
			for _, r := range tok.Readings {
				if re2.MatchString(r.POS) {
					hit = true
					break
				}
			}
			_ = re
		} else {
			for _, r := range tok.Readings {
				if r.POS == pt.Postag {
					hit = true
					break
				}
			}
		}
		if pt.NegatePos {
			hit = !hit
		}
		if !hit {
			return false
		}
	}
	return true
}

func matchStr(pt pattern.PatToken, text string) bool {
	if pt.Regexp && pt.Re != nil {
		return pt.Re.MatchString(text)
	}
	if pt.CaseSensitive {
		return text == pt.Value
	}
	return strings.EqualFold(text, pt.Value)
}

func parse(r io.Reader) ([]*Rule, error) {
	dec := xml.NewDecoder(r)
	var rules []*Rule
	var cur *Rule
	var inPattern, inMarker, inDisambig, inWd bool
	var markerStart int
	var wdPOS, wdLemma string
	var tokDepth int

	flush := func() {
		if cur != nil && cur.ID != "" {
			rules = append(rules, cur)
		}
		cur = nil
	}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return rules, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "rule":
				flush()
				cur = &Rule{
					ID:         attr(t, "id"),
					MarkerFrom: -1,
					MarkerTo:   -1,
				}
				if cur.ID == "" {
					cur.ID = attr(t, "name")
				}
			case "pattern":
				inPattern = true
			case "marker":
				inMarker = true
				if cur != nil {
					markerStart = len(cur.Tokens)
					cur.MarkerFrom = markerStart
				}
			case "token":
				if cur == nil || !inPattern {
					continue
				}
				pt, err := parseToken(dec, t)
				if err != nil {
					return nil, err
				}
				if pt.NeedsPOS() {
					// ok for disambig
				}
				if pt.Chunk != "" {
					cur.Incomplete = true
				}
				cur.Tokens = append(cur.Tokens, pt)
				tokDepth++
			case "and", "or", "unify":
				_ = skip(dec, t)
				if cur != nil {
					cur.Incomplete = true
				}
			case "disambig":
				inDisambig = true
				if cur == nil {
					continue
				}
				action := attr(t, "action")
				switch action {
				case "remove":
					cur.Action = ActionRemove
				case "add":
					cur.Action = ActionAdd
				case "replace", "filter", "":
					if action == "replace" {
						cur.Action = ActionReplace
					} else {
						cur.Action = ActionFilter
					}
				default:
					cur.Incomplete = true
				}
				if p := attr(t, "postag"); p != "" {
					cur.AssignPOS = p
					cur.Action = ActionReplace
				}
			case "wd":
				inWd = true
				wdPOS = attr(t, "pos")
				wdLemma = attr(t, "lemma")
				if cur != nil {
					if wdPOS != "" {
						cur.TargetPOS = wdPOS
					}
					if wdLemma != "" {
						cur.TargetLemma = wdLemma
					}
				}
			case "unification", "equivalence", "rulegroup":
				// skip complex blocks partially - rulegroup contains rules handled as nested?
				if t.Name.Local == "rulegroup" {
					// continue parsing children as rules via same loop — rule start will fire
				} else {
					_ = skip(dec, t)
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "pattern":
				inPattern = false
			case "marker":
				inMarker = false
				if cur != nil {
					cur.MarkerTo = len(cur.Tokens)
				}
			case "disambig":
				inDisambig = false
			case "wd":
				inWd = false
			case "rule":
				if cur != nil && inMarker {
					// unclosed marker
				}
				_ = inDisambig
				_ = inWd
				_ = tokDepth
				flush()
			}
		case xml.CharData:
			if inWd && cur != nil {
				s := strings.TrimSpace(string(t))
				if s != "" && cur.TargetLemma == "" {
					cur.TargetLemma = s
				}
			}
		}
	}
	flush()
	return rules, nil
}

func parseToken(dec *xml.Decoder, start xml.StartElement) (pattern.PatToken, error) {
	pt := pattern.PatToken{
		CaseSensitive: attr(start, "case_sensitive") == "yes",
		Regexp:        attr(start, "regexp") == "yes",
		Negate:        attr(start, "negate") == "yes",
		Inflected:     attr(start, "inflected") == "yes",
		Postag:        attr(start, "postag"),
		PostagRegexp:  attr(start, "postag_regexp") == "yes",
		Chunk:         attr(start, "chunk"),
		NegatePos:     attr(start, "negate_pos") == "yes",
		SpaceBefore:   attr(start, "spacebefore"),
		Min:           1,
		Max:           1,
	}
	if v := attr(start, "min"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			pt.Min = n
		}
	}
	if v := attr(start, "max"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			pt.Max = n
		}
	}
	if v := attr(start, "chunk_re"); v != "" {
		pt.Chunk = v // mark incomplete via Needs
	}
	var val strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return pt, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			_ = skip(dec, t)
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				pt.Value = strings.TrimSpace(val.String())
				if pt.Regexp && pt.Value != "" {
					if re, err := jregex.Compile(pt.Value, pt.CaseSensitive); err == nil {
						pt.Re = re
					}
				}
				return pt, nil
			}
		case xml.CharData:
			val.Write(t)
		}
	}
}

func attr(se xml.StartElement, name string) string {
	for _, a := range se.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func skip(dec *xml.Decoder, start xml.StartElement) error {
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	return nil
}

// LoadEnglish loads en disambiguation.xml from data root.
func LoadEnglish(dataRoot string) (*Engine, error) {
	p := filepath.Join(dataRoot, "languagetool-language-modules", "en", "src", "main", "resources",
		"org", "languagetool", "resource", "en", "disambiguation.xml")
	if _, err := os.Stat(p); err != nil {
		return nil, fmt.Errorf("disambiguation.xml: %w", err)
	}
	return LoadFile(p)
}
