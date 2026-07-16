package language

// Contributor ports org.languagetool.language.Contributor.
type Contributor struct {
	Name string
	URL  string // optional
}

func NewContributor(name string) Contributor {
	return NewContributorWithURL(name, "")
}

func NewContributorWithURL(name, url string) Contributor {
	if name == "" {
		panic("name cannot be null")
	}
	return Contributor{Name: name, URL: url}
}

func (c Contributor) GetName() string { return c.Name }
func (c Contributor) GetURL() string  { return c.URL }
func (c Contributor) String() string  { return c.Name }
