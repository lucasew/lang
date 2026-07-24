package languagetool

// LanguageMaintainedState ports org.languagetool.LanguageMaintainedState.
type LanguageMaintainedState string

const (
	ActivelyMaintained      LanguageMaintainedState = "ActivelyMaintained"
	LookingForNewMaintainer LanguageMaintainedState = "LookingForNewMaintainer"
)
