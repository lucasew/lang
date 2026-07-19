package rules

// CreativeWritingCategory ports Java Category(CREATIVE_WRITING, messages
// "category_creative_writing", Location.INTERNAL, false) used by
// AbstractStatisticStyleRule, AbstractStatisticSentenceStyleRule,
// AbstractStyleTooOftenUsedWordRule, AbstractFillerWordsRule, and DE style rules.
func CreativeWritingCategory(messages map[string]string) *Category {
	name := "Stylistic hints for creative writing"
	if messages != nil {
		if s := messages["category_creative_writing"]; s != "" {
			name = s
		}
	}
	// onByDefault=false → category DefaultOff (Java Location.INTERNAL, false)
	return NewCategoryFull(NewCategoryId("CREATIVE_WRITING"), name, CategoryInternal, false, "")
}
