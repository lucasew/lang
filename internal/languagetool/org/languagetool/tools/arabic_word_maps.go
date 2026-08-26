package tools

// IsolatedToAttachedPronoun ports ArabicConstantsMaps.isolatedToAttachedPronoun.
var IsolatedToAttachedPronoun = map[string]string{
	"أنا":   "ني",
	"نحن":   "نا",
	"هو":    "ه",
	"هي":    "ها",
	"هم":    "هم",
	"هن":    "هن",
	"أنتما": "كما",
	"أنتم":  "كم",
	"أنتن":  "كن",
}

// GetAttachedPronoun ports ArabicWordMaps.getAttachedPronoun.
func GetAttachedPronoun(word string) string {
	if word == "" {
		return ""
	}
	if v, ok := IsolatedToAttachedPronoun[word]; ok {
		return v
	}
	return ""
}
