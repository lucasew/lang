package ro

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RomanianWordRepeatBeginningRule ports org.languagetool.rules.ro.RomanianWordRepeatBeginningRule.
// Extends WordRepeatBeginningRule with Romanian POS-based isAdverb (G* tags; see resource/ro/coduri.html).
//
// Java ctor takes ResourceBundle + Language; Language is unused by this type and
// the core WordRepeatBeginningRule match loop — Go takes messages only (same as
// other language WRB ports).
type RomanianWordRepeatBeginningRule struct {
	*rules.WordRepeatBeginningRule
	// allowAmbiguousAdverbs ports Java protected allowAmbiguousAdverbs() default false.
	// Ambiguous adverbs have an adverb (G*) tag along with other tags
	// (e.g. Romanian "Și" can be adverb predicativ, conjuncție, or pronoun).
	allowAmbiguousAdverbs bool
}

// NewRomanianWordRepeatBeginningRule ports RomanianWordRepeatBeginningRule(messages, language).
func NewRomanianWordRepeatBeginningRule(messages map[string]string) *RomanianWordRepeatBeginningRule {
	base := rules.NewWordRepeatBeginningRule(messages)
	// Java @Override getId
	base.IDOverride = "ROMANIAN_WORD_REPEAT_BEGINNING_RULE"
	// Java has no addExamplePair on this class.
	r := &RomanianWordRepeatBeginningRule{WordRepeatBeginningRule: base}
	// Java @Override isAdverb — no isException / getSuggestions overrides.
	base.IsAdverbFn = r.isAdverb
	return r
}

// AllowAmbiguousAdverbs ports Java protected allowAmbiguousAdverbs() (default false).
func (r *RomanianWordRepeatBeginningRule) AllowAmbiguousAdverbs() bool {
	if r == nil {
		return false
	}
	return r.allowAmbiguousAdverbs
}

// isAdverb ports RomanianWordRepeatBeginningRule.isAdverb:
//
//	boolean isAdverb = false;
//	for (AnalyzedToken analyzedToken : token.getReadings()) {
//	  if (analyzedToken.getPOSTag() != null) {
//	    if (analyzedToken.getPOSTag().startsWith("G")) {
//	      isAdverb = true;
//	    } else {
//	      if (!allowAmbiguousAdverbs()) {
//	        return false;
//	      }
//	    }
//	  }
//	}
//	return isAdverb;
//
// Control flow is bug-for-bug: any non-null non-G POS returns false when
// allowAmbiguousAdverbs is false, even if another reading is G*.
func (r *RomanianWordRepeatBeginningRule) isAdverb(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return false
	}
	isAdverb := false
	for _, analyzedToken := range token.GetReadings() {
		if analyzedToken == nil {
			continue
		}
		// Java: if (analyzedToken.getPOSTag() != null)
		if analyzedToken.GetPOSTag() != nil {
			tag := *analyzedToken.GetPOSTag()
			if strings.HasPrefix(tag, "G") { // see resource/ro/coduri.html — G=adverb
				isAdverb = true
			} else {
				if !r.AllowAmbiguousAdverbs() {
					return false
				}
			}
		}
	}
	return isAdverb
}
