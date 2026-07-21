package de

import (
	"regexp"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// InsertCommaFilter ports org.languagetool.rules.de.InsertCommaFilter.
// TagToken ports GermanTagger.INSTANCE.tag for surface forms of suggestion parts.
// When TagToken is nil, only the 2-token path runs (POS-dependent branches fail-closed).
type InsertCommaFilter struct {
	// TagToken returns POS tags for a surface form; nil skips POS-dependent branches.
	TagToken func(word string) []string
}

func NewInsertCommaFilter() *InsertCommaFilter {
	return &InsertCommaFilter{}
}

// Java: Pattern.compile("\\s") without UNICODE_CHARACTER_CLASS → [ \t\n\x0B\f\r]
// [Ss]agt?, der|die|..., denke|..., bei|für|mit, [Di]ir|...
var (
	insertCommaWS    = regexp.MustCompile(`[ \t\n\v\f\r]`)
	insertCommaSagt  = regexp.MustCompile(`^[Ss]agt?$`)
	insertCommaDenke = regexp.MustCompile(`^(denke|dachte|glaube|schätze|vermute|behaupte)$`)
	insertCommaDer   = regexp.MustCompile(`^(der|die|das|seine|ihre|deine|unsere|meine|folgender|dieser)$`)
	insertCommaBei   = regexp.MustCompile(`^(bei|für|mit)$`)
	insertCommaDir   = regexp.MustCompile(`^([Dd]ir|[Dd]ich|[Ee]uer|[Ee]uch)$`)

	insertCommaTagMu      sync.RWMutex
	defaultInsertCommaTag func(string) []string
)

// SetDefaultInsertCommaTagger wires German tagger for POS-aware comma placement.
func SetDefaultInsertCommaTagger(tag func(string) []string) {
	insertCommaTagMu.Lock()
	defer insertCommaTagMu.Unlock()
	defaultInsertCommaTag = tag
}

func (f *InsertCommaFilter) resolveTag() func(string) []string {
	if f != nil && f.TagToken != nil {
		return f.TagToken
	}
	insertCommaTagMu.RLock()
	defer insertCommaTagMu.RUnlock()
	return defaultInsertCommaTag
}

// AcceptRuleMatch ports InsertCommaFilter.acceptRuleMatch.
func (f *InsertCommaFilter) AcceptRuleMatch(match *rules.RuleMatch, _ map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	tag := f.resolveTag()
	var suggestions []string
	for _, replacement := range match.GetSuggestedReplacements() {
		suggestions = append(suggestions, f.suggestOne(replacement, tag, patternTokenPos, match, patternTokens)...)
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	out.IssueType = match.IssueType
	out.SetSuggestedReplacements(suggestions)
	return out
}

// suggestOne ports the per-replacement body of InsertCommaFilter.acceptRuleMatch.
func (f *InsertCommaFilter) suggestOne(replacement string, tag func(string) []string, patternTokenPos int,
	match *rules.RuleMatch, patternTokens []*languagetool.AnalyzedTokenReadings) []string {
	// Java WHITESPACE.split — single \s, may yield empty parts on double spaces (bug-for-bug).
	parts := insertCommaWS.Split(replacement, -1)
	var out []string
	switch {
	case len(parts) == 2:
		out = append(out, parts[0]+", "+parts[1])
	case len(parts) == 3:
		if tag == nil {
			return nil
		}
		t1, t2, t3 := tag(parts[0]), tag(parts[1]), tag(parts[2])
		if hasTagPrefix(t1, "VER:") && hasTagPrefix(t2, "PRO:PER:") {
			// "Ich hoffe(,) es geht Ihnen gut."
			out = append(out, parts[0]+", "+parts[1]+" "+parts[2])
		} else if insertCommaSagt.MatchString(parts[0]) && parts[1] == "mal" && hasTagPrefix(t3, "VER:") {
			// "Sag mal(,) hast du"
			out = append(out, parts[0]+" "+parts[1]+", "+parts[2])
		} else if hasTagPrefix(t1, "VER:") && hasTagPrefix(t2, "ADV:") && hasTagPrefix(t3, "VER:") {
			// "Ich denke(,) hier kann aber auch ..."
			out = append(out, parts[0]+", "+parts[1]+" "+parts[2])
		}
	case len(parts) >= 4 && len(parts) <= 7:
		if tag == nil {
			return nil
		}
		t1, t2, t3, t4 := tag(parts[0]), tag(parts[1]), tag(parts[2]), tag(parts[3])
		rest1 := strings.Join(parts[1:], " ")
		// Java: patternTokenPos <= 2 || (patternTokenPos == 3 && sentence tokens[1] ADV:)
		early := patternTokenPos <= 2
		if !early && patternTokenPos == 3 && match != nil && match.Sentence != nil {
			toks := match.Sentence.GetTokens()
			if len(toks) >= 2 && toks[1] != nil && toks[1].HasPosTagStartingWith("ADV:") {
				early = true
			}
		}
		if early {
			if len(parts) == 5 &&
				hasTagPrefix(t1, "VER:") && hasTagPrefix(t2, "ART:") && hasTagPrefix(t3, "SUB:") &&
				hasTagPrefix(tag(parts[3]), "SUB:") && hasTagPrefix(tag(parts[4]), "VER:") {
				// "Ist der Kunde Verbraucher(,) gilt ..."
				out = append(out, parts[0]+" "+parts[1]+" "+parts[2]+" "+parts[3]+",")
			} else if len(parts) == 4 &&
				len(patternTokens) >= 2 && patternTokens[0] != nil && patternTokens[1] != nil &&
				patternTokens[0].HasPosTagStartingWith("VER:") &&
				insertCommaDer.MatchString(patternTokens[1].GetToken()) {
				// "Aristoteles meint(,) das Genussleben führe nicht zum Glück."
				out = append(out, parts[0]+", "+rest1)
			} else if hasTagPrefix(t1, "VER:") && hasTagPrefix(t2, "PRO:POS:") && hasTagPrefix(t3, "SUB:") {
				// "Ich glaube(,) eure Premium-Accounts sind noch aktiv."
				out = append(out, parts[0]+", "+rest1)
			} else if hasTagPrefix(t1, "VER:") && hasTagPrefix(t2, "PRO:PER:") && hasTagPrefix(t3, "ADV:INR") {
				// "Weißt du(,) warum diese Regel aus ist?"
				rest2 := strings.Join(parts[2:], " ")
				out = append(out, parts[0]+" "+parts[1]+", "+rest2)
			} else if hasTagPrefix(t1, "VER:") && hasTagPrefix(t2, "PRO:POS:") && hasTagPrefix(t3, "ADJ:") {
				// "Ich glaube(,) eure individuellen Premium-Accounts sind noch aktiv."
				out = append(out, parts[0]+", "+rest1)
			} else if insertCommaDenke.MatchString(parts[0]) && hasTagPrefix(t2, "PRO:DEM:") && hasTagPrefix(t3, "SUB:") {
				// "Ich schätze(,) diese Krawatte passt gut zum Anzug."
				out = append(out, parts[0]+", "+rest1)
			} else if patternTokenPos == 1 && insertCommaBei.MatchString(parts[1]) &&
				insertCommaDir.MatchString(parts[2]) && hasTagPrefix(t4, "VER:") {
				// "Hoffe(,) bei euch ist alles gut."
				out = append(out, parts[0]+", "+rest1)
			}
		}
	}
	return out
}

func hasTagPrefix(tags []string, prefix string) bool {
	for _, t := range tags {
		if strings.HasPrefix(t, prefix) {
			return true
		}
	}
	return false
}

// Suggest is the simple helper used by unit tests (patternTokenPos=1, no pattern tokens/sentence).
func (f *InsertCommaFilter) Suggest(replacement string) []string {
	return f.suggestOne(replacement, f.resolveTag(), 1, nil, nil)
}
