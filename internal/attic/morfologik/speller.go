package morfologik

// Speller ports morfologik.speller.Speller (2.2.0).
//
// Unlike Dictionary.IsInDictionary (cold / per-call local), a Speller keeps
// containsSeparators sticky across isInDictionary calls, matching the Java field
// mutation. One Speller is created per MorfologikSpeller / runtime dict load;
// Dictionary itself stays immutable and cache-safe.
type Speller struct {
	*SpellerFSA
}

// NewSpeller ports Speller(Dictionary, editDistance).
func NewSpeller(dict *Dictionary, editDistance int) *Speller {
	return &Speller{SpellerFSA: NewSpellerFSA(dict, editDistance)}
}

// ContainsSeparators reports the sticky Speller.containsSeparators field (tests / debug).
func (s *Speller) ContainsSeparators() bool {
	if s == nil || s.SpellerFSA == nil {
		return true
	}
	return s.containsSeparators
}

// IsInDictionary ports Speller.isInDictionary; mutates containsSeparators (Java).
func (s *Speller) IsInDictionary(word string) bool {
	if s == nil || s.SpellerFSA == nil {
		return false
	}
	return s.SpellerFSA.IsInDictionary(word)
}

// IsMisspelled ports Speller.isMisspelled using sticky isInDictionary.
func (s *Speller) IsMisspelled(word string) bool {
	if s == nil || s.Dict == nil || word == "" {
		return false
	}
	d := s.Dict
	wordToCheck := applyConversionPairs(word, d.InputConversion)
	if wordToCheck == "" {
		return false
	}
	r := []rune(wordToCheck)
	isAlpha := len(r) != 1 || isAlphabeticRune(r[0])
	if d.IgnorePunctuation && !isAlpha {
		return false
	}
	if d.IgnoreNumbers && containsDigitRunes(wordToCheck) {
		return false
	}
	if d.IgnoreCamelCase && isCamelCase(wordToCheck) {
		return false
	}
	if d.IgnoreAllUppercase && isAlpha && isAllUppercase(wordToCheck) {
		return false
	}
	if s.IsInDictionary(wordToCheck) {
		return false
	}
	if d.ConvertCase && !isMixedCase(wordToCheck) {
		low := d.ToLower(wordToCheck)
		if s.IsInDictionary(low) {
			return false
		}
		if isAllUppercase(wordToCheck) {
			iu := d.initialUppercase(wordToCheck)
			if iu != wordToCheck && s.IsInDictionary(iu) {
				return false
			}
		}
	}
	return true
}

// ConvertsCase ports Speller.convertsCase().
func (s *Speller) ConvertsCase() bool {
	return s != nil && s.Dict != nil && s.Dict.ConvertCase
}

// LoadDictReplacementPairs wires Dictionary .info replacement pairs into findRepl maps.
func (s *Speller) LoadDictReplacementPairs() {
	if s == nil || s.Dict == nil || len(s.Dict.ReplacementShort) == 0 {
		return
	}
	pairs := make([]struct{ From, To string }, len(s.Dict.ReplacementShort))
	for i, p := range s.Dict.ReplacementShort {
		pairs[i].From, pairs[i].To = p.From, p.To
	}
	s.LoadReplacementPairs(pairs)
}

// SyncFromDict copies Speller-relevant Dictionary flags onto the FSA walker.
func (s *Speller) SyncFromDict() {
	if s == nil || s.Dict == nil {
		return
	}
	d := s.Dict
	s.IgnoreDiacritics = d.IgnoreDiacritics
	s.ConvertCase = d.ConvertCase
	s.EquivalentChars = d.EquivalentChars
	s.LoadDictReplacementPairs()
}
