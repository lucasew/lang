package languagetool

// FragmentWithLanguage ports org.languagetool.FragmentWithLanguage.
type FragmentWithLanguage struct {
	LangCode string
	Fragment string
}

// NewFragmentWithLanguage ports FragmentWithLanguage(String, String).
// Java only rejects null (Objects.requireNonNull); empty strings are allowed.
func NewFragmentWithLanguage(langCode, fragment string) FragmentWithLanguage {
	return FragmentWithLanguage{LangCode: langCode, Fragment: fragment}
}

func (f FragmentWithLanguage) GetLangCode() string { return f.LangCode }
func (f FragmentWithLanguage) GetFragment() string { return f.Fragment }

func (f FragmentWithLanguage) String() string {
	return "| " + f.LangCode + ": " + f.Fragment + " |"
}
