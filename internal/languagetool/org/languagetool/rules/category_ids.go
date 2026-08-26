package rules

// CategoryIdsType is the struct type of the CategoryIds value.
type CategoryIdsType = struct {
	Typography, Casing, Grammar, Typos, Punctuation,
	ConfusedWords, Redundancy, Style, GenderNeutrality,
	Semantics, Colloquialisms, Wikipedia, Barbarism, Misc CategoryId
}

// CategoryIds ports org.languagetool.rules.CategoryIds as named accessors.
// Prefer the Category* package variables for direct use.
var CategoryIds = struct {
	Typography, Casing, Grammar, Typos, Punctuation,
	ConfusedWords, Redundancy, Style, GenderNeutrality,
	Semantics, Colloquialisms, Wikipedia, Barbarism, Misc CategoryId
}{
	Typography:       CategoryTypography,
	Casing:           CategoryCasing,
	Grammar:          CategoryGrammar,
	Typos:            CategoryTypos,
	Punctuation:      CategoryPunctuation,
	ConfusedWords:    CategoryConfusedWords,
	Redundancy:       CategoryRedundancy,
	Style:            CategoryStyle,
	GenderNeutrality: CategoryGenderNeutrality,
	Semantics:        CategorySemantics,
	Colloquialisms:   CategoryColloquialisms,
	Wikipedia:        CategoryWikipedia,
	Barbarism:        CategoryBarbarism,
	Misc:             CategoryMisc,
}
