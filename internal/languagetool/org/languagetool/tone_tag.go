package languagetool

// ToneTag ports org.languagetool.ToneTag.
type ToneTag string

const (
	ToneClarity                ToneTag = "clarity"
	ToneFormal                 ToneTag = "formal"
	ToneProfessional           ToneTag = "professional"
	ToneConfident              ToneTag = "confident"
	ToneAcademic               ToneTag = "academic"
	TonePovRem                 ToneTag = "povrem"
	ToneScientific             ToneTag = "scientific"
	ToneObjective              ToneTag = "objective"
	TonePersuasive             ToneTag = "persuasive"
	ToneInformal               ToneTag = "informal"
	TonePovAdd                 ToneTag = "povadd"
	TonePositive               ToneTag = "positive"
	ToneGeneral                ToneTag = "general"
	ToneNoToneRule             ToneTag = "NO_TONE_RULE"
	ToneAllToneRules           ToneTag = "ALL_TONE_RULES"
	ToneAllWithoutGoalSpecific ToneTag = "ALL_WITHOUT_GOAL_SPECIFIC"
)

// RealToneTags ports ToneTag.REAL_TONE_TAGS (excludes meta tags).
func RealToneTags() []ToneTag {
	var out []ToneTag
	for _, t := range []ToneTag{
		ToneClarity, ToneFormal, ToneProfessional, ToneConfident, ToneAcademic,
		TonePovRem, ToneScientific, ToneObjective, TonePersuasive, ToneInformal,
		TonePovAdd, TonePositive, ToneGeneral,
	} {
		out = append(out, t)
	}
	return out
}
