package bitext

// BitextReader ports org.languagetool.bitext.BitextReader.
// Java extends Iterable<StringPair>; in Go iteration is HasNext/Next on
// concrete readers (TabBitextReader, WordFastTMReader), not on this interface.
type BitextReader interface {
	GetLineCount() int
	GetColumnCount() int
	GetTargetColumnCount() int
	GetSentencePosition() int
	GetCurrentLine() string
}

// Compile-time checks: concrete readers implement BitextReader getters.
var (
	_ BitextReader = (*TabBitextReader)(nil)
	_ BitextReader = (*WordFastTMReader)(nil)
)
