package tokenizers

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/attic/srx"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SRXSentenceTokenizer ports org.languagetool.tokenizers.SRXSentenceTokenizer.
// It applies LanguageTool's official segment.srx (embedded) via attic/srx, matching
// Java's loomchild SrxTextIterator with cascade="yes".
type SRXSentenceTokenizer struct {
	LanguageCode string
	// SrxPath is the resource path (default /segment.srx) for API parity.
	SrxPath string
	// Segment is an optional custom segmenter; nil uses embedded segment.srx.
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

// Tokenize splits text into sentences using embedded segment.srx (Java parity).
// languageCode is passed as Java Language.getShortCode() would be; maps match
// both "pt_two" and "pt-PT_two" via (PT|pt).*.
func (t *SRXSentenceTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
	if t.Segment != nil {
		return t.Segment(text, t.LanguageCode+t.parCode)
	}
	doc, err := srx.DefaultDocument()
	if err != nil || doc == nil {
		// Fallback only if embed/parse failed (should not happen in normal builds).
		return defaultSrxLikeTokenize(text, t.parCode)
	}
	return doc.Split(text, t.LanguageCode, t.parCode)
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
		// Fold segments that Java String.trim() would treat as empty (not Unicode Zs-only).
		if len(out) > 0 && tools.JavaStringTrimIsEmpty(s) {
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
// Forwards paragraph mode to SimpleSentenceTokenizer (segment-simple.srx _one/_two maps).
type sentenceTokenizerAdapter struct {
	*SimpleSentenceTokenizer
}

func (a *sentenceTokenizerAdapter) SetSingleLineBreaksMarksParagraph(v bool) {
	if a != nil && a.SimpleSentenceTokenizer != nil {
		a.SimpleSentenceTokenizer.SetSingleLineBreaksMarksParagraph(v)
	}
}
func (a *sentenceTokenizerAdapter) SingleLineBreaksMarksPara() bool {
	if a == nil || a.SimpleSentenceTokenizer == nil {
		return false
	}
	return a.SimpleSentenceTokenizer.SingleLineBreaksMarksPara()
}

// AsSentenceTokenizer wraps SimpleSentenceTokenizer.
func (t *SimpleSentenceTokenizer) AsSentenceTokenizer() SentenceTokenizer {
	return &sentenceTokenizerAdapter{SimpleSentenceTokenizer: t}
}

// isUpper is reserved for future abbreviation-aware breaks.
var _ = unicode.IsUpper
