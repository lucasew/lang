package morfologik

import (
	"unicode"
	"unicode/utf16"
)

// ReplaceRunOnWordCandidates ports Speller.replaceRunOnWordCandidates (2.2.0).
// Cold Dictionary path (fresh Speller each call). Prefer Speller.ReplaceRunOnWordCandidates
// for sticky containsSeparators.
func (d *Dictionary) ReplaceRunOnWordCandidates(original string) []CandidateData {
	if d == nil {
		return nil
	}
	return NewSpeller(d, 1).ReplaceRunOnWordCandidates(original)
}

// ReplaceRunOnWordCandidates ports Speller.replaceRunOnWordCandidates on this Speller
// (sticky isInDictionary / containsSeparators).
func (s *Speller) ReplaceRunOnWordCandidates(original string) []CandidateData {
	if s == nil || s.Dict == nil || original == "" || !s.Dict.SupportRunOnWords {
		return nil
	}
	d := s.Dict
	wordToCheck := applyConversionPairs(original, d.InputConversion)
	if wordToCheck == "" || s.IsInDictionary(wordToCheck) {
		return nil
	}
	// Java: for (i = 1; i < wordToCheck.length(); i++) with substring UTF-16
	u := utf16.Encode([]rune(wordToCheck))
	if len(u) < 2 {
		return nil
	}
	var candidates []CandidateData
	for i := 1; i < len(u); i++ {
		prefix := string(utf16.Decode(u[:i]))
		suffix := string(utf16.Decode(u[i:]))
		// suffix: in dict OR capitalized with lowercase form in dict (GreatElephant)
		suffixOK := s.IsInDictionary(suffix)
		if !suffixOK && !isNotCapitalizedWord(suffix) {
			suffixOK = s.IsInDictionary(d.ToLower(suffix))
		}
		if !suffixOK {
			continue
		}
		if s.IsInDictionary(prefix) {
			candidates = append(candidates, d.addRunOnReplacement(prefix+" "+suffix))
		} else if prefix != "" {
			pr := []rune(prefix)
			// Java: Character.isUpperCase(prefix.charAt(0))
			if len(pr) > 0 && unicode.IsUpper(pr[0]) && s.IsInDictionary(d.ToLower(prefix)) {
				candidates = append(candidates, d.addRunOnReplacement(prefix+" "+suffix))
			}
		}
	}
	return candidates
}

// ReplaceRunOnWords ports Speller.replaceRunOnWords (strings only).
func (d *Dictionary) ReplaceRunOnWords(original string) []string {
	cds := d.ReplaceRunOnWordCandidates(original)
	if len(cds) == 0 {
		return nil
	}
	out := make([]string, 0, len(cds))
	for _, c := range cds {
		out = append(out, c.Word)
	}
	return out
}

// addRunOnReplacement ports Speller.addReplacement (output conversion + CandidateData dist 1).
func (d *Dictionary) addRunOnReplacement(replacement string) CandidateData {
	if len(d.OutputConversion) > 0 {
		replacement = applyConversionPairs(replacement, d.OutputConversion)
	}
	// CandidateData(replacement, 1) → weighted by frequency of replacement string
	freq := 0
	if d != nil {
		// Java getFrequency on full "a b" string — usually 0 for space forms
		freq = d.GetFrequency(replacement)
		if freq < 0 {
			freq = 0
		}
	}
	orig := 1
	dist := orig*FreqRanges + FreqRanges - freq - 1
	return CandidateData{Word: replacement, OrigDistance: orig, Distance: dist}
}
