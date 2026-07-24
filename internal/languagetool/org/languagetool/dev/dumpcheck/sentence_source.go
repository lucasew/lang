package dumpcheck

// SentenceSource ports org.languagetool.dev.dumpcheck.SentenceSource as a Go interface.
type SentenceSource interface {
	HasNext() bool
	Next() (Sentence, error)
	GetSource() string
}
