package languagetool

// Tag ports org.languagetool.Tag (style category tags used in XML; all-lowercase).
type Tag string

const (
	TagPicky        Tag = "picky"
	TagAcademic     Tag = "academic"
	TagClarity      Tag = "clarity"
	TagProfessional Tag = "professional"
	TagCreative     Tag = "creative"
	TagCustomer     Tag = "customer"
	TagJobApp       Tag = "jobapp"
	TagObjective    Tag = "objective"
	TagElegant      Tag = "elegant"
)

// AllTags lists every org.languagetool.Tag value.
func AllTags() []Tag {
	return []Tag{
		TagPicky, TagAcademic, TagClarity, TagProfessional, TagCreative,
		TagCustomer, TagJobApp, TagObjective, TagElegant,
	}
}
