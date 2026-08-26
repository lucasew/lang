package language

// Contributor ports org.languagetool.language.Contributor.
type Contributor struct {
	Name string
	// URL is optional homepage; empty string maps Java null.
	URL string
}

// NewContributor ports Contributor(String name) with url = null.
func NewContributor(name string) Contributor {
	return NewContributorWithURL(name, "")
}

// NewContributorWithURL ports Contributor(String name, String url).
// name may be empty (Java Objects.requireNonNull only rejects null).
func NewContributorWithURL(name, url string) Contributor {
	return Contributor{Name: name, URL: url}
}

func (c Contributor) GetName() string { return c.Name }
func (c Contributor) GetURL() string  { return c.URL }
func (c Contributor) String() string  { return c.Name }
