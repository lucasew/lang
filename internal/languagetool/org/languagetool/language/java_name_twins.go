package language

// Contributors is the Java-name twin for shared contributor constants.
type Contributors struct{}

// Known returns well-known multi-language contributors.
func (Contributors) Known() []Contributor {
	return []Contributor{DanielNaber, MarcinMilkowski, DominiquePelle}
}

// LanguageBuilder is the Java-name twin for MakeAdditionalLanguage.
type LanguageBuilder struct{}

// MakeAdditionalLanguage parses rules-<code>-<Name>.xml filenames.
func (LanguageBuilder) MakeAdditionalLanguage(filename string) (LanguageMetaFromFile, error) {
	return MakeAdditionalLanguage(filename)
}
