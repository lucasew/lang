package pt

// Common Portuguese contractions → synthetic multi-lemma POS (green MapWordTagger aid).
// Format: surface → list of "lemma|POS" (simplified).
var portugueseContractions = map[string][]struct {
	Lemma, POS string
}{
	"do":   {{"de", "SPS00"}, {"o", "DA0MS0"}},
	"da":   {{"de", "SPS00"}, {"a", "DA0FS0"}},
	"dos":  {{"de", "SPS00"}, {"os", "DA0MP0"}},
	"das":  {{"de", "SPS00"}, {"as", "DA0FP0"}},
	"no":   {{"em", "SPS00"}, {"o", "DA0MS0"}},
	"na":   {{"em", "SPS00"}, {"a", "DA0FS0"}},
	"nos":  {{"em", "SPS00"}, {"os", "DA0MP0"}},
	"nas":  {{"em", "SPS00"}, {"as", "DA0FP0"}},
	"ao":   {{"a", "SPS00"}, {"o", "DA0MS0"}},
	"à":    {{"a", "SPS00"}, {"a", "DA0FS0"}},
	"pelo": {{"por", "SPS00"}, {"o", "DA0MS0"}},
	"pela": {{"por", "SPS00"}, {"a", "DA0FS0"}},
}

// ContractionReadings returns synthetic readings for known contractions.
func ContractionReadings(surface string) []struct{ Lemma, POS string } {
	low := surface
	// lower ASCII; keep accented
	b := []rune(surface)
	for i, r := range b {
		if r >= 'A' && r <= 'Z' {
			b[i] = r + ('a' - 'A')
		}
	}
	low = string(b)
	if v, ok := portugueseContractions[low]; ok {
		return v
	}
	return nil
}
