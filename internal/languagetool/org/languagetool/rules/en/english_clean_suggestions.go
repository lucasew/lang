package en

import "strings"

// english_clean_suggestions ports AbstractEnglishSpellerRule.cleanSuggestions (#2562).
// Drops multi-word suggestions that look like bad prefix/suffix splits.

var englishCleanStartsWithLC = []string{
	"re ",
	"en ",
	"co ",
	"cl ",
	"de ",
	"ex ",
	"es ",
	"ab ",
	"ty ",
	"mid ",
	"non ",
	"bio ",
	"bi ",
	"op ",
	"con ",
	"pre ",
	"mis ",
	"socio ",
	"proto ",
	"neo ",
	"geo ",
	"inter ",
	"multi ",
	"retro ",
	"extra ",
	"mega ",
	"meta ",
	"poly ",
	"para ",
	"uni ",
	"anti ",
	"necro ",
	"photo ",
	"post ",
	"sub ",
	"auto ",
	"pl ",
	"ht ",
	"dis ",
	"est ",
	"mono ",
	"trans ",
	"neuro ",
	"hetero ",
	"ultra ",
	"mini ",
	"hyper ",
	"micro ",
	"counter ",
	"over ",
	"overs ",
	"overt ",
	"under ",
	"cyber ",
	"hydro ",
	"ergo ",
	"fore ",
	"pro ",
	"pseudo ",
	"psycho ",
	"mi ",
	"nano ",
	"ans ",
	"semi ",
	"infra ",
	"hypo ",
	"syn ",
	"adv ",
	"com ",
	"res ",
	"resp ",
	"lo ",
	"ed ",
	"ac ",
	"al ",
	"ea ",
	"ge ",
	"mu ",
	"ma ",
	"la ",
	"bis ",
	"ger ",
	"inf ",
	"tar ",
	"f ",
	"k ",
	"l ",
	"b ",
	"e ",
	"c ",
	"d ",
	"p ",
	"v ",
	"h ",
	"r ",
	"s ",
	"t ",
	"u ",
	"w ",
	"um ",
	"oft ",
}

var englishCleanStartsWithCS = []string{
	"i ",
	"sh ",
	"li ",
	"ha ",
	"st ",
	"ins ",
}

var englishCleanEndsWith = []string{
	" i",
	" ING",
	" able",
	" om",
	" ox",
	" ht",
	" wide",
	" less",
	" sly",
	" OO",
	" HHH",
	" ally",
	" ize",
	" sh",
	" st",
	" est",
	" em",
	" ward",
	" ability",
	" ware",
	" logy",
	" ting",
	" ion",
	" ions",
	" cal",
	" ted",
	" sphere",
	" ell",
	" co",
	" con",
	" com",
	" sis",
	" like",
	" full",
	" en",
	" ne",
	" ed",
	" al",
	" ans",
	" mans",
	" ti",
	" de",
	" ea",
	" ge",
	" ab",
	" rs",
	" mi",
	" tar",
	" adv",
	" re",
	" e",
	" c",
	" v",
	" h",
	" s",
	" r",
	" l",
	" u",
	" um",
	" er",
	" es",
	" ex",
	" na",
	" ifs",
	" gs",
	" don",
	" dons",
	" la",
	" ism",
	" ma",
}

// EnglishCleanSuggestions filters lazy suggestion list like Java cleanSuggestions.
func EnglishCleanSuggestions(suggestions []string) []string {
	if len(suggestions) == 0 {
		return suggestions
	}
	out := make([]string, 0, len(suggestions))
	for _, rep := range suggestions {
		if !strings.Contains(rep, " ") {
			out = append(out, rep)
			continue
		}
		if shouldDropSplitSuggestion(rep) {
			continue
		}
		out = append(out, rep)
	}
	return out
}

func shouldDropSplitSuggestion(rep string) bool {
	if !strings.Contains(rep, " ") {
		return false
	}
	repLc := strings.ToLower(rep)
	for _, p := range englishCleanStartsWithLC {
		if strings.HasPrefix(repLc, p) {
			return true
		}
	}
	for _, p := range englishCleanStartsWithCS {
		if strings.HasPrefix(rep, p) {
			return true
		}
	}
	for _, p := range englishCleanEndsWith {
		if strings.HasSuffix(rep, p) {
			return true
		}
	}
	return false
}
