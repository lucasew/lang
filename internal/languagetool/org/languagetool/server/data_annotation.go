package server

import (
	"encoding/json"
	"fmt"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
)

// dataAnnotationRoot is the soft subset of LT's data JSON parameter.
// Example: {"annotation":[{"text":"See "},{"markup":"<b>"},{"text":"a error"},{"markup":"</b>"}]}
type dataAnnotationRoot struct {
	Annotation []dataAnnotationPart `json:"annotation"`
}

type dataAnnotationPart struct {
	Text       string `json:"text"`
	Markup     string `json:"markup"`
	InterpretAs string `json:"interpretAs"`
}

// ParseDataAnnotation builds AnnotatedText from the API "data" JSON body.
func ParseDataAnnotation(dataJSON string) (*markup.AnnotatedText, error) {
	if dataJSON == "" {
		return nil, fmt.Errorf("empty data")
	}
	var root dataAnnotationRoot
	if err := json.Unmarshal([]byte(dataJSON), &root); err != nil {
		return nil, NewBadRequestError("Could not parse JSON from 'data' parameter: " + err.Error())
	}
	if len(root.Annotation) == 0 {
		return nil, NewBadRequestError("'data' JSON must contain non-empty 'annotation' array")
	}
	b := markup.NewAnnotatedTextBuilder()
	for _, p := range root.Annotation {
		if p.Markup != "" {
			if p.InterpretAs != "" {
				b.AddMarkupInterpretAs(p.Markup, p.InterpretAs)
			} else {
				b.AddMarkup(p.Markup)
			}
			continue
		}
		if p.Text != "" {
			b.AddText(p.Text)
		}
	}
	return b.Build(), nil
}
