package languagetool

import (
	"strings"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
)

// RegisterBinaryPOSTagger installs lt.TagWord from a Morfologik POS dictionary
// (CFSA2 or FSA5), matching Java BaseTagger/MorfologikTagger lookup behavior.
// Returns false if the dictionary cannot be opened.
func RegisterBinaryPOSTagger(lt *JLanguageTool, dictPath string) bool {
	if lt == nil || dictPath == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	lt.TagWord = BinaryPOSTagWord(d)
	return true
}

// BinaryPOSTagWord returns a TagWord inject from an opened POS dictionary.
// Case logic follows Java PolishTagger (and similar BaseTagger ports): always
// look up the surface form, then for non-lowercase surfaces also merge the
// lowercase dictionary readings (so "Białym" keeps adj:… lemma biały as well as
// proper-noun subst readings). Only when both are empty, try Title case.
func BinaryPOSTagWord(d *atticmorfo.Dictionary) func(token string) []TokenTag {
	if d == nil {
		return nil
	}
	lookup := func(w string) []TokenTag {
		forms, err := d.Lookup(w)
		if err != nil || len(forms) == 0 {
			return nil
		}
		// Polish Morfeusz-style dicts store *multiple* POS tags joined with '+',
		// each segment a full colon tag (subst:sg:nom:m1+adj:…). Java splits those
		// into separate AnalyzedToken readings.
		// Italian morph-it (and similar) use '+' *inside* one tag
		// (VER:part+past+s+m) — BaseTagger keeps the whole string so patterns like
		// VER:part.+past.* match. Only split when every '+' segment contains ':'.
		out := make([]TokenTag, 0, len(forms)*2)
		for _, f := range forms {
			if f.Tag == "" || !strings.Contains(f.Tag, "+") {
				out = append(out, TokenTag{POS: f.Tag, Lemma: f.Stem})
				continue
			}
			parts := strings.Split(f.Tag, "+")
			splitMulti := len(parts) > 1
			if splitMulti {
				for _, part := range parts {
					if !strings.Contains(part, ":") {
						splitMulti = false
						break
					}
				}
			}
			if !splitMulti {
				out = append(out, TokenTag{POS: f.Tag, Lemma: f.Stem})
				continue
			}
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				out = append(out, TokenTag{POS: part, Lemma: f.Stem})
			}
		}
		return out
	}
	return func(token string) []TokenTag {
		if token == "" {
			return nil
		}
		low := strings.ToLower(token)
		var out []TokenTag
		seen := map[string]struct{}{}
		add := func(tags []TokenTag) {
			for _, t := range tags {
				key := t.POS + "\x00" + t.Lemma
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				out = append(out, t)
			}
		}
		// normal case
		add(lookup(token))
		// Java PolishTagger: if (!isLowercase) addTokens(lowerTaggerTokens)
		if low != token {
			add(lookup(low))
		}
		// uppercase of lower only when both empty (Java PolishTagger)
		if len(out) == 0 && low != "" {
			runes := []rune(low)
			if len(runes) > 0 {
				title := strings.ToUpper(string(runes[0])) + string(runes[1:])
				if title != token && title != low {
					add(lookup(title))
				}
			}
		}
		return out
	}
}
