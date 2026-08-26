package markup

import "strings"

// hiddenChars from AnnotatedTextBuilder — soft hyphen, LRO, PDF.
const hiddenChars = "\u00AD\u202D\u202C"

// AnnotatedTextBuilder ports org.languagetool.markup.AnnotatedTextBuilder.
type AnnotatedTextBuilder struct {
	parts          []TextPart
	metaData       map[MetaDataKey]string
	customMetaData map[string]string
}

func NewAnnotatedTextBuilder() *AnnotatedTextBuilder {
	return &AnnotatedTextBuilder{
		metaData:       map[MetaDataKey]string{},
		customMetaData: map[string]string{},
	}
}

func (b *AnnotatedTextBuilder) AddGlobalMetaDataKey(key MetaDataKey, value string) *AnnotatedTextBuilder {
	b.metaData[key] = value
	return b
}

func (b *AnnotatedTextBuilder) AddGlobalMetaData(key, value string) *AnnotatedTextBuilder {
	b.customMetaData[key] = value
	return b
}

func (b *AnnotatedTextBuilder) AddText(text string) *AnnotatedTextBuilder {
	// StringTokenizer with delimiters returned
	for _, tok := range tokenizeKeepHidden(text) {
		if strings.Contains(hiddenChars, tok) {
			b.parts = append(b.parts, NewTextPart(tok, TextPartMarkup))
		} else {
			b.parts = append(b.parts, NewTextPart(tok, TextPartText))
		}
	}
	return b
}

func tokenizeKeepHidden(text string) []string {
	if text == "" {
		return nil
	}
	var out []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for _, r := range text {
		if strings.ContainsRune(hiddenChars, r) {
			flush()
			out = append(out, string(r))
		} else {
			cur.WriteRune(r)
		}
	}
	flush()
	return out
}

func (b *AnnotatedTextBuilder) AddMarkup(markup string) *AnnotatedTextBuilder {
	b.parts = append(b.parts, NewTextPart(markup, TextPartMarkup))
	return b
}

func (b *AnnotatedTextBuilder) AddMarkupInterpretAs(markup, interpretAs string) *AnnotatedTextBuilder {
	b.parts = append(b.parts, NewTextPart(markup, TextPartMarkup))
	b.parts = append(b.parts, NewTextPart(interpretAs, TextPartFakeContent))
	return b
}

func (b *AnnotatedTextBuilder) Add(part TextPart) {
	b.parts = append(b.parts, part)
}

// Build creates the AnnotatedText with plain→original position mapping.
// Lengths use UTF-16 code units (Java String).
func (b *AnnotatedTextBuilder) Build() *AnnotatedText {
	plainTextPosition := 0
	totalPosition := 0
	mapping := map[int]MappingValue{}
	for i := 0; i < len(b.parts); i++ {
		part := b.parts[i]
		switch part.Type {
		case TextPartText:
			plainTextPosition += utf16Len(part.Part)
			totalPosition += utf16Len(part.Part)
			mapping[plainTextPosition] = MappingValue{TotalPosition: totalPosition}
		case TextPartMarkup:
			totalPosition += utf16Len(part.Part)
			if hasFakeContent(i, b.parts) {
				plainTextPosition += utf16Len(b.parts[i+1].Part)
				i++
				if _, ok := mapping[plainTextPosition]; !ok {
					mapping[plainTextPosition] = MappingValue{
						TotalPosition:    totalPosition,
						FakeMarkupLength: utf16Len(part.Part),
					}
				}
			}
		}
	}
	// copy maps
	meta := map[MetaDataKey]string{}
	for k, v := range b.metaData {
		meta[k] = v
	}
	custom := map[string]string{}
	for k, v := range b.customMetaData {
		custom[k] = v
	}
	return &AnnotatedText{
		parts:          append([]TextPart(nil), b.parts...),
		mapping:        mapping,
		metaData:       meta,
		customMetaData: custom,
	}
}

func hasFakeContent(i int, parts []TextPart) bool {
	next := i + 1
	return next < len(parts) && parts[next].Type == TextPartFakeContent
}

func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}
