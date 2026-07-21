package de

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Additional top suggestions (GermanSpellerRule.getAdditionalTopSuggestionsString)
// plus exact ADDITIONAL_SUGGESTIONS put() map (see german_speller_additional_exact.go).
// Regexp ADDITIONAL_SUGGESTIONS putRepl/put(pattern) included; sortSuggestionByQuality not complete.

var (
	reEndsIbelkeit = regexp.MustCompile(`(?i).*(?:ibelkeit|iblichkeit)$`)
	reAllmahllig   = regexp.MustCompile(`(?i)^[aA]llmähll?i(?:g|ch)(?:e[mnrs]?)?$`)
	// Java: .*[mM]a[jy]onn?[äe]se.*
	reMayonnaise = regexp.MustCompile(`(?i).*ma[jy]onn?[äe]se.*`)
	// Java: .*[rR]es(a|er)[vw]i[he]?rung(en)?
	reReservierung  = regexp.MustCompile(`(?i).*res(?:a|er)[vw]i[he]?rung(?:en)?.*`)
	reReschaschier  = regexp.MustCompile(`(?i)^reschaschier.+`)
	reLaborants     = regexp.MustCompile(`(?i).*laborants$`)
	reProfessionell = regexp.MustCompile(`(?i)^proff?ess?ion[äe]h?ll?(?:e[mnrs]?)?$`)
	reVerstandnis   = regexp.MustCompile(`(?i)^verstehendniss?(?:es?)?$`)
)

// dictAccepts is true when FilterDict is wired and word is not misspelled.
func dictAccepts(word string) bool {
	if !FilterDictAvailable() || word == "" {
		return false
	}
	return !FilterDictIsMisspelled(word)
}

// AdditionalTopSuggestions ports getAdditionalTopSuggestionsString fixed cases
// and rewrite-with-dict checks. Returns nil if no curated additional list applies.
func (r *GermanSpellerRule) AdditionalTopSuggestions(word string) []string {
	if r == nil || word == "" {
		return nil
	}
	if strings.EqualFold(word, "WIFI") {
		return []string{"Wi-Fi"}
	}
	if strings.EqualFold(word, "W-Lan") {
		return []string{"WLAN"}
	}
	switch word {
	case "Endstadion":
		return []string{"Endstadium"}
	case "Endstadions":
		return []string{"Endstadiums"}
	case "genomen":
		return []string{"genommen"}
	case "Preis-Leistungsverhältnis":
		return []string{"Preis-Leistungs-Verhältnis"}
	case "getz":
		return []string{"jetzt", "geht's"}
	case "Trons":
		return []string{"Trance"}
	case "ei":
		return []string{"ein"}
	case "jo", "jepp", "jopp":
		return []string{"ja"}
	case "Jo", "Jepp", "Jopp":
		return []string{"Ja"}
	case "Ne":
		return []string{"Nein", "Eine"}
	case "is":
		return []string{"ist"}
	case "Is":
		return []string{"Ist"}
	case "un":
		return []string{"und"}
	case "Un":
		return []string{"Und"}
	case "Std":
		return []string{"Std."}
	case "gin":
		return []string{"ging"}
	case "dh", "dh.":
		return []string{"d. h."}
	case "ua", "ua.":
		return []string{"u. a."}
	case "uvm", "uvm.":
		return []string{"u. v. m."}
	case "udgl", "udgl.":
		return []string{"u. dgl."}
	case "Ruhigkeit":
		return []string{"Ruhe"}
	case "angepreist":
		return []string{"angepriesen"}
	case "halo":
		return []string{"hallo"}
	case "ca":
		return []string{"ca."}
	case "Jezt":
		return []string{"Jetzt"}
	case "Wollst":
		return []string{"Wolltest"}
	case "wollst":
		return []string{"wolltest"}
	case "Rolladen":
		return []string{"Rollladen"}
	case "Maßname":
		return []string{"Maßnahme"}
	case "Maßnamen":
		return []string{"Maßnahmen"}
	case "nanten":
		return []string{"nannten"}
	case "diees":
		return []string{"dieses", "dies"}
	case "Diees":
		return []string{"Dieses", "Dies"}
	case "Lobbies":
		return []string{"Lobbys"}
	case "Parties":
		return []string{"Partys"}
	case "Babies":
		return []string{"Babys"}
	case "Hallochen":
		return []string{"Hallöchen", "hallöchen"}
	case "hallochen":
		return []string{"hallöchen"}
	case "ok":
		return []string{"okay", "O. K."}
	case "gesuchen":
		return []string{"gesuchten", "gesucht"}
	case "Germanistiker":
		return []string{"Germanist", "Germanisten"}
	case "Abschlepper":
		return []string{"Abschleppdienst", "Abschleppwagen"}
	case "par":
		return []string{"paar"}
	case "iwie":
		return []string{"irgendwie"}
	case "schwarzfarbenden":
		return []string{"schwarzfarbenen"}
	case "bzgl":
		return []string{"bzgl."}
	case "bau":
		return []string{"baue"}
	case "sry":
		return []string{"sorry"}
	case "Sry":
		return []string{"Sorry"}
	case "thx":
		return []string{"danke"}
	case "Thx":
		return []string{"Danke"}
	case "Zynik":
		return []string{"Zynismus"}
	case "pieksen":
		return []string{"piksen"}
	case "piekst":
		return []string{"pikst"}
	case "gepiekst":
		return []string{"gepikst"}
	case "wiederspiegeln":
		return []string{"widerspiegeln"}
	case "ch":
		return []string{"ich"}
	}
	// Java: equalsIgnoreCase("email")
	if strings.EqualFold(word, "email") {
		return []string{"E-Mail"}
	}
	// Java: word.length() > 9 && startsWith("Email") → E-Mail- + suggest(suffix)
	if utf16LenDE(word) > 9 && strings.HasPrefix(word, "Email") {
		suffix := substringByUTF16(word, 5, utf16LenDE(word)) // "Email".length()==5
		if !dictAccepts(suffix) {
			// hunspell.suggest(uppercaseFirstChar(suffix)) — first suggestion if any
			up := tools.UppercaseFirstChar(suffix)
			if sugs := FilterDictSuggest(up); len(sugs) > 0 {
				suffix = sugs[0]
			} else if sugs := FilterDictSuggest(suffix); len(sugs) > 0 {
				suffix = sugs[0]
			}
		}
		if utf16LenDE(suffix) == 0 {
			return nil
		}
		// Java: "E-Mail-"+Character.toUpperCase(suffix.charAt(0))+suffix.substring(1)
		first := javaCharAtDE(suffix, 0)
		rest := substringByUTF16(suffix, 1, utf16LenDE(suffix))
		return []string{"E-Mail-" + string(unicode.ToUpper(first)) + rest}
	}
	if strings.EqualFold(word, "zumindestens") {
		return []string{strings.Replace(word, "ens", "", 1)}
	}
	// zb / zB abbreviations
	if word == "zb" || word == "zB" || word == "zb." || word == "zB." {
		return []string{"z. B."}
	}

	// rewrite patterns — only if dict accepts the suggestion
	if reEndsIbelkeit.MatchString(word) {
		sug := regexp.MustCompile(`(?i)el[hk]eit$`).ReplaceAllString(word, "ilität")
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if strings.HasSuffix(strings.ToLower(word), "aquise") {
		// replaceFirst aquise$ → akquise
		low := word
		idx := strings.LastIndex(strings.ToLower(low), "aquise")
		if idx >= 0 {
			sug := word[:idx] + "akquise"
			if dictAccepts(sug) {
				return []string{sug}
			}
		}
	}
	if strings.HasSuffix(word, "standart") {
		sug := strings.TrimSuffix(word, "standart") + "standard"
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if strings.HasSuffix(word, "standarts") {
		sug := strings.TrimSuffix(word, "standarts") + "standards"
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if strings.HasSuffix(word, "tips") {
		sug := strings.TrimSuffix(word, "tips") + "tipps"
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if strings.HasSuffix(word, "tip") {
		sug := word + "p"
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if strings.Contains(word, "entfehlung") {
		sug := strings.Replace(word, "ent", "emp", 1)
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if strings.HasSuffix(word, "oullie") {
		sug := strings.TrimSuffix(word, "oullie") + "ouille"
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if strings.HasPrefix(word, "Bundstift") {
		sug := "Buntstift" + strings.TrimPrefix(word, "Bundstift")
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if reAllmahllig.MatchString(word) {
		sug := regexp.MustCompile(`llmähll?i(g|ch)`).ReplaceAllString(word, "llmählich")
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if reMayonnaise.MatchString(word) {
		sug := regexp.MustCompile(`(?i)a[jy]onn?[äe]se`).ReplaceAllString(word, "ayonnaise")
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if reReservierung.MatchString(word) {
		sug := regexp.MustCompile(`(?i)es(a|er)[vw]i[he]?rung`).ReplaceAllString(word, "eservierung")
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if reReschaschier.MatchString(word) {
		sug := strings.Replace(word, "schaschier", "cherchier", 1)
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if reLaborants.MatchString(word) {
		sug := regexp.MustCompile(`(?i)ts$`).ReplaceAllString(word, "ten")
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if reProfessionell.MatchString(word) {
		sug := regexp.MustCompile(`(?i)roff?ess?ion([äe])h?l{1,2}`).ReplaceAllString(word, "rofessionell")
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if reVerstandnis.MatchString(word) {
		sug := regexp.MustCompile(`(?i)[vV]erstehendnis`).ReplaceAllString(word, "Verständnis")
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if strings.HasPrefix(word, "koregier") {
		sug := strings.Replace(word, "reg", "rrig", 1)
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if re := regexp.MustCompile(`(?i)^diagno[sz]ier`); re.MatchString(word) {
		sug := regexp.MustCompile(`(?i)gno[sz]ier`).ReplaceAllString(word, "gnostizier")
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if strings.Contains(word, "eiss") {
		sug := strings.Replace(word, "eiss", "eiß", 1)
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if strings.Contains(word, "Akkupressur") {
		sug := strings.Replace(word, "Akkupressur", "Akupressur", 1)
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if strings.Contains(word, "farbend") {
		sug := strings.Replace(word, "farbend", "farben", 1)
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if strings.Contains(word, "uess") {
		sug := strings.Replace(word, "uess", "üß", 1)
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if strings.HasSuffix(word, "derbies") {
		sug := strings.TrimSuffix(word, "derbies") + "derbys"
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if strings.HasSuffix(word, "stories") {
		sug := strings.TrimSuffix(word, "stories") + "storys"
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	if strings.HasSuffix(word, "parties") {
		sug := strings.TrimSuffix(word, "parties") + "partys"
		if dictAccepts(sug) {
			return []string{sug}
		}
	}
	return nil
}

// Suggest ports getSuggestions entry order:
// 1) getAdditionalTopSuggestionsString fixed/rewrite cases
// 2) ADDITIONAL_SUGGESTIONS exact put() map
// 3) putRepl / regex put / multi / replace-lambdas
// 4) uppercase-if-dict, ABK abbreviation, hyphenated compound parts
// 5) FilterDictSuggest + SortSuggestionByQuality
// Past-tense/participle via Synthesize+TagPOS/LemmaOf hooks (fail-closed if nil).
func (r *GermanSpellerRule) Suggest(word string) []string {
	if r == nil || word == "" {
		return nil
	}
	// Java calcSuggestions: getOnlySuggestions exclusive first
	if only := r.OnlySuggestions(word); len(only) > 0 {
		return r.FilterNoSuggestWords(r.FilterProhibitedSuggestions(only))
	}
	if add := r.AdditionalTopSuggestions(word); len(add) > 0 {
		return r.FilterNoSuggestWords(r.FilterProhibitedSuggestions(add))
	}
	if sugs, ok := additionalSuggestionsExact[word]; ok && len(sugs) > 0 {
		return r.FilterNoSuggestWords(r.FilterProhibitedSuggestions(append([]string(nil), sugs...)))
	}
	if sugs := lookupAdditionalSuggestionsRegexp(word); len(sugs) > 0 {
		return r.FilterNoSuggestWords(r.FilterProhibitedSuggestions(sugs))
	}
	// Java: past tense / participle (synth) before uppercase / abbrev / hyphen
	if sugs := r.pastTenseVerbSuggestion(word); len(sugs) > 0 {
		return r.FilterNoSuggestWords(r.FilterProhibitedSuggestions(sugs))
	}
	if sugs := r.participleSuggestion(word); len(sugs) > 0 {
		return r.FilterNoSuggestWords(r.FilterProhibitedSuggestions(sugs))
	}
	// Java getAdditionalTopSuggestionsString: lowercase misspelling → uppercase form if in dict
	if !startsWithUppercase(word) {
		uc := uppercaseFirstChar(word)
		if dictAccepts(uc) && !strings.HasSuffix(uc, ".") {
			return r.FilterNoSuggestWords(r.FilterProhibitedSuggestions([]string{uc}))
		}
	}
	if sugs := r.abbreviationSuggestion(word); len(sugs) > 0 {
		return r.FilterNoSuggestWords(r.FilterProhibitedSuggestions(sugs))
	}
	if sugs := r.suggestHyphenatedCompound(word); len(sugs) > 0 {
		return r.finalizeSuggestions(word, sugs)
	}
	if r.MorfoSpeller == nil && !FilterDictAvailable() {
		return nil
	}
	// CompoundAwareHunspellRule.getSuggestions mix:
	// simple = getCorrectWords(getCandidates) → getFilteredSuggestions
	// noSplit = morfoSpeller.getSuggestions + uppercase-of-lowercase
	// Java uses getSpeller multi (plain-text spelling lists); FilterDict is fall-over.
	simple := r.GetFilteredSuggestions(r.getCorrectWords(r.GetCandidates(word)))
	noSplit := r.morfoSuggest(word)
	var noSplitLower []string
	if startsWithUppercase(word) && !isAllUpperCase(word) {
		for _, s := range r.morfoSuggest(strings.ToLower(word)) {
			noSplitLower = append(noSplitLower, uppercaseFirstChar(s))
		}
	}
	// trailing punctuation handling (word ends with . / ...)
	for _, punct := range []string{"...", "."} {
		if strings.HasSuffix(word, punct) {
			base := strings.TrimSuffix(word, punct)
			for _, s := range r.morfoSuggest(base) {
				noSplit = append(noSplit, s+punct)
			}
		}
	}
	mixed := interleaveSuggestions(noSplit, noSplitLower, simple)
	mixed = dedupeSuggestions(mixed)
	mixed = r.FilterForLanguage(mixed)
	return r.finalizeSuggestions(word, mixed)
}

// finalizeSuggestions applies Accept/prohibit/no-suggest, DE getSuggestions
// stream filters, period restore, sort, and MAX_SUGGESTIONS cap.
func (r *GermanSpellerRule) finalizeSuggestions(word string, sugs []string) []string {
	if r == nil {
		return nil
	}
	sugs = r.FilterNoSuggestWords(r.FilterProhibitedSuggestions(sugs))
	// AcceptSuggestion already inside FilterProhibitedSuggestions
	if strings.HasSuffix(word, ".") {
		for i, s := range sugs {
			if !strings.HasSuffix(s, ".") {
				sugs[i] = s + "."
			}
		}
	}
	sugs = postFilterGetSuggestions(word, sugs)
	sugs = r.SortSuggestionByQuality(word, sugs)
	if len(sugs) > maxGermanSuggestions {
		sugs = sugs[:maxGermanSuggestions]
	}
	return sugs
}

// abbreviationSuggestion ports getAbbreviationSuggestion: short ABK-tagged word → word+"."
func (r *GermanSpellerRule) abbreviationSuggestion(word string) []string {
	// Java: word.length() >= 5 (UTF-16) → no abbreviation suggestion
	if r == nil || r.TagPOS == nil || utf16LenDE(word) >= 5 {
		return nil
	}
	for _, t := range r.TagPOS(word) {
		if strings.HasPrefix(t, "ABK") {
			return []string{word + "."}
		}
	}
	return nil
}

// suggestHyphenatedCompound ports the hyphen-split cartesian suggestion branch
// (Java getAdditionalTopSuggestionsString end): ≤3 segment lists, max 5 results.
func (r *GermanSpellerRule) suggestHyphenatedCompound(word string) []string {
	if r == nil || !strings.Contains(word, "-") || !FilterDictAvailable() {
		return nil
	}
	parts := strings.Split(word, "-")
	if len(parts) < 2 {
		return nil
	}
	startAt, stopAt := 0, len(parts)
	var prefixLocked []string // e.g. Au-pair kept as one joined segment list item
	var suffixLocked string

	if len(parts) >= 2 {
		partial := parts[0] + "-" + parts[1]
		if r.IgnoreWord(partial) || r.IsIgnoredInCompounds(partial) {
			startAt = 2
			prefixLocked = []string{partial}
		}
	}
	if len(parts) >= 2 {
		partial := parts[len(parts)-2] + "-" + parts[len(parts)-1]
		if r.IgnoreWord(partial) || r.IsIgnoredInCompounds(partial) {
			stopAt = len(parts) - 2
			suffixLocked = partial
		}
	}

	// Build list of suggestion options per slot (Java suggestionLists)
	var slots [][]string
	if len(prefixLocked) > 0 {
		slots = append(slots, prefixLocked)
	}
	for i := startAt; i < stopAt; i++ {
		p := parts[i]
		if FilterDictIsMisspelled(p) {
			sugs := r.SortSuggestionByQuality(p, FilterDictSuggest(p))
			if len(sugs) == 0 {
				return nil
			}
			slots = append(slots, sugs)
		} else {
			slots = append(slots, []string{p})
		}
	}
	if suffixLocked != "" {
		slots = append(slots, []string{suffixLocked})
	}
	// Java: only combine when suggestionLists.size() <= 3
	if len(slots) == 0 || len(slots) > 3 {
		return nil
	}
	// Cartesian product
	cur := []string{""}
	for si, opts := range slots {
		var next []string
		for _, base := range cur {
			for _, o := range opts {
				if si == 0 || base == "" {
					next = append(next, o)
				} else {
					next = append(next, base+"-"+o)
				}
			}
		}
		cur = next
	}
	if len(cur) == 0 {
		return nil
	}
	// first min(5, n) results
	if len(cur) > 5 {
		cur = cur[:5]
	}
	// if nothing actually changed from original, skip
	if len(cur) == 1 && cur[0] == word {
		return nil
	}
	return cur
}
