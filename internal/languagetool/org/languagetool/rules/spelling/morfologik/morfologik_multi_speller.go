package morfologik

// MorfologikMultiSpeller ports org.languagetool.rules.spelling.morfologik.MorfologikMultiSpeller
// as an ordered list of spellers (user dict, main dict, ...).
type MorfologikMultiSpeller struct {
	Spellers []*MorfologikSpeller
}

func NewMorfologikMultiSpeller(spellers ...*MorfologikSpeller) *MorfologikMultiSpeller {
	return &MorfologikMultiSpeller{Spellers: append([]*MorfologikSpeller(nil), spellers...)}
}

// IsMisspelled is true only if all spellers consider the word misspelled.
func (m *MorfologikMultiSpeller) IsMisspelled(word string) bool {
	if m == nil || len(m.Spellers) == 0 {
		return false
	}
	for _, s := range m.Spellers {
		if s != nil && !s.IsMisspelled(word) {
			return false
		}
	}
	return true
}

// GetSuggestions merges replacements from all spellers (first-seen order).
func (m *MorfologikMultiSpeller) GetSuggestions(word string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range m.Spellers {
		if s == nil {
			continue
		}
		for _, sug := range s.FindReplacements(word) {
			if _, ok := seen[sug]; ok {
				continue
			}
			seen[sug] = struct{}{}
			out = append(out, sug)
		}
	}
	return out
}
