package morfologik

// WeightedSuggestion ports org.languagetool.rules.spelling.morfologik.WeightedSuggestion.
type WeightedSuggestion struct {
	Word   string
	Weight int
}

func NewWeightedSuggestion(word string, weight int) WeightedSuggestion {
	if word == "" {
		// Java requireNonNull only
	}
	return WeightedSuggestion{Word: word, Weight: weight}
}

func (w WeightedSuggestion) GetWord() string { return w.Word }
func (w WeightedSuggestion) GetWeight() int  { return w.Weight }

func (w WeightedSuggestion) Less(o WeightedSuggestion) bool {
	return w.Weight < o.Weight
}

func (w WeightedSuggestion) String() string {
	return w.Word + "/" + itoa(w.Weight)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// SortByWeight sorts ascending by weight (stable not required).
func SortByWeight(s []WeightedSuggestion) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[j].Weight < s[i].Weight {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}
