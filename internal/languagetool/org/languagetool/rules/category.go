package rules

// CategoryLocation ports Category.Location.
type CategoryLocation int

const (
	// CategoryInternal is part of the main LT distribution.
	CategoryInternal CategoryLocation = iota
	// CategoryExternal is not part of the main distribution.
	CategoryExternal
)

// CategoryId ports org.languagetool.rules.CategoryId.
type CategoryId struct {
	id string
}

// NewCategoryId creates a non-empty category identifier.
func NewCategoryId(id string) CategoryId {
	if id == "" {
		panic("Category id must not be null/empty")
	}
	empty := true
	for _, r := range id {
		if r != ' ' && r != '\t' && r != '\n' {
			empty = false
			break
		}
	}
	if empty {
		panic("Category id must not be empty")
	}
	return CategoryId{id: id}
}

func (c CategoryId) String() string           { return c.id }
func (c CategoryId) Equals(o CategoryId) bool { return c.id == o.id }

// Category ports org.languagetool.rules.Category.
type Category struct {
	ID         CategoryId
	Name       string
	Location   CategoryLocation
	DefaultOff bool
	TabName    string // optional UI tab; empty = none
}

// NewCategory builds an internal category that is on by default.
func NewCategory(id CategoryId, name string) *Category {
	return NewCategoryFull(id, name, CategoryInternal, true, "")
}

// NewCategoryFull builds a category with full options (onByDefault like Java).
func NewCategoryFull(id CategoryId, name string, loc CategoryLocation, onByDefault bool, tabName string) *Category {
	if name == "" {
		panic("Category name must not be empty")
	}
	return &Category{
		ID:         id,
		Name:       name,
		Location:   loc,
		DefaultOff: !onByDefault,
		TabName:    tabName,
	}
}

func (c *Category) GetID() CategoryId             { return c.ID }
func (c *Category) GetName() string               { return c.Name }
func (c *Category) GetTabName() string            { return c.TabName }
func (c *Category) IsDefaultOff() bool            { return c.DefaultOff }
func (c *Category) GetLocation() CategoryLocation { return c.Location }
func (c *Category) String() string                { return c.Name }

// Standard category ids (CategoryIds.java).
var (
	CategoryTypography       = NewCategoryId("TYPOGRAPHY")
	CategoryCasing           = NewCategoryId("CASING")
	CategoryGrammar          = NewCategoryId("GRAMMAR")
	CategoryTypos            = NewCategoryId("TYPOS")
	CategoryPunctuation      = NewCategoryId("PUNCTUATION")
	CategoryConfusedWords    = NewCategoryId("CONFUSED_WORDS")
	CategoryRedundancy       = NewCategoryId("REDUNDANCY")
	CategoryStyle            = NewCategoryId("STYLE")
	CategoryGenderNeutrality = NewCategoryId("GENDER_NEUTRALITY")
	CategorySemantics        = NewCategoryId("SEMANTICS")
	CategoryColloquialisms   = NewCategoryId("COLLOQUIALISMS")
	CategoryWikipedia        = NewCategoryId("WIKIPEDIA")
	CategoryBarbarism        = NewCategoryId("BARBARISM")
	CategoryMisc             = NewCategoryId("MISC")
)
