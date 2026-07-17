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
		// two or more newlines end a paragraph; keep breaks on the preceding segment
		// (Java SRX: "He won't\n\n" / "Really.")
		paras = splitParagraphsTwoBreaks(text)
	}
	var out []string
	for _, p := range paras {
		if p == "" {
			continue
		}
		sents := simpleSentenceSplit(p)
		sents = mergeTrailingWhitespaceSents(sents)
		out = append(out, sents...)
	}
	if len(out) == 0 && text != "" {
		return []string{text}
	}
	return out
}

// splitParagraphsTwoBreaks splits on \n{2,}, attaching the break to the previous paragraph.
func splitParagraphsTwoBreaks(text string) []string {
	if text == "" {
		return nil
	}
	re := regexp.MustCompile(`\n{2,}`)
	idxs := re.FindAllStringIndex(text, -1)
	if len(idxs) == 0 {
		return []string{text}
	}
	var paras []string
	start := 0
	for _, m := range idxs {
		paras = append(paras, text[start:m[1]])
		start = m[1]
	}
	if start < len(text) {
		paras = append(paras, text[start:])
	}
	var out []string
	for _, p := range paras {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// mergeTrailingWhitespaceSents folds pure-whitespace segments into the previous sentence
// so paragraph breaks stay on the last content sentence (EMPTY_LINE / isParagraphEnd parity).
func mergeTrailingWhitespaceSents(sents []string) []string {
	if len(sents) <= 1 {
		return sents
	}
	out := make([]string, 0, len(sents))
	for _, s := range sents {
		if len(out) > 0 && strings.TrimSpace(s) == "" {
			out[len(out)-1] += s
			continue
		}
		out = append(out, s)
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
