package languagetool

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"

// AnalyzedTokenReadings ports org.languagetool.AnalyzedTokenReadings (stub — fill 1:1 in Phase 2).
type AnalyzedTokenReadings struct {
	// fields filled when implemented
}

func NewAnalyzedTokenReadings(tok *AnalyzedToken) *AnalyzedTokenReadings {
	tools.Unimplemented("AnalyzedTokenReadings.NewAnalyzedTokenReadings")
	return nil
}

func NewAnalyzedTokenReadingsList(tokens []*AnalyzedToken, startPos int) *AnalyzedTokenReadings {
	tools.Unimplemented("AnalyzedTokenReadings.NewAnalyzedTokenReadingsList")
	return nil
}

func (r *AnalyzedTokenReadings) IsLinebreak() bool {
	tools.Unimplemented("AnalyzedTokenReadings.IsLinebreak")
	return false
}
func (r *AnalyzedTokenReadings) IsSentenceEnd() bool {
	tools.Unimplemented("AnalyzedTokenReadings.IsSentenceEnd")
	return false
}
func (r *AnalyzedTokenReadings) IsParagraphEnd() bool {
	tools.Unimplemented("AnalyzedTokenReadings.IsParagraphEnd")
	return false
}
func (r *AnalyzedTokenReadings) IsSentenceStart() bool {
	tools.Unimplemented("AnalyzedTokenReadings.IsSentenceStart")
	return false
}
func (r *AnalyzedTokenReadings) SetSentEnd() {
	tools.Unimplemented("AnalyzedTokenReadings.SetSentEnd")
}
func (r *AnalyzedTokenReadings) AddReading(tok *AnalyzedToken, ruleID string) {
	tools.Unimplemented("AnalyzedTokenReadings.AddReading")
}
func (r *AnalyzedTokenReadings) GetAnalyzedToken(n int) *AnalyzedToken {
	tools.Unimplemented("AnalyzedTokenReadings.GetAnalyzedToken")
	return nil
}
func (r *AnalyzedTokenReadings) RemoveReading(tok *AnalyzedToken, ruleID string) {
	tools.Unimplemented("AnalyzedTokenReadings.RemoveReading")
}
func (r *AnalyzedTokenReadings) LeaveReading(tok *AnalyzedToken) {
	tools.Unimplemented("AnalyzedTokenReadings.LeaveReading")
}
func (r *AnalyzedTokenReadings) GetReadingsLength() int {
	tools.Unimplemented("AnalyzedTokenReadings.GetReadingsLength")
	return 0
}
func (r *AnalyzedTokenReadings) GetToken() string {
	tools.Unimplemented("AnalyzedTokenReadings.GetToken")
	return ""
}
func (r *AnalyzedTokenReadings) HasPosTag(pos string) bool {
	tools.Unimplemented("AnalyzedTokenReadings.HasPosTag")
	return false
}
func (r *AnalyzedTokenReadings) HasPartialPosTag(pos string) bool {
	tools.Unimplemented("AnalyzedTokenReadings.HasPartialPosTag")
	return false
}
func (r *AnalyzedTokenReadings) MatchesPosTagRegex(re string) bool {
	tools.Unimplemented("AnalyzedTokenReadings.MatchesPosTagRegex")
	return false
}
func (r *AnalyzedTokenReadings) String() string {
	tools.Unimplemented("AnalyzedTokenReadings.String")
	return ""
}
func (r *AnalyzedTokenReadings) Immunize(n int) {
	tools.Unimplemented("AnalyzedTokenReadings.Immunize")
}

// Readings returns a snapshot for range-loop parity with Java Iterable.
func (r *AnalyzedTokenReadings) Readings() []*AnalyzedToken {
	tools.Unimplemented("AnalyzedTokenReadings.Readings")
	return nil
}
