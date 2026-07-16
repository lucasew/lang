package pattern

import "regexp"

// Rule is one LanguageTool pattern rule (or rulegroup member).
type Rule struct {
	ID       string
	SubID    string
	Name     string
	Category string
	Default  string // "on", "off", "temp_off"
	Tokens   []PatToken
	Anti     [][]PatToken // antipatterns
	Message  string
	ShortMsg string
	// Suggestions are raw XML text of <suggestion> elements (may contain <match>).
	Suggestions []string
	// Examples for testing / goldens
	Examples []Example
	// RequiresPOS is true if any token needs postag/chunk/inflected lemma matching.
	RequiresPOS bool
	// IssueType from rule attribute if present
	IssueType string
}

// FullID is id or id[subId].
func (r *Rule) FullID() string {
	if r.SubID == "" {
		return r.ID
	}
	return r.ID + "[" + r.SubID + "]"
}

// Example from grammar.xml.
type Example struct {
	// Correction is empty for 'correct' examples; otherwise expected replacement text of marker.
	Correction string
	// Text is example text with markers stripped.
	Text string
	// MarkerFrom/To are rune offsets of <marker> span in Text.
	MarkerFrom int
	MarkerTo   int
	HasMarker  bool
}

// PatToken is one <token> in a pattern.
type PatToken struct {
	Value         string
	Regexp        bool
	CaseSensitive bool
	Negate        bool
	Inflected     bool
	Min           int // default 1; 0 means optional
	Max           int // default 1; -1 means unlimited skip via skip attr separately
	Skip          int // max tokens to skip after this one
	SpaceBefore   string // "", "yes", "no"
	Postag        string
	PostagRegexp  bool
	Chunk         string
	NegatePos     bool
	InsideMarker  bool
	// Compiled regexp if Regexp
	Re *regexp.Regexp
	// Exceptions: nested <exception> tokens
	Exceptions []PatToken
	// And/Or groups (simplified: and means all must match same token)
	And []PatToken
	Or  []PatToken
}

// NeedsPOS reports whether matching requires morphological analysis.
func (t PatToken) NeedsPOS() bool {
	if t.Postag != "" || t.Chunk != "" || t.Inflected {
		return true
	}
	for _, e := range t.Exceptions {
		if e.NeedsPOS() {
			return true
		}
	}
	for _, e := range t.And {
		if e.NeedsPOS() {
			return true
		}
	}
	for _, e := range t.Or {
		if e.NeedsPOS() {
			return true
		}
	}
	return false
}
