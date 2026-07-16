package languagetool

// DynamicLanguage ports org.languagetool.DynamicLanguage metadata surface
// (dict-backed language without full Language hierarchy).
type DynamicLanguage struct {
	Name     string
	Code     string // may include variant, e.g. en-US
	DictPath string
}

func NewDynamicLanguage(name, code, dictPath string) DynamicLanguage {
	if name == "" || code == "" || dictPath == "" {
		panic("name, code, and dictPath required")
	}
	return DynamicLanguage{Name: name, Code: code, DictPath: dictPath}
}

// GetShortCode strips -variant suffix.
func (d DynamicLanguage) GetShortCode() string {
	if i := indexDash(d.Code); i >= 0 {
		return d.Code[:i]
	}
	return d.Code
}

func (d DynamicLanguage) GetName() string { return d.Name }

func (d DynamicLanguage) GetShortCodeWithCountryAndVariant() string { return d.Code }

func indexDash(s string) int {
	for i, r := range s {
		if r == '-' {
			return i
		}
	}
	return -1
}
