package rules

// CategoryDef ports org.languagetool.rules.Categories entries (id + message key).
type CategoryDef struct {
	ID         string
	MessageKey string
}

func (c CategoryDef) GetID() CategoryId {
	return NewCategoryId(c.ID)
}

// GetCategory builds a Category using a message lookup for MessageKey.
// When messages is nil or missing the key, the id is used as the display name.
func (c CategoryDef) GetCategory(messages map[string]string) *Category {
	name := c.ID
	if messages != nil {
		if s, ok := messages[c.MessageKey]; ok && s != "" {
			name = s
		}
	}
	return NewCategory(NewCategoryId(c.ID), name)
}

// Pre-defined Categories (Categories.java).
var (
	CatCasing           = CategoryDef{"CASING", "category_case"}
	CatCompounding      = CategoryDef{"COMPOUNDING", "category_compounding"}
	CatGrammar          = CategoryDef{"GRAMMAR", "category_grammar"}
	CatTypos            = CategoryDef{"TYPOS", "category_typo"}
	CatPunctuation      = CategoryDef{"PUNCTUATION", "category_punctuation"}
	CatTypography       = CategoryDef{"TYPOGRAPHY", "category_typography"}
	CatConfusedWords    = CategoryDef{"CONFUSED_WORDS", "category_confused_words"}
	CatRepetitions      = CategoryDef{"REPETITIONS", "category_repetitions"}
	CatRedundancy       = CategoryDef{"REDUNDANCY", "category_redundancy"}
	CatRepetitionsStyle = CategoryDef{"REPETITIONS_STYLE", "cateogry_repetitions_style"} // typo kept from Java
	CatStyle            = CategoryDef{"STYLE", "category_style"}
	CatPlainEnglish     = CategoryDef{"PLAIN_ENGLISH", "category_plain_english"}
	CatGenderNeutrality = CategoryDef{"GENDER_NEUTRALITY", "category_gender_neutrality"}
	CatSemantics        = CategoryDef{"SEMANTICS", "category_semantics"}
	CatColloquialisms   = CategoryDef{"COLLOQUIALISMS", "category_colloquialism"}
	CatRegionalisms     = CategoryDef{"REGIONALISMS", "category_regionalisms"}
	CatFalseFriends     = CategoryDef{"FALSE_FRIENDS", "category_false_friend"}
	CatWikipedia        = CategoryDef{"WIKIPEDIA", "category_wikipedia"}
	CatMisc             = CategoryDef{"MISC", "category_misc"}
)

// AllCategories is Categories.ALL order from Java.
var AllCategories = []CategoryDef{
	CatStyle, CatRepetitionsStyle, CatRepetitions, CatCasing, CatCompounding,
	CatColloquialisms, CatConfusedWords, CatFalseFriends, CatGenderNeutrality,
	CatGrammar, CatMisc, CatPlainEnglish, CatRedundancy, CatRegionalisms,
	CatPunctuation, CatTypography, CatWikipedia, CatTypos,
}

// Categories is the Java-name twin grouping CategoryDef constants.
// Prefer Cat* variables (CatGrammar, CatTypos, …).
type Categories = CategoryDef
