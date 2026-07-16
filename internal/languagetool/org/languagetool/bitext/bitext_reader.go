package bitext

// BitextReader ports org.languagetool.bitext.BitextReader.
type BitextReader interface {
	// Next returns the next pair and whether one was available.
	Next() (StringPair, bool)
	GetLineCount() int
	GetColumnCount() int
	GetTargetColumnCount() int
	GetSentencePosition() int
	GetCurrentLine() string
}
