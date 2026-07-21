package tokenizers

import (
	_ "embed"
	"os"
	"path/filepath"
	"sync"
	"unicode"

	"github.com/lucasew/lang/internal/attic/srx"
)

//go:embed data/segment-simple.srx
var segmentSimpleSRX []byte

// SimpleSentenceTokenizer ports org.languagetool.tokenizers.SimpleSentenceTokenizer.
//
// Java extends SRXSentenceTokenizer with
// "/org/languagetool/tokenizers/segment-simple.srx" and AnyLanguage shortCode "xx".
// Loads the same segment-simple.srx (embedded) via attic/srx for Default rules only
// (map pattern ".*" → Default; ByLineBreak/ByTwoLineBreaks require _one/_two codes).
type SimpleSentenceTokenizer struct {
	parCode string // "_one" or "_two" — same as SRXSentenceTokenizer
}

func NewSimpleSentenceTokenizer() *SimpleSentenceTokenizer {
	t := &SimpleSentenceTokenizer{}
	t.SetSingleLineBreaksMarksParagraph(false)
	return t
}

func (t *SimpleSentenceTokenizer) SetSingleLineBreaksMarksParagraph(lineBreakParagraphs bool) {
	if t == nil {
		return
	}
	if lineBreakParagraphs {
		t.parCode = "_one"
	} else {
		t.parCode = "_two"
	}
}

func (t *SimpleSentenceTokenizer) SingleLineBreaksMarksPara() bool {
	return t != nil && t.parCode == "_one"
}

// Tokenize ports SRXSentenceTokenizer.tokenize with segment-simple.srx.
func (t *SimpleSentenceTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
	par := "_two"
	if t != nil && t.parCode != "" {
		par = t.parCode
	}
	doc, err := segmentSimpleDocument()
	if err != nil || doc == nil {
		// Fail-closed to local Default-rule twin (same segment-simple Default rules).
		return simpleSrxDefaultRulesTokenize(text)
	}
	// Java AnyLanguage.getShortCode() → "xx"
	return doc.Split(text, "xx", par)
}

var (
	simpleDocOnce sync.Once
	simpleDoc     *srx.Document
	simpleDocErr  error
)

// segmentSimpleDocument loads the embedded official segment-simple.srx once.
// attic/srx only exposes Load(path); materialize embed to a temp file (read-only use).
func segmentSimpleDocument() (*srx.Document, error) {
	simpleDocOnce.Do(func() {
		// Prefer in-tree copy next to this package for deterministic Load without temp I/O when CWD is module root.
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
		// Materialize embed (works regardless of CWD).
		f, err := os.CreateTemp("", "segment-simple-*.srx")
		if err != nil {
			simpleDocErr = err
			return
		}
		name := f.Name()
		if _, err := f.Write(segmentSimpleSRX); err != nil {
			_ = f.Close()
			simpleDocErr = err
			return
		}
		_ = f.Close()
		simpleDoc, simpleDocErr = srx.Load(name)
		// leave temp file for process lifetime (Load may re-open); best-effort cleanup not required
	})
	return simpleDoc, simpleDocErr
}

// simpleSrxDefaultRulesTokenize is the segment-simple.srx Default languagerule only,
// used if SRX load fails. Java SRX uses Pattern.UNICODE_CHARACTER_CLASS so \s ≈ unicode.IsSpace.
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
		// beforebreak [\.!?…]\s → break after one whitespace
		if j+1 < len(runes) && unicode.IsSpace(runes[j+1]) {
			end := j + 2
			if end > len(runes) {
				end = len(runes)
			}
			out = append(out, string(runes[start:end]))
			start = end
			i = end - 1
			continue
		}
		// beforebreak [\.!?…]\p{Lu} → break before uppercase
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
