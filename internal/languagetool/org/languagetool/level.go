package languagetool

// Level ports JLanguageTool.Level.
type Level string

const (
	LevelDefault      Level = "DEFAULT"
	LevelPicky        Level = "PICKY"
	LevelAcademic     Level = "ACADEMIC"
	LevelClarity      Level = "CLARITY"
	LevelProfessional Level = "PROFESSIONAL"
	LevelCreative     Level = "CREATIVE"
	LevelCustomer     Level = "CUSTOMER"
	LevelJobApp       Level = "JOBAPP"
	LevelObjective    Level = "OBJECTIVE"
	LevelElegant      Level = "ELEGANT"
)
