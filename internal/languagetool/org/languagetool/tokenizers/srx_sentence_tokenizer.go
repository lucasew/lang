package tokenizers

import (
	"regexp"
	"strings"
	"unicode"
)

// SRXSentenceTokenizer ports org.languagetool.tokenizers.SRXSentenceTokenizer.
// Full segment.srx (loomchild) is not embedded; this applies a practical SRX-like
// rule set for common Latin punctuation, with language code reserved for later.
type SRXSentenceTokenizer struct {
	LanguageCode string
	// SrxPath is the resource path (default /segment.srx) for API parity.
	SrxPath string
	// Segment is an optional custom segmenter; nil uses built-in rules.
	Segment func(text, languageCode string) []string

	parCode string // "_one" or "_two"
}

func NewSRXSentenceTokenizer(languageCode string) *SRXSentenceTokenizer {
	return NewSRXSentenceTokenizerWithPath(languageCode, "/segment.srx")
}

func NewSRXSentenceTokenizerWithPath(languageCode, srxInClassPath string) *SRXSentenceTokenizer {
	t := &SRXSentenceTokenizer{
		LanguageCode: languageCode,
		SrxPath:      srxInClassPath,
	}
	t.SetSingleLineBreaksMarksParagraph(false)
	return t
}

func (t *SRXSentenceTokenizer) SetSingleLineBreaksMarksParagraph(lineBreakParagraphs bool) {
	if lineBreakParagraphs {
		t.parCode = "_one"
	} else {
		t.parCode = "_two"
	}
}

func (t *SRXSentenceTokenizer) SingleLineBreaksMarksPara() bool {
	return t.parCode == "_one"
}

// Tokenize splits text into sentences.
func (t *SRXSentenceTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
	code := t.LanguageCode + t.parCode
	if t.Segment != nil {
		return t.Segment(text, code)
	}
	return defaultSrxLikeTokenize(text, t.parCode)
}

func defaultSrxLikeTokenize(text, parCode string) []string {
	// paragraph splits first
	var paras []string
	if parCode == "_one" {
		paras = splitKeep(text, "\n")
	} else {
		// two or more newlines end a paragraph
		re := regexp.MustCompile(`\n{2,}`)
		parts := re.Split(text, -1)
		// re-attach separators approximately
		paras = parts
	}
	var out []string
	for _, p := range paras {
		if p == "" {
			continue
		}
		out = append(out, simpleSentenceSplit(p)...)
	}
	if len(out) == 0 && text != "" {
		return []string{text}
	}
	return out
}

func simpleSentenceSplit(text string) []string {
	// reuse SimpleSentenceTokenizer logic
	return NewSimpleSentenceTokenizer().Tokenize(text)
}

// splitKeep splits on sep, keeping sep attached to preceding segment when possible.
func splitKeep(text, sep string) []string {
	if text == "" {
		return nil
	}
	parts := strings.SplitAfter(text, sep)
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Ensure interfaces.
var (
	_ Tokenizer         = (*SRXSentenceTokenizer)(nil)
	_ SentenceTokenizer = (*SRXSentenceTokenizer)(nil)
	_ SentenceTokenizer = (*sentenceTokenizerAdapter)(nil)
)

// sentenceTokenizerAdapter lets SimpleSentenceTokenizer satisfy SentenceTokenizer.
type sentenceTokenizerAdapter struct {
	*SimpleSentenceTokenizer
	single bool
}

func (a *sentenceTokenizerAdapter) SetSingleLineBreaksMarksParagraph(v bool) { a.single = v }
func (a *sentenceTokenizerAdapter) SingleLineBreaksMarksPara() bool          { return a.single }

// AsSentenceTokenizer wraps SimpleSentenceTokenizer.
func (t *SimpleSentenceTokenizer) AsSentenceTokenizer() SentenceTokenizer {
	return &sentenceTokenizerAdapter{SimpleSentenceTokenizer: t}
}

// isUpper is reserved for future abbreviation-aware breaks.
var _ = unicode.IsUpper
