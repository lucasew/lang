package wikipedia

import (
	"strings"
	"unicode"
)

// MatchSpan is a minimal stand-in for rules.RuleMatch positions + suggestions.
type MatchSpan struct {
	FromPos               int
	ToPos                 int
	SuggestedReplacements []string
}

// SuggestionReplacer ports org.languagetool.dev.wikipedia.SuggestionReplacer.
type SuggestionReplacer struct {
	mapping      *PlainTextMapping
	originalText string
	errorMarker  ErrorMarker
}

func NewSuggestionReplacer(mapping *PlainTextMapping, originalText string) *SuggestionReplacer {
	return NewSuggestionReplacerWithMarker(mapping, originalText, DefaultErrorMarker())
}

func NewSuggestionReplacerWithMarker(mapping *PlainTextMapping, originalText string, marker ErrorMarker) *SuggestionReplacer {
	return &SuggestionReplacer{
		mapping:      mapping,
		originalText: originalText,
		errorMarker:  marker,
	}
}

// ApplySuggestionsToOriginalText ports SuggestionReplacer.applySuggestionsToOriginalText.
// Positions are character offsets compatible with AbsolutePositionFor (one unit per rune for BMP).
func (r *SuggestionReplacer) ApplySuggestionsToOriginalText(match MatchSpan) ([]*RuleMatchApplication, error) {
	origRunes := []rune(r.originalText)
	plainRunes := []rune(r.mapping.GetPlainText())

	replacements := append([]string(nil), match.SuggestedReplacements...)
	hasReal := len(replacements) > 0
	if !hasReal {
		if match.FromPos < 0 || match.ToPos > len(plainRunes) || match.FromPos > match.ToPos {
			return nil, mappingError("match positions out of range")
		}
		replacements = append(replacements, string(plainRunes[match.FromPos:match.ToPos]))
	}

	fromLine, fromCol, err := r.mapping.OriginalTextPositionFor(match.FromPos + 1)
	if err != nil {
		return nil, err
	}
	toLine, toCol, err := r.mapping.OriginalTextPositionFor(match.ToPos + 1)
	if err != nil {
		return nil, err
	}
	fromPos, err := AbsolutePositionFor(fromLine, fromCol, r.originalText)
	if err != nil {
		return nil, err
	}
	toPos, err := AbsolutePositionFor(toLine, toCol, r.originalText)
	if err != nil {
		return nil, err
	}

	var out []*RuleMatchApplication
	errorText := string(plainRunes[match.FromPos:match.ToPos])
	for _, replacement := range replacements {
		contextFrom := FindNextWhitespaceToTheLeft(r.originalText, fromPos)
		contextTo := FindNextWhitespaceToTheRight(r.originalText, toPos)
		context := string(origRunes[contextFrom:contextTo])
		text := string(origRunes[:contextFrom]) +
			r.errorMarker.StartMarker + context + r.errorMarker.EndMarker +
			string(origRunes[contextTo:])

		real := hasReal
		var newContext string
		if strings.Count(context, errorText) == 1 {
			newContext = strings.Replace(context, errorText, replacement, 1)
		} else {
			newContext = context
			real = false
		}
		newText := string(origRunes[:contextFrom]) +
			r.errorMarker.StartMarker + newContext + r.errorMarker.EndMarker +
			string(origRunes[contextTo:])

		var app *RuleMatchApplication
		if real {
			app, err = ForMatchWithReplacement(text, newText, r.errorMarker, match.FromPos, match.ToPos)
		} else {
			app, err = ForMatchWithoutReplacement(text, newText, r.errorMarker, match.FromPos, match.ToPos)
		}
		if err != nil {
			return nil, err
		}
		out = append(out, app)
	}
	return out, nil
}

// FindNextWhitespaceToTheRight ports SuggestionReplacer.findNextWhitespaceToTheRight (rune offsets).
func FindNextWhitespaceToTheRight(text string, position int) int {
	runes := []rune(text)
	if position < 0 {
		position = 0
	}
	for i := position; i < len(runes); i++ {
		if unicode.IsSpace(runes[i]) {
			return i
		}
	}
	return len(runes)
}

// FindNextWhitespaceToTheLeft ports SuggestionReplacer.findNextWhitespaceToTheLeft (rune offsets).
func FindNextWhitespaceToTheLeft(text string, position int) int {
	runes := []rune(text)
	if position >= len(runes) {
		position = len(runes) - 1
	}
	for i := position; i >= 0; i-- {
		if unicode.IsSpace(runes[i]) {
			return i + 1
		}
	}
	return 0
}
