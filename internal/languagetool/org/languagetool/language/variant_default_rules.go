package language

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

func init() {
	// Wire without j_language_tool importing language package.
	languagetool.VariantDefaultRulesHook = VariantDefaultRulesForCode
}

// Default enabled/disabled rule IDs for language variants.
// Ports Language.getDefaultEnabledRulesForVariant / getDefaultDisabledRulesForVariant.

// VariantDefaultRulesForCode returns enabled/disabled default rule IDs for a language
// short code with country/variant (e.g. ca-ES-valencia, fr-BE, es-AR).
// Unknown codes → nil, nil (Language.java empty lists).
func VariantDefaultRulesForCode(code string) (enabled, disabled []string) {
	switch code {
	case ValencianCatalan.ShortCode:
		return ValencianCatalan.GetDefaultEnabledRulesForVariant(), ValencianCatalan.GetDefaultDisabledRulesForVariant()
	case BalearicCatalan.ShortCode:
		return BalearicCatalan.GetDefaultEnabledRulesForVariant(), BalearicCatalan.GetDefaultDisabledRulesForVariant()
	case BelgianFrench.ShortCode:
		return BelgianFrench.GetDefaultEnabledRulesForVariant(), BelgianFrench.GetDefaultDisabledRulesForVariant()
	case SwissFrench.ShortCode:
		return SwissFrench.GetDefaultEnabledRulesForVariant(), SwissFrench.GetDefaultDisabledRulesForVariant()
	case CanadianFrench.ShortCode:
		return CanadianFrench.GetDefaultEnabledRulesForVariant(), CanadianFrench.GetDefaultDisabledRulesForVariant()
	case SpanishVoseo.ShortCode:
		return SpanishVoseo.GetDefaultEnabledRulesForVariant(), SpanishVoseo.GetDefaultDisabledRulesForVariant()
	default:
		return nil, nil
	}
}

// GetDefaultEnabledRulesForVariant ports Catalan variants.
// ValencianCatalan and BalearicCatalan override; base Catalan → empty.
func (v CatalanVariant) GetDefaultEnabledRulesForVariant() []string {
	if v.Valencian {
		// ValencianCatalan.java
		return []string{
			"EXIGEIX_VERBS_VALENCIANS",
			"EXIGEIX_ACCENTUACIO_VALENCIANA",
			"EXIGEIX_POSSESSIUS_U",
			"EXIGEIX_VERBS_EIX",
			"EXIGEIX_VERBS_ISC",
			"PER_PER_A_INFINITIU",
			"FINS_EL_AVL",
			"LES_HA_FETES",
		}
	}
	if v.Balearic {
		// BalearicCatalan.java
		return []string{"EXIGEIX_VERBS_BALEARS"}
	}
	return nil
}

// GetDefaultDisabledRulesForVariant ports Catalan variants.
func (v CatalanVariant) GetDefaultDisabledRulesForVariant() []string {
	if v.Valencian {
		// ValencianCatalan.java — "Important: Java rules are not disabled here"
		return []string{
			"EXIGEIX_VERBS_CENTRAL",
			"EXIGEIX_ACCENTUACIO_GENERAL",
			"EXIGEIX_POSSESSIUS_V",
			"EVITA_PRONOMS_VALENCIANS",
			"EVITA_DEMOSTRATIUS_EIXE",
			"VOCABULARI_VALENCIA",
			"EXIGEIX_US",
			"FINS_EL_GENERAL",
			"EVITA_INFINITIUS_INDRE",
			"EVITA_DEMOSTRATIUS_ESTE",
			"CASTIC_CASTIG",
		}
	}
	if v.Balearic {
		// BalearicCatalan.java
		return []string{"EXIGEIX_VERBS_CENTRAL", "CA_SIMPLE_REPLACE_BALEARIC"}
	}
	return nil
}

// GetDefaultEnabledRulesForVariant ports French — only disabled overrides on BE/CH/CA.
func (v FrenchVariant) GetDefaultEnabledRulesForVariant() []string { return nil }

// GetDefaultDisabledRulesForVariant ports BelgianFrench/SwissFrench/CanadianFrench.
// All three disable DOUBLER_UNE_CLASSE; FrenchFrance has no override.
func (v FrenchVariant) GetDefaultDisabledRulesForVariant() []string {
	switch v.ShortCode {
	case "fr-BE", "fr-CH", "fr-CA":
		return []string{"DOUBLER_UNE_CLASSE"}
	default:
		return nil
	}
}

// GetDefaultEnabledRulesForVariant ports Spanish — no enabled overrides.
func (v SpanishVariant) GetDefaultEnabledRulesForVariant() []string { return nil }

// GetDefaultDisabledRulesForVariant ports SpanishVoseo → disables VOSEO.
func (v SpanishVariant) GetDefaultDisabledRulesForVariant() []string {
	if v.Voseo {
		return []string{"VOSEO"}
	}
	return nil
}
