package de

import (
	"regexp"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
)

// InsertCommaFilter ports org.languagetool.rules.de.InsertCommaFilter.
// Specific to KOMMA_ZWISCHEN_HAUPT_UND_NEBENSATZ — sets the comma suggestion when possible.
//
// Java getTag uses GermanTagger.INSTANCE.tag(singletonList(part)).
// Go: DiscoveredGermanTagger (wired GermanTagger.INSTANCE) or optional Tag override.
type InsertCommaFilter struct {
	// Tag ports GermanTagger.INSTANCE for suggestion-part lookups (tests may inject).
	// Nil → DiscoveredGermanTagger / defaultInsertCommaTagger.
	Tag *detag.GermanTagger
}

func NewInsertCommaFilter() *InsertCommaFilter {
	return &InsertCommaFilter{}
}

// Java: Pattern.compile("\\s") without UNICODE_CHARACTER_CLASS → [ \t\n\x0B\f\r]
// Patterns: [Ss]agt?, der|die|…, denke|…, bei|für|mit, [Di]ir|… (Matcher.matches = full string)
var (
	insertCommaWS    = regexp.MustCompile(`[ \t\n\v\f\r]`)
	insertCommaSagt  = regexp.MustCompile(`^[Ss]agt?$`)
	insertCommaDenke = regexp.MustCompile(`^(denke|dachte|glaube|schätze|vermute|behaupte)$`)
	insertCommaDer   = regexp.MustCompile(`^(der|die|das|seine|ihre|deine|unsere|meine|folgender|dieser)$`)
	insertCommaBei   = regexp.MustCompile(`^(bei|für|mit)$`)
	// Java: compile("[Di]ir|[Dd]ich|[Ee]uer|[Ee]uch") — note [Di]ir not [Dd]ir (bug-for-bug).
	insertCommaDir = regexp.MustCompile(`^([Di]ir|[Dd]ich|[Ee]uer|[Ee]uch)$`)

	insertCommaTaggerMu      sync.RWMutex
	defaultInsertCommaTagger *detag.GermanTagger
)

// SetDefaultInsertCommaTagger wires process-wide GermanTagger.INSTANCE for this filter.
// Prefer WireGermanTaggerDefaults; kept for POS-func wiring compatibility.
func SetDefaultInsertCommaTagger(tag func(string) []string) {
	// Historical POS-func hook: ensure INSTANCE tagger is installed when resources open.
	// Real tagging uses GermanTagger.Tag (see getTag), not the []string func.
	_ = tag
	insertCommaTaggerMu.Lock()
	defer insertCommaTaggerMu.Unlock()
	if defaultInsertCommaTagger == nil {
		defaultInsertCommaTagger = DiscoveredGermanTagger()
	}
}

// SetDefaultInsertCommaGermanTagger installs the GermanTagger.INSTANCE twin.
func SetDefaultInsertCommaGermanTagger(t *detag.GermanTagger) {
	insertCommaTaggerMu.Lock()
	defer insertCommaTaggerMu.Unlock()
	defaultInsertCommaTagger = t
}

func (f *InsertCommaFilter) resolveTagger() *detag.GermanTagger {
	if f != nil && f.Tag != nil {
		return f.Tag
	}
	insertCommaTaggerMu.RLock()
	t := defaultInsertCommaTagger
	insertCommaTaggerMu.RUnlock()
	if t != nil {
		return t
	}
	return DiscoveredGermanTagger()
}

// AcceptRuleMatch ports InsertCommaFilter.acceptRuleMatch.
func (f *InsertCommaFilter) AcceptRuleMatch(match *rules.RuleMatch, _ map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	// Java: new RuleMatch(match.getRule(), match.getSentence(), match.getFromPos(), match.getToPos(), match.getMessage(), match.getShortMessage())
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	var suggestions []string
	for _, replacement := range match.GetSuggestedReplacements() {
		// Java try/catch IOException → RuntimeException; Go tagger has no IO error path.
		suggestions = append(suggestions, f.suggestionsForReplacement(replacement, patternTokenPos, match, patternTokens)...)
	}
	out.SetSuggestedReplacements(suggestions)
	// Java: ruleMatch.setType(match.getType())
	out.IssueType = match.IssueType
	return out
}

// suggestionsForReplacement ports the per-replacement body of acceptRuleMatch.
func (f *InsertCommaFilter) suggestionsForReplacement(replacement string, patternTokenPos int,
	match *rules.RuleMatch, patternTokens []*languagetool.AnalyzedTokenReadings) []string {
	// Java WHITESPACE.split — String.split discards trailing empties.
	parts := javaWhitespaceSplit(replacement)
	var suggestions []string
	if len(parts) == 2 {
		suggestions = append(suggestions, parts[0]+", "+parts[1])
	} else if len(parts) == 3 {
		tags1 := f.getTag(0, parts)
		tags2 := f.getTag(1, parts)
		tags3 := f.getTag(2, parts)
		if hasInsertCommaTag(tags1, "VER:") && hasInsertCommaTag(tags2, "PRO:PER:") {
			// "Ich hoffe(,) es geht Ihnen gut."
			suggestions = append(suggestions, parts[0]+", "+parts[1]+" "+parts[2])
		} else if insertCommaSagt.MatchString(parts[0]) && parts[1] == "mal" && hasInsertCommaTag(tags3, "VER:") {
			// "Sag mal(,) hast du"
			suggestions = append(suggestions, parts[0]+" "+parts[1]+", "+parts[2])
		} else if hasInsertCommaTag(tags1, "VER:") && hasInsertCommaTag(tags2, "ADV:") && hasInsertCommaTag(tags3, "VER:") {
			// "Ich denke(,) hier kann aber auch ..."
			suggestions = append(suggestions, parts[0]+", "+parts[1]+" "+parts[2])
		}
	} else if len(parts) >= 4 && len(parts) <= 7 {
		tags1 := f.getTag(0, parts)
		tags2 := f.getTag(1, parts)
		tags3 := f.getTag(2, parts)
		tags4 := f.getTag(3, parts)
		rest1 := strings.Join(parts[1:], " ")
		// Java: patternTokenPos <= 2 || (patternTokenPos == 3 && match.getSentence().getTokens().length >= 1 && match.getSentence().getTokens()[1].hasPosTagStartingWith("ADV:"))
		early := patternTokenPos <= 2
		if !early && patternTokenPos == 3 && match != nil && match.Sentence != nil {
			toks := match.Sentence.GetTokens()
			// Java: length >= 1 then tokens[1] — needs index 1 present.
			if len(toks) >= 2 && toks[1] != nil && toks[1].HasPosTagStartingWith("ADV:") {
				early = true
			}
		}
		if early {
			if len(parts) == 5 &&
				hasInsertCommaTag(tags1, "VER:") && hasInsertCommaTag(tags2, "ART:") && hasInsertCommaTag(tags3, "SUB:") &&
				hasInsertCommaTag(f.getTag(3, parts), "SUB:") && hasInsertCommaTag(f.getTag(4, parts), "VER:") {
				// "Ist der Kunde Verbraucher(,) gilt ..."
				suggestions = append(suggestions, parts[0]+" "+parts[1]+" "+parts[2]+" "+parts[3]+",")
			} else if len(parts) == 4 &&
				len(patternTokens) >= 2 && patternTokens[0] != nil && patternTokens[1] != nil &&
				patternTokens[0].HasPosTagStartingWith("VER:") &&
				insertCommaDer.MatchString(patternTokens[1].GetToken()) {
				// "Aristoteles meint(,) das Genussleben führe nicht zum Glück."
				suggestions = append(suggestions, parts[0]+", "+rest1)
			} else if hasInsertCommaTag(tags1, "VER:") && hasInsertCommaTag(tags2, "PRO:POS:") && hasInsertCommaTag(tags3, "SUB:") {
				// "Ich glaube(,) eure Premium-Accounts sind noch aktiv."
				suggestions = append(suggestions, parts[0]+", "+rest1)
			} else if hasInsertCommaTag(tags1, "VER:") && hasInsertCommaTag(tags2, "PRO:PER:") && hasInsertCommaTag(tags3, "ADV:INR") {
				// "Weißt du(,) warum diese Regel aus ist?"
				rest2 := strings.Join(parts[2:], " ")
				suggestions = append(suggestions, parts[0]+" "+parts[1]+", "+rest2)
			} else if hasInsertCommaTag(tags1, "VER:") && hasInsertCommaTag(tags2, "PRO:POS:") && hasInsertCommaTag(tags3, "ADJ:") {
				// "Ich glaube(,) eure individuellen Premium-Accounts sind noch aktiv."
				suggestions = append(suggestions, parts[0]+", "+rest1)
			} else if insertCommaDenke.MatchString(parts[0]) && hasInsertCommaTag(tags2, "PRO:DEM:") && hasInsertCommaTag(tags3, "SUB:") {
				// "Ich schätze(,) diese Krawatte passt gut zum Anzug."
				suggestions = append(suggestions, parts[0]+", "+rest1)
			} else if patternTokenPos == 1 && insertCommaBei.MatchString(parts[1]) &&
				insertCommaDir.MatchString(parts[2]) && hasInsertCommaTag(tags4, "VER:") {
				// "Hoffe(,) bei euch ist alles gut."
				suggestions = append(suggestions, parts[0]+", "+rest1)
			}
		}
	}
	return suggestions
}

// getTag ports InsertCommaFilter.getTag: GermanTagger.INSTANCE.tag(singletonList(parts[i])).
func (f *InsertCommaFilter) getTag(i int, parts []string) []*languagetool.AnalyzedTokenReadings {
	if i < 0 || i >= len(parts) {
		return nil
	}
	tagger := f.resolveTagger()
	if tagger == nil {
		return nil
	}
	// Java: return GermanTagger.INSTANCE.tag(singletonList(parts[i]));
	return tagger.Tag([]string{parts[i]})
}

// hasInsertCommaTag ports InsertCommaFilter.hasTag:
// tags.stream().anyMatch(k -> k.hasPosTagStartingWith(tagStart))
func hasInsertCommaTag(tags []*languagetool.AnalyzedTokenReadings, tagStart string) bool {
	for _, k := range tags {
		if k != nil && k.HasPosTagStartingWith(tagStart) {
			return true
		}
	}
	return false
}

// javaWhitespaceSplit ports Java String.split("\\s") / Pattern.split: trailing empties dropped.
func javaWhitespaceSplit(s string) []string {
	parts := insertCommaWS.Split(s, -1)
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

// Suggest is a thin helper for unit tests (patternTokenPos=1, no pattern tokens/sentence).
func (f *InsertCommaFilter) Suggest(replacement string) []string {
	return f.suggestionsForReplacement(replacement, 1, nil, nil)
}
