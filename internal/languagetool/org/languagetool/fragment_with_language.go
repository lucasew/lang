package languagetool

// FragmentWithLanguage ports org.languagetool.FragmentWithLanguage.
type FragmentWithLanguage struct {
	LangCode string
	Fragment string
}

func NewFragmentWithLanguage(langCode, fragment string) FragmentWithLanguage {
	if langCode == "" || fragment == "" {
		panic("langCode and fragment required")
	}
	return FragmentWithLanguage{LangCode: langCode, Fragment: fragment}
}

func (f FragmentWithLanguage) GetLangCode() string { return f.LangCode }
func (f FragmentWithLanguage) GetFragment() string { return f.Fragment }

func (f FragmentWithLanguage) String() string {
	return "| " + f.LangCode + ": " + f.Fragment + " |"
}
