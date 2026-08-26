package chunking

import (
	"fmt"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// RussianChunker ports org.languagetool.chunking.RussianChunker (@Experimental).
// REGEXES1 / REGEXES2 are OpenRegex patterns; tokens start as O (no invent POS→BIO).
// Control flow matches Java: getBasicChunks (REGEXES1) → REGEXES2 → assignChunksToReadings.
type RussianChunker struct{}

func NewRussianChunker() *RussianChunker { return &RussianChunker{} }

// FILTER_TAGS ports Java RussianChunker.FILTER_TAGS (overwrite mode removes these exact names).
var russianFilterTags = map[string]bool{
	"PP": true, "NPP": true, "NPS": true, "MayMissingYO": true,
	"VP": true, "SBAR": true, "ADJP": true, "DPT": true,
}

// russianChunkerDebug mirrors Java RussianChunker.debug (setDebug / isDebug).
var russianChunkerDebug bool

// SetRussianChunkerDebug ports RussianChunker.setDebug (deprecated for internal use only).
func SetRussianChunkerDebug(debugMode bool) { russianChunkerDebug = debugMode }

// IsRussianChunkerDebug ports RussianChunker.isDebug.
func IsRussianChunkerDebug() bool { return russianChunkerDebug }

type russianPhraseType int

const (
	ruPhraseNP russianPhraseType = iota
	ruPhraseNPS
	ruPhraseNPP
	ruPhrasePP
	ruPhraseMayMissingYO
	ruPhraseVP
	ruPhraseSBAR
	ruPhraseADJP
	ruPhraseDPT
)

func (p russianPhraseType) name() string {
	switch p {
	case ruPhraseNP:
		return "NP"
	case ruPhraseNPS:
		return "NPS"
	case ruPhraseNPP:
		return "NPP"
	case ruPhrasePP:
		return "PP"
	case ruPhraseMayMissingYO:
		return "MayMissingYO"
	case ruPhraseVP:
		return "VP"
	case ruPhraseSBAR:
		return "SBAR"
	case ruPhraseADJP:
		return "ADJP"
	case ruPhraseDPT:
		return "DPT"
	default:
		return ""
	}
}

type russianRegex struct {
	pattern   string
	phrase    russianPhraseType
	overwrite bool
}

// Java RussianChunker.REGEXES1 (order + overwrite flags exact).
var russianRegexes1 = []russianRegex{
	// Иванов Иван Иванович
	{`<posre='NN:(Name|Fam|Patr):.*'> <posre='NN:(Name|Fam|Patr):.*'>+ `, ruPhraseNP, true},
	// Иванов И.И.
	{`<posre='NN:Fam:.*'> <regexCS=[А-ЯЁ]> <.> <regexCS=[А-ЯЁ]> <.> `, ruPhraseNP, true},
	// И.И. Иванов
	{`<regexCS=[А-ЯЁ]> <.> <regexCS=[А-ЯЁ]> <.> <posre='NN:Fam:.*'> `, ruPhraseNP, true},
	// verb+verb
	{`<posre='VB:.*:.*' & !posre='NN:.*'>* `, ruPhraseVP, false},
	{`<если>`, ruPhraseSBAR, false},
	{`<поэтому>`, ruPhraseSBAR, false},
	// noun phrase
	{`<posre='ADJ:Posit:.*:.*'> <posre='NN:(Anim|Inanim):.*' & !posre='NN:(Anim|Inanim):.*:(R|D|T|P)'> `, ruPhraseNP, true},
	{`<posre='ADJ:Posit:.*:.*'> <posre='NN:(Anim|Inanim):.*' & !posre='NN:(Anim|Inanim):.*:(R|D|T|P)'> <posre='NN:(Anim|Inanim):.*'> `, ruPhraseNP, true},
	// adj → participle phrase
	{`<posre='ADJ:Posit:.*:.*'> <posre='NN:(Anim|Inanim):.*' & !posre='NN:(Anim|Inanim):.*:(Nom|V)'> <posre='NN:(Anim|Inanim):.*:(Nom|V)' & !posre='NN:(Anim|Inanim):.*:(R|D|T|P)'> `, ruPhraseADJP, true},
	// adverbial participle
	{`<posre='DPT:.*:.*' & !pos='PREP'> `, ruPhraseDPT, false},
	{`<posre='DPT:.*:.*' & !pos='PREP'> <posre='NN:.*:.*:(R|D|T|P)' > `, ruPhraseDPT, true},
	{`<posre='DPT:.*:.*' & !pos='PREP'> <posre='PREP'> <posre='NN:.*:.*:(R|D|T|P)' > `, ruPhraseDPT, true},
	// participle
	{`<posre='PT:.*:.*'> `, ruPhraseADJP, false},
	{`<posre='PT:.*:.*'> <pos='ADV' > `, ruPhraseADJP, true},
	{`<posre='PT:.*:.*'> <posre='NN:.*:.*:(R|D|T|P)' > `, ruPhraseADJP, true},
	{`<posre='PT:.*:.*'> <posre='PREP'> <posre='NN:.*:.*:(R|D|T|P|V)' > `, ruPhraseADJP, true},
	{`<posre='PT:.*:.*'> <posre='PREP'> <posre='ADJ:.*:.*:(R|D|T|P|V)' > <posre='NN:.*:.*:(R|D|T|P|V)' > `, ruPhraseADJP, true},
	{`<posre='PT:.*:.*'> <posre='NN:(Anim|Inanim):.*' & !posre='NN:(Anim|Inanim):.*:(Nom|V)'> <posre='NN:(Anim|Inanim):.*:(Nom|V)' & !posre='NN:(Anim|Inanim):.*:(R|D|T|P)'> `, ruPhraseADJP, true},
	{`<posre='PT:.*:.*'> <posre='PNN:.*' & !posre='PNN:.*:Nom:.*'> <posre='NN:(Anim|Inanim):.*:(Nom|V)' & !posre='NN:(Anim|Inanim):.*:(R|D|T|P)'> `, ruPhraseADJP, true},
	{`<posre='PT:.*:.*'> <posre='ADJ:.*:.*' > `, ruPhraseADJP, false},
	{`<тов>`, ruPhraseNP, false},
}

// Java RussianChunker.REGEXES2 (order + overwrite flags exact).
var russianRegexes2 = []russianRegex{
	// "Маша и Миша":
	{`<posre=NN:Name:.*> <и> <posre=NN:Name:.*>`, ruPhraseNPP, true},
	{`<posre=NN:Name:.*> <или> <posre=NN:Name:.*>`, ruPhraseNPP, true},
	// не + VB
	{`<не> <posre='VB:.*:.*' & !posre='NN:.*'>* `, ruPhraseVP, false},
}

// AddChunkTags ports RussianChunker.addChunkTags:
// getBasicChunks (REGEXES1) → apply REGEXES2 → assignChunksToReadings.
func (c *RussianChunker) AddChunkTags(tokens []*languagetool.AnalyzedTokenReadings) {
	if c == nil {
		return
	}
	toks := getRussianBasicChunks(tokens)
	factory := NewChunkTokenFactory(false)
	for _, regex := range russianRegexes2 {
		applyRussianRegex(regex, toks, factory)
	}
	assignRussianChunksToReadings(toks)
}

// GetBasicChunks ports RussianChunker.getBasicChunks — REGEXES1 only; does not mutate readings.
func (c *RussianChunker) GetBasicChunks(tokens []*languagetool.AnalyzedTokenReadings) []ChunkTaggedToken {
	if c == nil {
		return nil
	}
	return getRussianBasicChunks(tokens)
}

func getRussianBasicChunks(tokenReadings []*languagetool.AnalyzedTokenReadings) []ChunkTaggedToken {
	var chunkTaggedTokens []ChunkTaggedToken
	for _, tokenReading := range tokenReadings {
		if tokenReading == nil || tokenReading.IsWhitespace() {
			continue
		}
		// Java: (!tokenReading.isWhitespace()) && (!tokenReading.getChunkTags().contains(new ChunkTag("MayMissingYO")))
		skip := false
		for _, ct := range tokenReading.GetChunkTags() {
			if ct == "MayMissingYO" {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		chunkTaggedTokens = append(chunkTaggedTokens,
			NewChunkTaggedToken(tokenReading.GetToken(), []ChunkTag{NewChunkTag("O")}, tokenReading))
	}
	if russianChunkerDebug {
		fmt.Println("=============== CHUNKER INPUT ===============")
		fmt.Print(russianChunkDebugString(chunkTaggedTokens))
	}
	factory := NewChunkTokenFactory(false)
	for _, regex := range russianRegexes1 {
		applyRussianRegex(regex, chunkTaggedTokens, factory)
	}
	return chunkTaggedTokens
}

// applyRussianRegex ports RussianChunker.apply / doApplyRegex.
func applyRussianRegex(regex russianRegex, tokens []ChunkTaggedToken, factory func(string) func(ChunkTaggedToken) bool) {
	pat := ExpandRussianChunkSyntax(regex.pattern)
	re := CompileOpenRegex(pat, factory)
	prevDebug := ""
	if russianChunkerDebug {
		prevDebug = russianChunkDebugString(tokens)
	}
	// Java wraps exceptions as RuntimeException with pattern + tokens context.
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Sprintf("Could not apply chunk regexp '%s' to tokens: %v: %v", regex.pattern, tokens, r))
		}
	}()
	matches := re.FindAll(tokens)
	for _, m := range matches {
		for i := m.Start; i < m.End; i++ {
			token := tokens[i]
			newChunkTags := make([]ChunkTag, 0, len(token.ChunkTags)+1)
			newChunkTags = append(newChunkTags, token.ChunkTags...)
			if regex.overwrite {
				filtered := make([]ChunkTag, 0, len(newChunkTags))
				for _, ct := range newChunkTags {
					if !russianFilterTags[ct.GetChunkTag()] {
						filtered = append(filtered, ct)
					}
				}
				newChunkTags = filtered
			}
			newTag := russianChunkTag(regex.phrase, m, i)
			if newTag == "" {
				continue
			}
			has := false
			for _, ct := range newChunkTags {
				if ct.GetChunkTag() == newTag {
					has = true
					break
				}
			}
			// Java: if (!newChunkTags.contains(newTag)) { add; remove O }
			if !has {
				newChunkTags = append(newChunkTags, NewChunkTag(newTag))
				final := make([]ChunkTag, 0, len(newChunkTags))
				for _, ct := range newChunkTags {
					if ct.GetChunkTag() != "O" {
						final = append(final, ct)
					}
				}
				newChunkTags = final
			}
			tokens[i] = NewChunkTaggedToken(token.Token, newChunkTags, token.Readings)
		}
	}
	if russianChunkerDebug {
		debug := russianChunkDebugString(tokens)
		if debug != prevDebug {
			fmt.Printf("=== Applied %s <= %s (overwrite: %v) ===\n", regex.phrase.name(), regex.pattern, regex.overwrite)
			if regex.overwrite {
				fmt.Printf("Note: overwrite mode, replacing old %v tags\n", []string{"PP", "NPP", "NPS", "MayMissingYO", "VP", "SBAR", "ADJP", "DPT"})
			}
			fmt.Print(debug)
			fmt.Println()
		}
	}
}

// russianChunkTag ports RussianChunker.getChunkTag.
func russianChunkTag(phrase russianPhraseType, m SeqMatch, i int) string {
	atStart := i == m.Start
	switch phrase {
	case ruPhraseNP:
		if atStart {
			return "B-NP"
		}
		return "I-NP"
	case ruPhraseNPP:
		if atStart {
			return "B-NP-plural"
		}
		return "I-NP-plural"
	case ruPhraseVP:
		if atStart {
			return "B-VP"
		}
		return "I-VP"
	case ruPhraseADJP:
		if atStart {
			return "B-ADJP"
		}
		return "I-ADJP"
	case ruPhraseDPT:
		if atStart {
			return "B-DPT"
		}
		return "I-DPT"
	default:
		// NPS, PP, MayMissingYO, SBAR — Java: new ChunkTag(regex.phraseType.name())
		return phrase.name()
	}
}

func assignRussianChunksToReadings(chunkTaggedTokens []ChunkTaggedToken) {
	for _, tagged := range chunkTaggedTokens {
		if tagged.Readings == nil {
			continue
		}
		strs := make([]string, 0, len(tagged.ChunkTags))
		for _, ct := range tagged.ChunkTags {
			if s := ct.GetChunkTag(); s != "" {
				strs = append(strs, s)
			}
		}
		tagged.Readings.SetChunkTags(strs)
	}
}

func russianChunkDebugString(tokens []ChunkTaggedToken) string {
	if !russianChunkerDebug {
		return ""
	}
	var b strings.Builder
	for _, token := range tokens {
		b.WriteString("  ")
		b.WriteString(token.String())
		b.WriteString(" -- ")
		if token.Readings != nil {
			b.WriteString(token.Readings.String())
		}
		b.WriteByte('\n')
	}
	return b.String()
}

var _ Chunker = (*RussianChunker)(nil)
