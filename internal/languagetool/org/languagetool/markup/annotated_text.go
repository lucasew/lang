package markup

import (
	"fmt"
	"math"
	"strings"
)

// MetaDataKey ports AnnotatedText.MetaDataKey.
type MetaDataKey string

const (
	MetaDocumentTitle         MetaDataKey = "DocumentTitle"
	MetaEmailToAddress        MetaDataKey = "EmailToAddress"
	MetaFullName              MetaDataKey = "FullName"
	MetaEmailNumberOfAttachments MetaDataKey = "EmailNumberOfAttachments"
)

// AnnotatedText ports org.languagetool.markup.AnnotatedText.
type AnnotatedText struct {
	parts          []TextPart
	mapping        map[int]MappingValue // plain text pos → original mapping
	metaData       map[MetaDataKey]string
	customMetaData map[string]string
}

func (t *AnnotatedText) GetParts() []TextPart { return t.parts }

// GetOriginalText — plain text without markup and without FAKE_CONTENT.
func (t *AnnotatedText) GetOriginalText() string {
	var b strings.Builder
	for _, p := range t.parts {
		if p.Type == TextPartText {
			b.WriteString(p.Part)
		}
	}
	return b.String()
}

// GetPlainText — without markup but with FAKE_CONTENT (interpretAs).
func (t *AnnotatedText) GetPlainText() string {
	var b strings.Builder
	for _, p := range t.parts {
		if p.Type == TextPartText || p.Type == TextPartFakeContent {
			b.WriteString(p.Part)
		}
	}
	return b.String()
}

// GetTextWithMarkup — markup + text, without FAKE_CONTENT.
func (t *AnnotatedText) GetTextWithMarkup() string {
	var b strings.Builder
	for _, p := range t.parts {
		if p.Type != TextPartFakeContent {
			b.WriteString(p.Part)
		}
	}
	return b.String()
}

// GetOriginalTextPositionFor maps a plain-text offset to the markup text offset.
func (t *AnnotatedText) GetOriginalTextPositionFor(plainTextPosition int, isToPos bool) int {
	if plainTextPosition < 0 {
		panic(fmt.Sprintf("plainTextPosition must be >= 0: %d", plainTextPosition))
	}
	if len(t.mapping) == 0 {
		return 0
	}
	minDiff := math.MaxInt32
	var bestMatch *MappingValue
	// find the closest higher mapping key
	for maybeClosePosition, val := range t.mapping {
		if plainTextPosition < maybeClosePosition {
			diff := maybeClosePosition - plainTextPosition
			if diff > 0 && diff < minDiff {
				v := val
				bestMatch = &v
				minDiff = diff
			}
		}
	}
	if bestMatch == nil {
		panic(fmt.Sprintf("Could not map %d to original position. isToPos: %v", plainTextPosition, isToPos))
	}
	if !isToPos && bestMatch.FakeMarkupLength > 0 {
		minDiff = bestMatch.FakeMarkupLength
	}
	return bestMatch.TotalPosition - minDiff
}

func (t *AnnotatedText) GetGlobalMetaDataString(key, defaultValue string) string {
	if v, ok := t.customMetaData[key]; ok {
		return v
	}
	return defaultValue
}

func (t *AnnotatedText) GetGlobalMetaDataKey(key MetaDataKey, defaultValue string) string {
	if v, ok := t.metaData[key]; ok {
		return v
	}
	return defaultValue
}

func (t *AnnotatedText) String() string {
	var b strings.Builder
	for _, p := range t.parts {
		b.WriteString(p.Part)
	}
	return b.String()
}
