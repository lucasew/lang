package tokenizers

import (
	_ "embed"
	"path/filepath"
	"sync"
	"unicode"

	"github.com/lucasew/lang/internal/attic/srx"
)

//go:embed data/segment-simple.srx
var segmentSimpleSRX []byte

// SimpleSentenceTokenizer ports org.languagetool.tokenizers.SimpleSentenceTokenizer.
//
// Java: extends SRXSentenceTokenizer(AnyLanguage, "/org/languagetool/tokenizers/segment-simple.srx")
// with AnyLanguage.getShortCode() == "xx".
type SimpleSentenceTokenizer struct {
	*SRXSentenceTokenizer
}

func NewSimpleSentenceTokenizer() *SimpleSentenceTokenizer {
	inner := NewSRXSentenceTokenizerWithPath("xx", "/org/languagetool/tokenizers/segment-simple.srx")
	return &SimpleSentenceTokenizer{SRXSentenceTokenizer: inner}
}

// SetSingleLineBreaksMarksParagraph forwards to the embedded SRXSentenceTokenizer.
func (t *SimpleSentenceTokenizer) SetSingleLineBreaksMarksParagraph(lineBreakParagraphs bool) {
	if t == nil || t.SRXSentenceTokenizer == nil {
		return
	}
	t.SRXSentenceTokenizer.SetSingleLineBreaksMarksParagraph(lineBreakParagraphs)
}

// SingleLineBreaksMarksPara forwards to the embedded SRXSentenceTokenizer.
func (t *SimpleSentenceTokenizer) SingleLineBreaksMarksPara() bool {
	if t == nil || t.SRXSentenceTokenizer == nil {
		return false
	}
	return t.SRXSentenceTokenizer.SingleLineBreaksMarksPara()
}

// Tokenize uses the embedded SRXSentenceTokenizer (segment-simple.srx + "xx").
func (t *SimpleSentenceTokenizer) Tokenize(text string) []string {
	if t == nil || t.SRXSentenceTokenizer == nil {
		return NewSimpleSentenceTokenizer().Tokenize(text)
	}
	return t.SRXSentenceTokenizer.Tokenize(text)
}

var (
	simpleDocOnce sync.Once
	simpleDoc     *srx.Document
	simpleDocErr  error
)

// segmentSimpleDocument loads the embedded official segment-simple.srx once
// (SrxTools.createSrxDocument for "/org/languagetool/tokenizers/segment-simple.srx").
func segmentSimpleDocument() (*srx.Document, error) {
	simpleDocOnce.Do(func() {
		candidates := []string{
			filepath.Join("internal", "languagetool", "org", "languagetool", "tokenizers", "data", "segment-simple.srx"),
			filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources",
				"org", "languagetool", "resource", "org", "languagetool", "tokenizers", "segment-simple.srx"),
		}
		for _, p := range candidates {
			if doc, err := srx.Load(p); err == nil && doc != nil {
				simpleDoc, simpleDocErr = doc, nil
				return
			}
		}
		name, err := materializeEmbed("segment-simple", segmentSimpleSRX)
		if err != nil {
			simpleDocErr = err
			return
		}
		simpleDoc, simpleDocErr = srx.Load(name)
	})
	return simpleDoc, simpleDocErr
}

// simpleSrxDefaultRulesTokenize mirrors segment-simple Default rules if SRX load fails.
// Java SRX uses Pattern.UNICODE_CHARACTER_CLASS so \s ≈ unicode.IsSpace.
func simpleSrxDefaultRulesTokenize(text string) []string {
	var out []string
	start := 0
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r != '.' && r != '!' && r != '?' && r != '…' {
			continue
		}
		j := i
		for j+1 < len(runes) {
			n := runes[j+1]
			if n == '.' || n == '!' || n == '?' || n == '…' {
				j++
				continue
			}
			break
		}
		if j+1 < len(runes) && unicode.IsSpace(runes[j+1]) {
			end := j + 2
			out = append(out, string(runes[start:end]))
			start = end
			i = end - 1
			continue
		}
		if j+1 < len(runes) && unicode.IsUpper(runes[j+1]) {
			end := j + 1
			out = append(out, string(runes[start:end]))
			start = end
			i = end - 1
			continue
		}
		i = j
	}
	if start < len(runes) {
		out = append(out, string(runes[start:]))
	}
	return out
}
