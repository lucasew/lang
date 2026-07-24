package ga

// Utils is the Java-name twin for package-level Irish mutation helpers.
type Utils struct{}

func (Utils) Lenite(in string) string   { return Lenite(in) }
func (Utils) Eclipse(in string) string  { return Eclipse(in) }
func (Utils) IsVowel(c rune) bool       { return IsVowel(c) }
func (Utils) FixSuffix(in string) *Retaggable { return FixSuffix(in) }
func (Utils) Demutate(in string) *Retaggable  { return Demutate(in) }
func (Utils) MorphWord(in string) []*Retaggable { return MorphWord(in) }
func (Utils) UnLenite(in string) string { return UnLenite(in) }
