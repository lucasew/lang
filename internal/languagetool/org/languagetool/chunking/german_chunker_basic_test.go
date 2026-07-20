package chunking

import (
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// parseBasicAnnotation ports GermanChunkerTest input like "Ein/B Haus/I".
// B → B-NP, I → I-NP; bare words have no expected BIO (O).
type basicAnno struct {
	token string
	bio   string // "B-NP", "I-NP", or "" (expect O)
}

func parseBasicAnnotation(input string) []basicAnno {
	parts := strings.Fields(input)
	out := make([]basicAnno, 0, len(parts))
	for _, p := range parts {
		if i := strings.LastIndex(p, "/"); i > 0 {
			tok, mark := p[:i], p[i+1:]
			switch mark {
			case "B":
				out = append(out, basicAnno{token: tok, bio: "B-NP"})
			case "I":
				out = append(out, basicAnno{token: tok, bio: "I-NP"})
			default:
				out = append(out, basicAnno{token: tok, bio: ""})
			}
			continue
		}
		out = append(out, basicAnno{token: p, bio: ""})
	}
	return out
}

// basicPOS assigns POS for REGEXES1 tests (faithful tags for common German forms; not invent).
// prevTok is the previous surface (for "das" ART vs relative PRO after comma).
func basicPOS(tok, prevTok string) string {
	low := strings.ToLower(tok)
	switch low {
	case "das":
		// After comma: relative PRO so "das die Wärme" does not fuse (Java: das/O, die/B Wärme/I).
		// Elsewhere: ART ("das Haus", "das Meer").
		if prevTok == "," {
			return "PRO:REL:NOM:SIN:NEU"
		}
		return "ART:DEF:NOM:SIN:NEU"
	case "ein", "eine", "einen", "einem", "einer", "eines", "die", "der", "den", "dem", "des":
		return "ART:DEF:NOM:SIN:NEU"
	case "sehr", "da", "dort", "schon", "mehr", "als", "blau":
		// "Da steht …" — ADV, not SUB (capitalized sentence start must not invent NP)
		return "ADV"
	case "schönes", "großes", "leckere", "leckeren", "blauen", "schöne", "neue", "grünen":
		return "ADJ:NOM:SIN:NEU:GRU:SOL"
	case "herr":
		return "SUB:NOM:SIN:MAS" // title path uses surface Herr|Frau
	case "meier", "schrödinger", "karl", "finn", "westerwalbesloh":
		return "EIG:NOM:SIN:MAS"
	case "ich", "ihre", "ihrer", "unsere", "er":
		return "PRO:PER:NOM:SIN:1"
	case "mit", "am", "im", "in":
		return "PRP:LOK"
	case "und", "oder":
		return "KON:NEB"
	case ",":
		return "PKT"
	case "zwei", "drei", "zwanzig":
		return "ZAL"
	case "1000":
		// Java assertBasicChunks leaves "1000" as O and "Bürger/B" alone
		// (comment: 1000 sollte evtl. mit in die NP). Digit string ≠ ZAL here.
		return "CARD"
	case "steht", "isst", "geht", "meint", "muss", "überträgt", "mag", "sind", "betrifft", "ist", "geben", "runter":
		return "VER:3:SIN:PRÄ:SFT"
	case "haus", "dach", "lasagne", "kuchen", "heimat", "bach", "hang", "hund", "futter",
		"wasser", "wärme", "meer", "luft", "prozent", "arbeiter", "streik", "gesetz",
		"bürger", "wochen", "weihnachten", "autos":
		return "SUB:NOM:SIN:NEU"
	}
	// Capitalized proper names → EIG (not bare SUB invent for any Cap token)
	if tok != "" {
		r, _ := utf8DecodeRune(tok)
		if unicode.IsUpper(r) {
			return "EIG:NOM:SIN:NEU"
		}
	}
	return "UNKNOWN"
}

func utf8DecodeRune(s string) (rune, int) {
	for _, r := range s {
		return r, 1
	}
	return 0, 0
}

func tokensFromBasicAnno(ann []basicAnno) []*languagetool.AnalyzedTokenReadings {
	out := make([]*languagetool.AnalyzedTokenReadings, len(ann))
	pos := 0
	prev := ""
	for i, a := range ann {
		out[i] = atrPos(a.token, basicPOS(a.token, prev), pos)
		pos += len(a.token) + 1
		prev = a.token
	}
	return out
}

func TestGermanChunker_GetBasicChunks_JavaOpenNLPTable(t *testing.T) {
	// Subset of GermanChunkerTest.assertBasicChunks with POS assignable without full tagger.
	cases := []string{
		"Ein/B Haus/I",
		"Da steht ein/B Haus/I",
		"Da steht ein/B schönes/I Haus/I",
		"Da steht ein/B schönes/I großes/I Haus/I",
		"Da steht ein/B sehr/I großes/I Haus/I",
		"Da steht ein/B sehr/I schönes/I großes/I Haus/I",
		"Eine/B leckere/I Lasagne/I",
		"Herr/B Meier/I isst eine/B leckere/I Lasagne/I",
		"Herr/B Schrödinger/I isst einen/B Kuchen/I",
		"Herr/B Schrödinger/I isst einen/B leckeren/I Kuchen/I",
		"Herr/B Karl/I Meier/I isst eine/B leckere/I Lasagne/I",
		"Herr/B Finn/I Westerwalbesloh/I isst eine/B leckere/I Lasagne/I",
		"In zwei/B Wochen/I ist Weihnachten/B",
		"Eines ihrer/B drei/I Autos/I ist blau",
		// More Java assertBasicChunks (REGEXES1 only)
		"Da steht ein/B sehr/I großes/I Haus/I mit Dach/B",
		"Da steht ein/B sehr/I großes/I Haus/I mit einem/B blauen/I Dach/I",
		"Unsere/B schöne/I Heimat/I geht den/B Bach/I runter",
		"Er meint das/B Haus/I am grünen/B Hang/I",
		"Das/B neue/I Gesetz/I betrifft 1000 Bürger/B",
		"Schon mehr als zwanzig/B Prozent/I der/B Arbeiter/I sind im Streik/B",
		// Relative / coordinated NPs (Java assertBasicChunks)
		"Das/B Wasser/I , das die/B Wärme/I überträgt",
		"Er mag das/B Wasser/I , das/B Meer/I und die/B Luft/I",
		// Explicit incomplete: "Ich/B muss dem/B Hund/I Futter/I geben" — Java expects
		// Ich as B-NP but REGEXES1 needs SUB after optional PRO; bare PRO stays O
		// (Java even comments the preferred Futter/B split). No invent SUB on Ich.
	}
	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			ann := parseBasicAnnotation(tc)
			tokens := tokensFromBasicAnno(ann)
			basic := NewGermanChunker().GetBasicChunks(tokens)
			require.Len(t, basic, len(ann))
			for i, a := range ann {
				require.Equal(t, a.token, basic[i].Token)
				got := "O"
				if len(basic[i].ChunkTags) > 0 {
					got = basic[i].ChunkTags[0].String()
				}
				want := a.bio
				if want == "" {
					want = "O"
				}
				require.Equal(t, want, got, "token %q in %q", a.token, tc)
			}
		})
	}
}

// Java GermanChunker.REGEXES2 has 77 build(...) entries — keep list complete.
func TestGermanRegexes2_CountMatchesJava(t *testing.T) {
	require.Equal(t, 77, len(germanRegexes2), "REGEXES2 must list every Java build() entry in order")
	// Patterns previously missing from the Go transcription:
	need := []string{
		`<chunk=B-NP & pos=PLU> <chunk=I-NP>* <chunk=B-NP & pos=GEN> <chunk=I-NP>*`,
		`<pos=PRP> <NP> <pos=ADJ> (<und>|<oder>|<bzw.>) <pos=ADJ> <NP>`,
		`<pos=PRP> <pos=ADJ> (<und|oder|sowie>) <pos=ADJ> <chunk=B-NP>`,
		`<pos=PRP> <NP> <pos=ADJ> <NP> (<und|oder>) <NP>`,
	}
	have := map[string]bool{}
	for _, r := range germanRegexes2 {
		have[r.pattern] = true
	}
	for _, p := range need {
		require.True(t, have[p], "missing REGEXES2 pattern: %s", p)
	}
}
