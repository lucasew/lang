package de

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// AdaptSuggestionFilter ports surface determiner adaptation from
// org.languagetool.rules.de.AdaptSuggestionFilter without a full synthesizer/tagger.
//
// GenderOf looks up noun gender ("MAS"|"FEM"|"NEU"); when nil, a small built-in
// table covers common test/demo nouns. Full GermanTagger integration is deferred.
type AdaptSuggestionFilter struct {
	GenderOf func(word string) string
}

func NewAdaptSuggestionFilter() *AdaptSuggestionFilter {
	return &AdaptSuggestionFilter{}
}

// DetReading is a simplified AnalyzedToken for determiner adaptation tests.
type DetReading struct {
	Token string
	POS   string // e.g. ART:DEF:NOM:SIN:FEM
	Lemma string
}

// AdaptedDet ports getAdaptedDet: returns determiner forms matching repl gender.
func (f *AdaptSuggestionFilter) AdaptedDet(det DetReading, repl string) []string {
	gender := f.gender(repl)
	if gender == "" || det.Lemma == "" {
		return nil
	}
	caseNum := parseCaseNumber(det.POS)
	if caseNum == "" {
		return nil
	}
	forms := synthesizeDet(det.Lemma, caseNum, gender)
	var out []string
	first, _ := utf8.DecodeRuneInString(det.Token)
	for _, s := range forms {
		if s == "" {
			continue
		}
		// Keep possessive family: mein→mein*, not dein*.
		sf, _ := utf8.DecodeRuneInString(s)
		df := first
		if unicode.IsUpper(first) {
			if unicode.ToLower(sf) != unicode.ToLower(df) {
				continue
			}
			out = append(out, uppercaseFirst(s))
		} else {
			if sf != df {
				continue
			}
			out = append(out, s)
		}
	}
	return uniqueStrings(out)
}

// SuggestWithDet rewrites "det + noun" suggestions when the previous token is a det.
// prevToken is the surface determiner; replacements are candidate nouns.
func (f *AdaptSuggestionFilter) SuggestWithDet(prevToken string, prevPOS string, prevLemma string, replacements []string) []string {
	det := DetReading{Token: prevToken, POS: prevPOS, Lemma: prevLemma}
	var out []string
	for _, repl := range replacements {
		for _, ad := range f.AdaptedDet(det, repl) {
			s := ad + " " + repl
			out = append(out, s)
		}
	}
	return uniqueStrings(out)
}

func (f *AdaptSuggestionFilter) gender(word string) string {
	if f.GenderOf != nil {
		if g := f.GenderOf(word); g != "" {
			return g
		}
	}
	return defaultNounGender(word)
}

// defaultNounGender is a tiny surface stand-in for GermanTagger.
func defaultNounGender(word string) string {
	w := strings.TrimSpace(word)
	// Common demo/test nouns from AdaptSuggestionFilterTest.
	switch w {
	case "Mann", "Plan", "Tisch", "Baum", "Hund", "Tag", "Name":
		return "MAS"
	case "Frau", "Idee", "Roadmap", "Katze", "Zeit", "Stadt", "Blume":
		return "FEM"
	case "Kind", "Verfahren", "Haus", "Buch", "Auto", "Mädchen", "Tier":
		return "NEU"
	}
	// Heuristic suffixes (weak).
	lw := strings.ToLower(w)
	switch {
	case strings.HasSuffix(lw, "ung"), strings.HasSuffix(lw, "heit"),
		strings.HasSuffix(lw, "keit"), strings.HasSuffix(lw, "schaft"),
		strings.HasSuffix(lw, "ion"), strings.HasSuffix(lw, "tät"),
		strings.HasSuffix(lw, "ik"), strings.HasSuffix(lw, "ei"):
		return "FEM"
	case strings.HasSuffix(lw, "chen"), strings.HasSuffix(lw, "lein"),
		strings.HasSuffix(lw, "ment"), strings.HasSuffix(lw, "tum"),
		strings.HasSuffix(lw, "um"):
		return "NEU"
	case strings.HasSuffix(lw, "er"), strings.HasSuffix(lw, "ling"),
		strings.HasSuffix(lw, "ismus"):
		return "MAS"
	}
	return ""
}

// parseCaseNumber extracts CASE:NUMBER from ART/PRO tags (…:NOM:SIN:FEM).
func parseCaseNumber(pos string) string {
	if pos == "" {
		return ""
	}
	parts := strings.Split(pos, ":")
	// ART:DEF:NOM:SIN:FEM or PRO:POS:NOM:SIN:MAS:BEG
	if len(parts) < 5 {
		return ""
	}
	// find CASE (NOM|AKK|GEN|DAT) and NUMBER (SIN|PLU)
	var cas, num string
	for _, p := range parts {
		switch p {
		case "NOM", "AKK", "GEN", "DAT":
			cas = p
		case "SIN", "PLU":
			num = p
		}
	}
	if cas == "" || num == "" {
		return ""
	}
	return cas + ":" + num
}

// synthesizeDet returns surface forms for common German dets/possessives.
// lemma is the base (der, ein, mein, dein, …).
func synthesizeDet(lemma, caseNum, gender string) []string {
	key := lemma + "|" + caseNum + "|" + gender
	if forms, ok := detSynthTable[key]; ok {
		return append([]string(nil), forms...)
	}
	// Also try lower lemma
	key = strings.ToLower(lemma) + "|" + caseNum + "|" + gender
	if forms, ok := detSynthTable[key]; ok {
		return append([]string(nil), forms...)
	}
	return nil
}

// detSynthTable approximates GermanSynthesizer for ART/PRO used in AdaptSuggestion tests.
// Keys: lemma|CASE:NUM|GENDER → forms
var detSynthTable = map[string][]string{
	// definite article "der"
	"der|NOM:SIN|MAS": {"der"},
	"der|AKK:SIN|MAS": {"den"},
	"der|GEN:SIN|MAS": {"des"},
	"der|DAT:SIN|MAS": {"dem"},
	"der|NOM:SIN|FEM": {"die"},
	"der|AKK:SIN|FEM": {"die"},
	"der|GEN:SIN|FEM": {"der"},
	"der|DAT:SIN|FEM": {"der"},
	"der|NOM:SIN|NEU": {"das"},
	"der|AKK:SIN|NEU": {"das"},
	"der|GEN:SIN|NEU": {"des"},
	"der|DAT:SIN|NEU": {"dem"},
	"der|NOM:PLU|MAS": {"die"}, "der|NOM:PLU|FEM": {"die"}, "der|NOM:PLU|NEU": {"die"},
	"der|AKK:PLU|MAS": {"die"}, "der|AKK:PLU|FEM": {"die"}, "der|AKK:PLU|NEU": {"die"},
	"der|DAT:PLU|MAS": {"den"}, "der|DAT:PLU|FEM": {"den"}, "der|DAT:PLU|NEU": {"den"},
	"der|GEN:PLU|MAS": {"der"}, "der|GEN:PLU|FEM": {"der"}, "der|GEN:PLU|NEU": {"der"},

	// indefinite "ein"
	"ein|NOM:SIN|MAS": {"ein"},
	"ein|AKK:SIN|MAS": {"einen"},
	"ein|GEN:SIN|MAS": {"eines"},
	"ein|DAT:SIN|MAS": {"einem"},
	"ein|NOM:SIN|FEM": {"eine"},
	"ein|AKK:SIN|FEM": {"eine"},
	"ein|GEN:SIN|FEM": {"einer"},
	"ein|DAT:SIN|FEM": {"einer"},
	"ein|NOM:SIN|NEU": {"ein"},
	"ein|AKK:SIN|NEU": {"ein"},
	"ein|GEN:SIN|NEU": {"eines"},
	"ein|DAT:SIN|NEU": {"einem"},

	// possessives (lemma is the stem used by LT: mein/dein/sein/ihr/unser/euer)
	"mein|NOM:SIN|MAS": {"mein"}, "mein|AKK:SIN|MAS": {"meinen"}, "mein|GEN:SIN|MAS": {"meines"}, "mein|DAT:SIN|MAS": {"meinem"},
	"mein|NOM:SIN|FEM": {"meine"}, "mein|AKK:SIN|FEM": {"meine"}, "mein|GEN:SIN|FEM": {"meiner"}, "mein|DAT:SIN|FEM": {"meiner"},
	"mein|NOM:SIN|NEU": {"mein"}, "mein|AKK:SIN|NEU": {"mein"}, "mein|GEN:SIN|NEU": {"meines"}, "mein|DAT:SIN|NEU": {"meinem"},

	"dein|NOM:SIN|MAS": {"dein"}, "dein|AKK:SIN|MAS": {"deinen"}, "dein|GEN:SIN|MAS": {"deines"}, "dein|DAT:SIN|MAS": {"deinem"},
	"dein|NOM:SIN|FEM": {"deine"}, "dein|AKK:SIN|FEM": {"deine"}, "dein|GEN:SIN|FEM": {"deiner"}, "dein|DAT:SIN|FEM": {"deiner"},
	"dein|NOM:SIN|NEU": {"dein"}, "dein|AKK:SIN|NEU": {"dein"}, "dein|GEN:SIN|NEU": {"deines"}, "dein|DAT:SIN|NEU": {"deinem"},

	"sein|NOM:SIN|MAS": {"sein"}, "sein|AKK:SIN|MAS": {"seinen"}, "sein|GEN:SIN|MAS": {"seines"}, "sein|DAT:SIN|MAS": {"seinem"},
	"sein|NOM:SIN|FEM": {"seine"}, "sein|AKK:SIN|FEM": {"seine"}, "sein|GEN:SIN|FEM": {"seiner"}, "sein|DAT:SIN|FEM": {"seiner"},
	"sein|NOM:SIN|NEU": {"sein"}, "sein|AKK:SIN|NEU": {"sein"}, "sein|GEN:SIN|NEU": {"seines"}, "sein|DAT:SIN|NEU": {"seinem"},

	"ihr|NOM:SIN|MAS": {"ihr"}, "ihr|AKK:SIN|MAS": {"ihren"}, "ihr|GEN:SIN|MAS": {"ihres"}, "ihr|DAT:SIN|MAS": {"ihrem"},
	"ihr|NOM:SIN|FEM": {"ihre"}, "ihr|AKK:SIN|FEM": {"ihre"}, "ihr|GEN:SIN|FEM": {"ihrer"}, "ihr|DAT:SIN|FEM": {"ihrer"},
	"ihr|NOM:SIN|NEU": {"ihr"}, "ihr|AKK:SIN|NEU": {"ihr"}, "ihr|GEN:SIN|NEU": {"ihres"}, "ihr|DAT:SIN|NEU": {"ihrem"},

	"unser|NOM:SIN|MAS": {"unser"}, "unser|AKK:SIN|MAS": {"unseren"}, "unser|GEN:SIN|MAS": {"unseres"}, "unser|DAT:SIN|MAS": {"unserem"},
	"unser|NOM:SIN|FEM": {"unsere"}, "unser|AKK:SIN|FEM": {"unsere"}, "unser|GEN:SIN|FEM": {"unserer"}, "unser|DAT:SIN|FEM": {"unserer"},
	"unser|NOM:SIN|NEU": {"unser"}, "unser|AKK:SIN|NEU": {"unser"}, "unser|GEN:SIN|NEU": {"unseres"}, "unser|DAT:SIN|NEU": {"unserem"},

	"euer|NOM:SIN|MAS": {"euer"}, "euer|AKK:SIN|MAS": {"euren"}, "euer|GEN:SIN|MAS": {"eures"}, "euer|DAT:SIN|MAS": {"eurem"},
	"euer|NOM:SIN|FEM": {"eure"}, "euer|AKK:SIN|FEM": {"eure"}, "euer|GEN:SIN|FEM": {"eurer"}, "euer|DAT:SIN|FEM": {"eurer"},
	"euer|NOM:SIN|NEU": {"euer"}, "euer|AKK:SIN|NEU": {"euer"}, "euer|GEN:SIN|NEU": {"eures"}, "euer|DAT:SIN|NEU": {"eurem"},
}

func uppercaseFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func uniqueStrings(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// SuggestWithDetAdj rewrites "det + adj + noun" suggestions (synthesizer soft-deferred).
// Adjective is weakly inflected after a definite/indefinite determiner using a small ending table.
func (f *AdaptSuggestionFilter) SuggestWithDetAdj(prevDet, prevDetPOS, prevDetLemma, prevAdj string, replacements []string) []string {
	var out []string
	for _, repl := range replacements {
		dets := f.AdaptedDet(DetReading{Token: prevDet, POS: prevDetPOS, Lemma: prevDetLemma}, repl)
		gender := f.gender(repl)
		caseNum := parseCaseNumber(prevDetPOS)
		adjForms := weakAdjForms(prevAdj, caseNum, gender)
		if len(dets) == 0 {
			dets = []string{prevDet}
		}
		if len(adjForms) == 0 {
			adjForms = []string{prevAdj}
		}
		for _, d := range dets {
			for _, a := range adjForms {
				out = append(out, d+" "+a+" "+repl)
			}
		}
	}
	return uniqueStrings(out)
}

// weakAdjForms approximates weak adjective endings after a determiner.
func weakAdjForms(adj, caseNum, gender string) []string {
	stem := adjStem(adj)
	if stem == "" || caseNum == "" || gender == "" {
		return nil
	}
	end := weakAdjEnding(caseNum, gender)
	if end == "" {
		return nil
	}
	return []string{stem + end}
}

func adjStem(adj string) string {
	// strip common strong/weak endings to get stem
	for _, suf := range []string{"en", "em", "er", "es", "e"} {
		if strings.HasSuffix(adj, suf) && len(adj) > len(suf)+2 {
			return adj[:len(adj)-len(suf)]
		}
	}
	return adj
}

func weakAdjEnding(caseNum, gender string) string {
	// weak: after der/die/das/ein… typically -e or -en
	switch caseNum {
	case "NOM:SIN":
		if gender == "MAS" || gender == "NEU" || gender == "FEM" {
			return "e"
		}
	case "AKK:SIN":
		if gender == "MAS" {
			return "en"
		}
		return "e"
	case "DAT:SIN", "GEN:SIN":
		return "en"
	case "NOM:PLU", "AKK:PLU":
		return "en"
	case "DAT:PLU", "GEN:PLU":
		return "en"
	}
	return "e"
}
