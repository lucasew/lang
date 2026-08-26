package ga

// DativePluralsEntry ports org.languagetool.rules.ga.DativePluralsEntry.
type DativePluralsEntry struct {
	Form        string
	FormModern  string
	Lemma       string
	LemmaModern string
	Equivalent  string
	Replacement string
	Gender      string
}

func NewDativePluralsEntry(form, lemma, gender, replacement string) *DativePluralsEntry {
	return &DativePluralsEntry{
		Form: form, Lemma: lemma, Gender: gender, Replacement: replacement,
	}
}

func (e *DativePluralsEntry) GetForm() string        { return e.Form }
func (e *DativePluralsEntry) GetModern() string      { return e.FormModern }
func (e *DativePluralsEntry) GetLemma() string       { return e.Lemma }
func (e *DativePluralsEntry) GetLemmaModern() string { return e.LemmaModern }
func (e *DativePluralsEntry) GetEquivalent() string  { return e.Equivalent }
func (e *DativePluralsEntry) GetReplacement() string { return e.Replacement }
func (e *DativePluralsEntry) GetGender() string      { return e.Gender }

func (e *DativePluralsEntry) SetEquivalent(s string)  { e.Equivalent = s }
func (e *DativePluralsEntry) SetModernised(s string)  { e.FormModern = s }
func (e *DativePluralsEntry) SetModernLemma(s string) { e.LemmaModern = s }

func (e *DativePluralsEntry) HasEquivalent() bool {
	return e != nil && e.Equivalent != ""
}
func (e *DativePluralsEntry) HasModernised() bool {
	return e != nil && e.FormModern != ""
}
func (e *DativePluralsEntry) HasModernLemma() bool {
	return e != nil && e.LemmaModern != ""
}

func (e *DativePluralsEntry) GetBaseTag() string {
	if e.Gender == "f" {
		return "Noun:Fem:Dat:Pl"
	}
	return "Noun:Masc:Dat:Pl"
}

// GetStandard returns the modern equivalent if present, else the replacement.
func (e *DativePluralsEntry) GetStandard() string {
	if e.HasEquivalent() {
		return e.Equivalent
	}
	return e.Replacement
}
