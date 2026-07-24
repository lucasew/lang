package de

// Twin of CompoundInfinitivRuleTest — Java uses ZUS + VER:INF + speller.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCompoundInfinitivRule_Rule(t *testing.T) {
	// Speller: joined particle+infinitive is a known compound (Java !isMisspelled)
	// Also "anfangen" etc. for exception scan.
	known := map[string]struct{}{
		"saubermachen": {},
		"vorbeikommen": {},
		"vorbeilassen": {},
		"anfangen":     {},
		"aufhören":     {}, // if scanned as particle+verb
	}
	rule := NewCompoundInfinitivRule(nil)
	rule.IsMisspelled = func(w string) bool {
		_, ok := known[w]
		return !ok
	}

	// Java: "Ich brachte ihn dazu, mein Zimmer sauber zu machen." → 1
	sauber := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:MAS", "ich"),
		atrWithPOS("brachte", "VER:3:SIN:PRT:SFT", "bringen"),
		atrWithPOS("ihn", "PRO:PER:AKK:SIN:MAS", "er"),
		atrWithPOS("dazu", "ADV", "dazu"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("mein", "PRO:POS:AKK:SIN:NEU", "mein"),
		atrWithPOS("Zimmer", "SUB:AKK:SIN:NEU", "Zimmer"),
		atrWithPOS("sauber", "ZUS", "sauber"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("machen", "VER:INF:NON", "machen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(sauber)))

	// Java: "Du brauchst nicht bei mir vorbei zu kommen." → 1
	vorbei := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Du", "PRO:PER:NOM:SIN:MAS", "du"),
		atrWithPOS("brauchst", "VER:2:SIN:PRS:SFT", "brauchen"),
		atrWithPOS("nicht", "ADV", "nicht"),
		atrWithPOS("bei", "PRP:DAT", "bei"),
		atrWithPOS("mir", "PRO:PER:DAT:SIN:MAS", "ich"),
		atrWithPOS("vorbei", "ZUS", "vorbei"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("kommen", "VER:INF:NON", "kommen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(vorbei)))

	// Java: "… die alte Dame vorbei zu lassen." → 1
	lassen := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:MAS", "ich"),
		atrWithPOS("ging", "VER:3:SIN:PRT:NON", "gehen"),
		atrWithPOS("zur", "APPRART:DAT:FEM", "zu"),
		atrWithPOS("Seite", "SUB:DAT:SIN:FEM", "Seite"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("um", "KOUI", "um"),
		atrWithPOS("die", "ART:DEF:AKK:SIN:FEM", "die"),
		atrWithPOS("alte", "ADJ:AKK:SIN:FEM:GRU:DEF", "alt"),
		atrWithPOS("Dame", "SUB:AKK:SIN:FEM", "Dame"),
		atrWithPOS("vorbei", "ZUS", "vorbei"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("lassen", "VER:INF:NON", "lassen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(lassen)))

	// Java: "Seine Frau gab vor zu schlafen." → 0 (isException: tokens[n-2] VER)
	// "gab" VER, "vor" ZUS, "zu", "schlafen"
	vor := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Seine", "PRO:POS:NOM:SIN:FEM", "sein"),
		atrWithPOS("Frau", "SUB:NOM:SIN:FEM", "Frau"),
		atrWithPOS("gab", "VER:3:SIN:PRT:NON", "geben"),
		atrWithPOS("vor", "ZUS", "vor"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("schlafen", "VER:INF:NON", "schlafen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(vor)))

	// Java: "Mein Herz hörte auf zu schlagen." → 0 (VER before particle)
	auf := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Mein", "PRO:POS:NOM:SIN:NEU", "mein"),
		atrWithPOS("Herz", "SUB:NOM:SIN:NEU", "Herz"),
		atrWithPOS("hörte", "VER:3:SIN:PRT:SFT", "hören"),
		atrWithPOS("auf", "ZUS", "auf"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("schlagen", "VER:INF:NON", "schlagen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(auf)))

	// Java: "Fang an zu zählen." → 0 via isException verb scan (an+fangen known)
	fang := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Fang", "VER:IMP:SIN:SFT", "fangen"),
		atrWithPOS("an", "ZUS", "an"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("zählen", "VER:INF:NON", "zählen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(fang)))

	// Java: "Er hatte nichts weiter zu sagen" → 0 (weiter+sagen adj/special exception)
	// "weiter" is in ADJ_EXCEPTION? Looking Java: sagen+weiter exception
	weiter := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("hatte", "VER:3:SIN:PRT:NON", "haben"),
		atrWithPOS("nichts", "PIS", "nichts"),
		atrWithPOS("weiter", "ZUS", "weiter"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("sagen", "VER:INF:NON", "sagen"),
		atrWithPOS(".", "PKT", "."),
	))
	// n-2 is "nichts" not VER — but sagen+weiter exception
	require.Equal(t, 0, len(rule.Match(weiter)))

	// Java: "Hör auf zu schreien" → 0 (tokens[n-2] VER before particle)
	hoer := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Hör", "VER:IMP:SIN:SFT", "hören"),
		atrWithPOS("auf", "ZUS", "auf"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("schreien", "VER:INF:NON", "schreien"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(hoer)))

	// Java: "Den Sonnenaufgang von einem Berggipfel aus zu sehen, …" → 0 (von…aus exception)
	// also know aussehen as compound so without exception would fire
	known["aussehen"] = struct{}{}
	ausSehen := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Den", "ART:DEF:AKK:SIN:MAS", "der"),
		atrWithPOS("Sonnenaufgang", "SUB:AKK:SIN:MAS", "Sonnenaufgang"),
		atrWithPOS("von", "APPR:DAT", "von"),
		atrWithPOS("einem", "ART:IND:DAT:SIN:MAS", "ein"),
		atrWithPOS("Berggipfel", "SUB:DAT:SIN:MAS", "Berggipfel"),
		atrWithPOS("aus", "ZUS", "aus"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("sehen", "VER:INF:NON", "sehen"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("eine", "ART:IND:NOM:SIN:FEM", "ein"),
		atrWithPOS("Wonne", "SUB:NOM:SIN:FEM", "Wonne"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(ausSehen)))

	// Java: "Sie strengte sich an zu schwimmen." → 0 (an+strengen known via exception scan)
	known["anstrengen"] = struct{}{}
	anSchwimmen := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Sie", "PRO:PER:NOM:SIN:FEM", "sie"),
		atrWithPOS("strengte", "VER:3:SIN:PRT:SFT", "strengen"),
		atrWithPOS("sich", "PRF:AKK:SIN", "sich"),
		atrWithPOS("an", "ZUS", "an"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("schwimmen", "VER:INF:NON", "schwimmen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(anSchwimmen)))

	// Java: "Sie riss sich zusammen und fing wieder an zu reden." → 0 (an+fangen)
	// exception scan: "fing" VER before particle path also via known "anfangen"
	known["anfangen"] = struct{}{}
	anReden := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Sie", "PRO:PER:NOM:SIN:FEM", "sie"),
		atrWithPOS("riss", "VER:3:SIN:PRT:NON", "reißen"),
		atrWithPOS("sich", "PRF:AKK:SIN", "sich"),
		atrWithPOS("zusammen", "PTKVZ", "zusammen"),
		atrWithPOS("und", "KON", "und"),
		atrWithPOS("fing", "VER:3:SIN:PRT:NON", "fangen"),
		atrWithPOS("wieder", "ADV", "wieder"),
		atrWithPOS("an", "ZUS", "an"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("reden", "VER:INF:NON", "reden"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(anReden)))

	// Java: "Fang dort an zu lesen, wo du aufgehört hast." → 0
	fangDort := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Fang", "VER:IMP:SIN:SFT", "fangen"),
		atrWithPOS("dort", "ADV", "dort"),
		atrWithPOS("an", "ZUS", "an"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("lesen", "VER:INF:NON", "lesen"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("wo", "PWAV", "wo"),
		atrWithPOS("du", "PRO:PER:NOM:SIN:MAS", "du"),
		atrWithPOS("aufgehört", "VER:PA2:SFT", "aufhören"),
		atrWithPOS("hast", "VER:2:SIN:PRÄ:NON", "haben"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(fangDort)))

	// Java: "Der diensthabende Kollege hatte ganz schön zu tun." → 0
	// "schön" is ADJ not ZUS → isRelevant false
	schoen := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("diensthabende", "ADJA:NOM:SIN:MAS:GRU:DEF", "diensthabend"),
		atrWithPOS("Kollege", "SUB:NOM:SIN:MAS", "Kollege"),
		atrWithPOS("hatte", "VER:3:SIN:PRT:NON", "haben"),
		atrWithPOS("ganz", "ADV", "ganz"),
		atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("tun", "VER:INF:NON", "tun"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(schoen)))

	// Java: "Hab keine Lust, mir Gedanken darüber zu machen." → 0 ("darüber" not ZUS)
	darueber := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Hab", "VER:IMP:SIN:NON", "haben"),
		atrWithPOS("keine", "PIAT:AKK:SIN:FEM", "kein"),
		atrWithPOS("Lust", "SUB:AKK:SIN:FEM", "Lust"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("mir", "PRO:PER:DAT:SIN:MAS", "ich"),
		atrWithPOS("Gedanken", "SUB:AKK:PLU:MAS", "Gedanke"),
		atrWithPOS("darüber", "PRO:ADV", "darüber"),
		atrWithPOS("zu", "PTKZU", "zu"),
		atrWithPOS("machen", "VER:INF:NON", "machen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(darueber)))

	// Error match: message + suggestion morph (sauber zu machen → saubermachen)
	ms := rule.Match(sauber)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetMessage(), "saubermachen")
	// Java: setSuggestedReplacement(particle + "zu" + infinitive) e.g. sauberzumachen
	require.Equal(t, []string{"sauberzumachen"}, ms[0].GetSuggestedReplacements())
	// UTF-16 span: particle start → infinitive end
	toks := sauber.GetTokensWithoutWhitespace()
	var sauberTok, machenTok *languagetool.AnalyzedTokenReadings
	for _, tok := range toks {
		if tok != nil && tok.GetToken() == "sauber" {
			sauberTok = tok
		}
		if tok != nil && tok.GetToken() == "machen" {
			machenTok = tok
		}
	}
	require.NotNil(t, sauberTok)
	require.NotNil(t, machenTok)
	require.Equal(t, sauberTok.GetStartPos(), ms[0].GetFromPos())
	require.Equal(t, machenTok.GetEndPos(), ms[0].GetToPos())

	// untagged must not invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Du brauchst nicht bei mir vorbei zu kommen."))))

	// Without speller, fail closed (all joins misspelled)
	noSpell := NewCompoundInfinitivRule(nil)
	require.Equal(t, 0, len(noSpell.Match(sauber)))
}

func TestCompoundInfinitivRule_IsPunctuationUTF16(t *testing.T) {
	require.True(t, isPunctuationCI("…"))
	require.True(t, isPunctuationCI("."))
	require.False(t, isPunctuationCI(".."))
	require.False(t, isPunctuationCI(""))
}

func TestCompoundInfinitivRule_Meta(t *testing.T) {
	r := NewCompoundInfinitivRule(nil)
	require.Equal(t, "COMPOUND_INFINITIV_RULE", r.GetID())
	require.Equal(t, "Erweiterter Infinitiv mit zu (Zusammenschreibung)", r.GetDescription())
	require.Equal(t, "https://languagetool.org/insights/de/beitrag/zu-zusammen-oder-getrennt/", r.GetURL())
	require.NotEmpty(t, r.GetIncorrectExamples())
	require.NotEmpty(t, r.GetCorrectExamples())
}
