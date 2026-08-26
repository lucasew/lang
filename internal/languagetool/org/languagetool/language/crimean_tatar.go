package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	crhsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/crh"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	crhtag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/crh"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	crhtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/crh"
)

// CrimeanTatarLang ports org.languagetool.language.CrimeanTatar metadata + default factories.
// Java overrides getMaintainedState → ActivelyMaintained.
// Java uses lazy getTagger / getSentenceTokenizer / getWordTokenizer fields (not only createDefault*);
// Go surfaces those as CreateDefault* factories matching other language roots.
// No createDefaultSpellingRule / createDefaultDisambiguator (Java has none — do not invent).
type CrimeanTatarLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
}

// CrimeanTatar is the default Crimean Tatar language surface (Java new CrimeanTatar()).
var CrimeanTatar = CrimeanTatarLang{
	ShortCode:     "crh",
	Name:          "Crimean Tatar",
	SpellerRuleID: "MORFOLOGIK_RULE_CRH_UA",
	Countries:     []string{"UA"},
}

// NewCrimeanTatar ports new CrimeanTatar().
func NewCrimeanTatar() CrimeanTatarLang {
	return CrimeanTatar
}

// CrimeanTatarShortCode ports CrimeanTatar.getShortCode ("crh").
func CrimeanTatarShortCode() string { return "crh" }

func (v CrimeanTatarLang) GetName() string { return v.Name }

// GetShortCode ports CrimeanTatar.getShortCode ("crh").
func (v CrimeanTatarLang) GetShortCode() string { return "crh" }

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (v CrimeanTatarLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant("crh", v.Countries, "")
}

func (v CrimeanTatarLang) GetCountries() []string {
	return append([]string(nil), v.Countries...)
}

// GetCommonWordsPath ports Language.getCommonWordsPath → crh/common_words.txt.
func (v CrimeanTatarLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(v.GetShortCode())
}

// GetDefaultSpellingRuleID is the registered MorfologikCrimeanTatarSpellerRule.getId
// (MORFOLOGIK_RULE_CRH_UA). Java CrimeanTatar does not override createDefaultSpellingRule —
// speller is only listed in getRelevantRules; ID surface matches prior SmallLang / other roots.
func (v CrimeanTatarLang) GetDefaultSpellingRuleID() string {
	if v.SpellerRuleID != "" {
		return v.SpellerRuleID
	}
	return "MORFOLOGIK_RULE_CRH_UA"
}

// GetMaintainedState ports CrimeanTatar.getMaintainedState → ActivelyMaintained.
func (v CrimeanTatarLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports CrimeanTatar.getMaintainers → Andriy Rysin.
func (v CrimeanTatarLang) GetMaintainers() []Contributor {
	return []Contributor{NewContributor("Andriy Rysin")}
}

// CreateDefaultTagger ports CrimeanTatar.getTagger → new CrimeanTatarTagger().
// Java: super("/crh/crimean_tatar.dict", Locale("uk","UA"), false) via BaseTagger.
// Dict binary not in inspiration tree (built artifact); MapWordTagger until official
// dict is discoverable — fail-closed real dict path only (same as Asturian leaf gap).
func (v CrimeanTatarLang) CreateDefaultTagger() *crhtag.CrimeanTatarTagger {
	return crhtag.NewCrimeanTatarTagger(tagging.MapWordTagger{})
}

// CreateDefaultSentenceTokenizer ports CrimeanTatar.getSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "crh".
func (v CrimeanTatarLang) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports CrimeanTatar.getWordTokenizer
// → new CrimeanTatarWordTokenizer().
func (v CrimeanTatarLang) CreateDefaultWordTokenizer() *crhtok.CrimeanTatarWordTokenizer {
	return crhtok.NewCrimeanTatarWordTokenizer()
}

// CreateDefaultSynthesizer ports CrimeanTatar.createDefaultSynthesizer
// → CrimeanTatarSynthesizer.INSTANCE.
func (v CrimeanTatarLang) CreateDefaultSynthesizer() *crhsynth.CrimeanTatarSynthesizer {
	return crhsynth.Instance()
}

// GetRuleFileNames ports Language.getRuleFileNames default for "crh"
// (Java CrimeanTatar does not add RULE_FILES extras).
func (v CrimeanTatarLang) GetRuleFileNames() []string {
	return v.GetRuleFileNamesWithExists(nil)
}

// GetRuleFileNamesWithExists allows probing style/variant file existence
// (same pluggable exists as SmallLang / Language.getRuleFileNames).
func (v CrimeanTatarLang) GetRuleFileNamesWithExists(exists languagetool.RuleFileExists) []string {
	code := v.GetShortCode()
	return languagetool.GetRuleFileNames(code, code, "/org/languagetool/rules", exists)
}
