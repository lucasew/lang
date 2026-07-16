package wikipedia

import (
	"fmt"
	"strings"
)

// RuleMatchApplication ports org.languagetool.dev.wikipedia.RuleMatchApplication
// without depending on rules.RuleMatch (match identity is optional).
type RuleMatchApplication struct {
	Text               string
	TextWithCorrection string
	ErrorMarker        ErrorMarker
	HasRealReplacement bool
	// FromPos/ToPos of the underlying match in plain text (for debugging/tests).
	FromPos, ToPos int
}

func ForMatchWithReplacement(text, textWithCorrection string, marker ErrorMarker, from, to int) (*RuleMatchApplication, error) {
	return newApplication(text, textWithCorrection, marker, true, from, to)
}

func ForMatchWithoutReplacement(text, textWithCorrection string, marker ErrorMarker, from, to int) (*RuleMatchApplication, error) {
	return newApplication(text, textWithCorrection, marker, false, from, to)
}

func newApplication(text, textWithCorrection string, marker ErrorMarker, real bool, from, to int) (*RuleMatchApplication, error) {
	if !strings.Contains(textWithCorrection, marker.StartMarker) {
		return nil, fmt.Errorf("no start error marker (%s) found in text with correction", marker.StartMarker)
	}
	if !strings.Contains(textWithCorrection, marker.EndMarker) {
		return nil, fmt.Errorf("no end error marker (%s) found in text with correction", marker.EndMarker)
	}
	return &RuleMatchApplication{
		Text:               text,
		TextWithCorrection: textWithCorrection,
		ErrorMarker:        marker,
		HasRealReplacement: real,
		FromPos:            from,
		ToPos:              to,
	}, nil
}

func (a *RuleMatchApplication) GetOriginalText() string       { return a.Text }
func (a *RuleMatchApplication) GetTextWithCorrection() string { return a.TextWithCorrection }
func (a *RuleMatchApplication) GetErrorMarker() ErrorMarker   { return a.ErrorMarker }
func (a *RuleMatchApplication) HasRealRepl() bool             { return a.HasRealReplacement }
