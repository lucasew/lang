package chunking

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// RussianChunker ports org.languagetool.chunking.RussianChunker (@Experimental).
// REGEXES1/REGEXES2 run Java OpenRegex patterns via CompileOpenRegex — not invent POS→BIO.
type RussianChunker struct{}

func NewRussianChunker() *RussianChunker { return &RussianChunker{} }

// Java FILTER_TAGS for overwrite mode.
var russianFilterTags = map[string]bool{
	"PP": true, "NPP": true, "NPS": true, "MayMissingYO": true,
	"VP": true, "SBAR": true, "ADJP": true, "DPT": true,
	// also B/I forms removed? Java FILTER_TAGS is exact names above only.
}

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

type russianRegex struct {
	pattern   string
	phrase    russianPhraseType
	overwrite bool
}

// Java RussianChunker.REGEXES1 (order + overwrite).
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

// Java RussianChunker.REGEXES2
var russianRegexes2 = []russianRegex{
	{`<posre=NN:Name:.*> <и> <posre=NN:Name:.*>`, ruPhraseNPP, true},
	{`<posre=NN:Name:.*> <или> <posre=NN:Name:.*>`, ruPhraseNPP, true},
	{`<не> <posre='VB:.*:.*' & !posre='NN:.*'>* `, ruPhraseVP, false},
}

func (c *RussianChunker) AddChunkTags(tokens []*languagetool.AnalyzedTokenReadings) {
	if c == nil || len(tokens) == 0 {
		return
	}
	toks := getRussianBasicChunks(tokens)
	if len(toks) == 0 {
		return
	}
	factory := NewChunkTokenFactory(false)
	for _, spec := range russianRegexes2 {
		applyRussianRegex(spec, toks, factory)
	}
	assignRussianChunksToReadings(toks)
}

// GetBasicChunks ports RussianChunker.getBasicChunks (REGEXES1 only).
func (c *RussianChunker) GetBasicChunks(tokens []*languagetool.AnalyzedTokenReadings) []ChunkTaggedToken {
	return getRussianBasicChunks(tokens)
}

func getRussianBasicChunks(tokenReadings []*languagetool.AnalyzedTokenReadings) []ChunkTaggedToken {
	var chunkTaggedTokens []ChunkTaggedToken
	for _, tokenReading := range tokenReadings {
		if tokenReading == nil || tokenReading.IsWhitespace() {
			continue
		}
		// Java: skip tokens that already have MayMissingYO chunk tag
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
	factory := NewChunkTokenFactory(false)
	for _, spec := range russianRegexes1 {
		applyRussianRegex(spec, chunkTaggedTokens, factory)
	}
	return chunkTaggedTokens
}

func applyRussianRegex(spec russianRegex, tokens []ChunkTaggedToken, factory func(string) func(ChunkTaggedToken) bool) {
	pat := ExpandRussianChunkSyntax(spec.pattern)
	re := CompileOpenRegex(pat, factory)
	for _, m := range re.FindAll(tokens) {
		for i := m.Start; i < m.End; i++ {
			newTags := make([]ChunkTag, 0, len(tokens[i].ChunkTags)+1)
			for _, ct := range tokens[i].ChunkTags {
				s := ct.GetChunkTag()
				if spec.overwrite && russianFilterTags[s] {
					continue
				}
				// overwrite also drops B-NP/I-NP etc.? Java only FILTER_TAGS set above.
				newTags = append(newTags, ct)
			}
			newTag := russianChunkTag(spec.phrase, m, i)
			has := false
			for _, ct := range newTags {
				if ct.GetChunkTag() == newTag {
					has = true
					break
				}
			}
			if !has {
				newTags = append(newTags, NewChunkTag(newTag))
			}
			filtered := make([]ChunkTag, 0, len(newTags))
			for _, ct := range newTags {
				if ct.GetChunkTag() != "O" {
					filtered = append(filtered, ct)
				}
			}
			tokens[i] = NewChunkTaggedToken(tokens[i].Token, filtered, tokens[i].Readings)
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
	case ruPhraseNPS:
		return "NPS"
	case ruPhrasePP:
		return "PP"
	case ruPhraseMayMissingYO:
		return "MayMissingYO"
	case ruPhraseSBAR:
		return "SBAR"
	default:
		return ""
	}
}

func assignRussianChunksToReadings(chunkTaggedTokens []ChunkTaggedToken) {
	for _, tagged := range chunkTaggedTokens {
		if tagged.Readings == nil {
			continue
		}
		var strs []string
		for _, ct := range tagged.ChunkTags {
			if s := ct.GetChunkTag(); s != "" {
				strs = append(strs, s)
			}
		}
		tagged.Readings.SetChunkTags(strs)
	}
}

var _ Chunker = (*RussianChunker)(nil)
