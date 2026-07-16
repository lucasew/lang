package ca

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

const (
	windowForward  = 10
	windowBackward = 6
)

// CatalanMultitokenDisambiguator ports
// org.languagetool.tagging.disambiguation.ca.CatalanMultitokenDisambiguator.
// IsMisspelled is injectable (Morfologik multitoken speller deferred).
type CatalanMultitokenDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// IsMisspelled returns true if the multiword phrase is unknown.
	IsMisspelled func(phrase string) bool
}

func NewCatalanMultitokenDisambiguator() *CatalanMultitokenDisambiguator {
	return &CatalanMultitokenDisambiguator{}
}

var dictionaryFixes = map[string]struct{}{
	"Santa María": {},
	"San Agustin": {},
}

type searchType int

const (
	searchNone searchType = iota
	searchShrinkFromEnd
	searchShrinkFromStart
)

func (d *CatalanMultitokenDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil || d.IsMisspelled == nil {
		return input
	}
	tokens := append([]*languagetool.AnalyzedTokenReadings(nil), input.GetTokens()...)
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil || tok.IsWhitespace() || tok.IsTagged() {
			continue
		}
		surface := tok.GetToken()
		if surface == "" {
			continue
		}
		from, to := getTitleCaseIndexes(tokens, i)
		found := d.searchInDictAndTag(tokens, from, to, searchNone)
		if !found {
			r, _ := utf8.DecodeRuneInString(surface)
			if unicode.IsUpper(r) {
				toFwd := i + windowForward
				if toFwd > len(tokens)-1 {
					toFwd = len(tokens) - 1
				}
				found = d.searchInDictAndTag(tokens, i, toFwd, searchShrinkFromEnd)
			}
		}
		if !found {
			fromBwd := i - windowBackward
			if fromBwd < 1 {
				fromBwd = 1
			}
			d.searchInDictAndTag(tokens, fromBwd, i, searchShrinkFromStart)
		}
	}
	return languagetool.NewAnalyzedSentence(tokens)
}

func (d *CatalanMultitokenDisambiguator) searchInDictAndTag(tokens []*languagetool.AnalyzedTokenReadings, from, to int, shrink searchType) bool {
	currentFrom, currentTo := from, to
	for currentTo > currentFrom {
		text := getTextFromTo(tokens, currentFrom, currentTo)
		if _, fix := dictionaryFixes[text]; fix {
			return false
		}
		if text != "" && !strings.HasPrefix(text, " ") && !strings.HasSuffix(text, " ") && !d.IsMisspelled(text) {
			pos := "NPCNM00"
			for j := currentFrom; j <= currentTo; j++ {
				if tokens[j] == nil || tokens[j].IsWhitespace() {
					continue
				}
				at := languagetool.NewAnalyzedToken(tokens[j].GetToken(), &pos, &text)
				tokens[j].AddReading(at, "HybridDisamb")
			}
			return true
		}
		switch shrink {
		case searchShrinkFromEnd:
			currentTo--
		case searchShrinkFromStart:
			currentFrom++
		default:
			return false
		}
	}
	return false
}

func getTitleCaseIndexes(tokens []*languagetool.AnalyzedTokenReadings, startIndex int) (fromIndex, toIndex int) {
	fromIndex = startIndex
	for fromIndex > 1 {
		prev := tokens[fromIndex-1]
		if prev == nil {
			break
		}
		t := prev.GetToken()
		if t == "" {
			break
		}
		r, _ := utf8.DecodeRuneInString(t)
		if unicode.IsUpper(r) || prev.IsWhitespace() || utf8.RuneCountInString(t) < 3 {
			fromIndex--
			continue
		}
		break
	}
	for fromIndex < startIndex {
		t := tokens[fromIndex].GetToken()
		r, _ := utf8.DecodeRuneInString(t)
		if unicode.IsUpper(r) {
			break
		}
		fromIndex++
	}
	toIndex = startIndex
	for toIndex < len(tokens)-1 {
		next := tokens[toIndex+1]
		if next == nil {
			break
		}
		t := next.GetToken()
		if t == "" {
			break
		}
		r, _ := utf8.DecodeRuneInString(t)
		if unicode.IsUpper(r) || next.IsWhitespace() || utf8.RuneCountInString(t) < 3 {
			toIndex++
			continue
		}
		break
	}
	for toIndex > startIndex {
		t := tokens[toIndex].GetToken()
		r, _ := utf8.DecodeRuneInString(t)
		if unicode.IsUpper(r) {
			break
		}
		toIndex--
	}
	return fromIndex, toIndex
}

func getTextFromTo(tokens []*languagetool.AnalyzedTokenReadings, from, to int) string {
	var b strings.Builder
	for i := from; i <= to; i++ {
		if i < 0 || i >= len(tokens) || tokens[i] == nil {
			return ""
		}
		b.WriteString(tokens[i].GetToken())
	}
	return b.String()
}
