package tools

import "fmt"

// PseudoMatch ports org.languagetool.tools.PseudoMatch.
type PseudoMatch struct {
	replacements []string
	fromPos      int
	toPos        int
}

func NewPseudoMatch(replacement string, fromPos, toPos int) *PseudoMatch {
	return &PseudoMatch{
		replacements: []string{replacement},
		fromPos:      fromPos,
		toPos:        toPos,
	}
}

func (m *PseudoMatch) GetReplacements() []string { return m.replacements }
func (m *PseudoMatch) GetReplacement() string {
	if len(m.replacements) == 0 {
		return ""
	}
	return m.replacements[0]
}
func (m *PseudoMatch) GetFromPos() int { return m.fromPos }
func (m *PseudoMatch) GetToPos() int   { return m.toPos }
func (m *PseudoMatch) String() string {
	return fmt.Sprintf("%d-%d-%v", m.fromPos, m.toPos, m.replacements)
}
