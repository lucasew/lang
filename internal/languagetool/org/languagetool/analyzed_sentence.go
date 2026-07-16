package languagetool

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"

// AnalyzedSentence ports org.languagetool.AnalyzedSentence (stub).
type AnalyzedSentence struct{}

func NewAnalyzedSentence(words []*AnalyzedTokenReadings) *AnalyzedSentence {
	tools.Unimplemented("AnalyzedSentence.NewAnalyzedSentence")
	return nil
}

func (s *AnalyzedSentence) String() string {
	tools.Unimplemented("AnalyzedSentence.String")
	return ""
}

func (s *AnalyzedSentence) Copy(other *AnalyzedSentence) *AnalyzedSentence {
	tools.Unimplemented("AnalyzedSentence.Copy")
	return nil
}

func (s *AnalyzedSentence) Equals(o *AnalyzedSentence) bool {
	tools.Unimplemented("AnalyzedSentence.Equals")
	return false
}
