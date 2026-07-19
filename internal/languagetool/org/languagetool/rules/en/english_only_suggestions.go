package en

import (
	"regexp"
	"strings"
)

// english_only_suggestions ports AbstractEnglishSpellerRule.getOnlySuggestions.
// First matching Java if-arm wins. replaceOnce is case-sensitive substring replace.

type onlySugRule struct {
	re     *regexp.Regexp
	fixed  string   // full replacement when set (single)
	fixeds []string // multi-suggestion only-list when set
	old    string   // substring for replaceOnce
	neu    string
}

var englishOnlySuggestionRules []onlySugRule

func init() {
	// Java replaceOnce arms (pattern full match, then replaceOnce(old, new) on word)
	type rep struct{ pat, old, neu string }
	reps := []rep{
		{`^[Pp]rofileration$`, "rofileration", "roliferation"},
		{`^[Cc]emetary$`, "emetary", "emetery"},
		{`^[Cc]emetaries$`, "emetaries", "emeteries"},
		{`^[Bb]asicly$`, "asicly", "asically"},
		{`^[Bb]eleives?$`, "eleive", "elieve"},
		{`^[Bb]elives?$`, "elive", "elieve"},
		{`^[Bb]izzare$`, "izzare", "izarre"},
		{`^[Cc]ompletly$`, "ompletly", "ompletely"},
		{`^[Dd]issapears?$`, "issapear", "isappear"},
		{`^[Ff]arenheit$`, "arenheit", "ahrenheit"},
		{`^[Ff]reinds?$`, "reind", "riend"},
		{`^[Ii]ncidently$`, "ncidently", "ncidentally"},
		{`^[Ii]nterupts?$`, "nterupt", "nterrupt"},
		{`^[Ll]ollypops?$`, "ollypop", "ollipop"},
		{`^[Oo]cassions?$`, "cassion", "ccasion"},
		{`^[Oo]ccurances?$`, "ccurance", "ccurrence"},
		{`^[Pp]ersistant$`, "ersistant", "ersistent"},
		{`^[Pp]eices?$`, "eice", "iece"},
		{`^[Ss]eiges?$`, "eige", "iege"},
		{`^[Ss]upercedes?$`, "upercede", "upersede"},
		{`^[Tt]hreshholds?$`, "hreshhold", "hreshold"},
		{`^[Tt]ommorrows?$`, "ommorrow", "omorrow"},
		{`^[Tt]ounges?$`, "ounge", "ongue"},
		{`^[Ww]ierd$`, "ierd", "eird"},
		{`^[Ss]argent$`, "argent", "ergeant"},
		{`^[Aa]dmittingly$`, "dmittingly", "dmittedly"},
		// Java: intransparent(ly)? → replaceOnce "in" → "un" (first "in")
		{`^intransparent(ly)?$`, "in", "un"},
		{`^[Bb]onafide$`, "onafide", "ona fide"},
		{`^[Aa]llright$`, "llright", "lright"},
		{`^[Aa]ddons?$`, "ddon", "dd-on"},
		{`^[Ww]hereever$`, "hereever", "herever"},
		{`^[Uu]ninspirational$`, "ninspirational", "ninspiring"},
		{`^[Mm]acbooks?$`, "acbook", "acBook"},
		{`^[Ll]ikelyhood$`, "ikelyhood", "ikelihood"},
		{`^[Uu]necessary$`, "necessary", "nnecessary"},
		{`^[Ff]orseeable$`, "orseeable", "oreseeable"},
		{`^[Uu]nforseeable$`, "orseeable", "oreseeable"},
		{`^[Ff]orseeably$`, "orseeably", "oreseeably"},
		{`^[Uu]nforseeably$`, "orseeably", "oreseeably"},
	}
	for _, x := range reps {
		englishOnlySuggestionRules = append(englishOnlySuggestionRules, onlySugRule{
			re: regexp.MustCompile(x.pat), old: x.old, neu: x.neu,
		})
	}
	// Java fixed topMatch("…") arms
	type fix struct{ pat, s string }
	fixes := []fix{
		{`^swimmed$`, "swam"},
		{`^misspelt$`, "misspelled"},
		{`^[Ad]hoc$`, "ad hoc"},
		// Java DEACTIVE = [De]eactive
		{`^[De]eactive$`, "inactive"},
		{`^[Hh]ubspot$`, "HubSpot"},
		{`^[Uu]rl$`, "URL"},
		{`^[Hh]ttp$`, "HTTP"},
		{`^[Hh]ttps$`, "HTTPS"},
		{`^[Ff]yi$`, "FYI"},
		{`^european$`, "European"},
		{`^europeans$`, "Europeans"},
		{`^[Dd]evops$`, "DevOps"},
		{`^microsoft$`, "Microsoft"},
		{`^[Ll]anguagetool$`, "LanguageTool"},
		{`^[hH]on[kg]kong$`, "Hong Kong"},
		{`^october$`, "October"},
		{`^september$`, "September"},
		{`^december$`, "December"},
		{`^november$`, "November"},
		{`^april$`, "April"},
		{`^afaik$`, "AFAIK"},
		{`^january$`, "January"},
		{`^english$`, "English"},
		{`^spanish$`, "Spanish"},
		{`^undeterministic$`, "nondeterministic"},
		{`^[Ww]dyt$`, "WDYT"},
		{`^[UuIi]ncompliant$`, "non-compliant"},
		{`^ux$`, "UX"},
		{`^[Gg]itlab$`, "GitLab"},
		{`^[Ww]hatsapp$`, "WhatsApp"},
		{`^jetlagged$`, "jet-lagged"},
		{`^[Qq]uill?bot$`, "QuillBot"},
		{`^QuilBot$`, "QuillBot"},
	}
	for _, x := range fixes {
		englishOnlySuggestionRules = append(englishOnlySuggestionRules, onlySugRule{
			re: regexp.MustCompile(x.pat), fixed: x.s,
		})
	}
	// Java multi-suggestion only arms (QuillBot possessive, TV, jist)
	type multi struct {
		pat  string
		sugs []string
	}
	multis := []multi{
		{`^QuillBots$`, []string{"QuillBot's", "QuillBot"}},
		{`^[Qq]uill?bots$`, []string{"QuillBot's", "QuillBot"}},
		{`^QuilBots$`, []string{"QuillBot's", "QuillBot"}},
		{`^tv$`, []string{"TV", "to"}},
		{`^[Jj]ist$`, []string{"just", "gist"}},
	}
	for _, x := range multis {
		englishOnlySuggestionRules = append(englishOnlySuggestionRules, onlySugRule{
			re: regexp.MustCompile(x.pat), fixeds: x.sugs,
		})
	}
}

// EnglishOnlySuggestions ports getOnlySuggestions; nil/empty when no arm matches.
func EnglishOnlySuggestions(word string) []string {
	if word == "" {
		return nil
	}
	for _, rule := range englishOnlySuggestionRules {
		if !rule.re.MatchString(word) {
			continue
		}
		if len(rule.fixeds) > 0 {
			return append([]string(nil), rule.fixeds...)
		}
		if rule.fixed != "" {
			return []string{rule.fixed}
		}
		// Apache StringUtils.replaceOnce: first exact substring occurrence
		if rule.old != "" {
			if i := strings.Index(word, rule.old); i >= 0 {
				return []string{word[:i] + rule.neu + word[i+len(rule.old):]}
			}
		}
	}
	return nil
}
