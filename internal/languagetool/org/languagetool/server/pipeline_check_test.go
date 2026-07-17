package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPipeline_Check(t *testing.T) {
	p := NewPipeline(NewPipelineSettings("en-US", "anon"))
	// a/an
	m := p.Check("This is an test.")
	require.NotEmpty(t, m)
	// multi space
	require.NotEmpty(t, p.Check("hello  world"))
	// disable a/an
	require.NoError(t, p.DisableRuleID("EN_A_VS_AN"))
	// still may match other rules on "an test" - only a/an disabled
	// "this this" word repeat still works
	require.NotEmpty(t, p.Check("this this"))

	// freeze blocks further disable
	p.SetupFinished()
	require.Error(t, p.DisableRuleID("WORD_REPEAT_RULE"))
}

func TestPipeline_CheckGerman(t *testing.T) {
	// RegisterCoreRules for de uses shared + base word-repeat (not DE-specific pack)
	p := NewPipeline(NewPipelineSettings("de", "u"))
	require.NotEmpty(t, p.Check("Hallo  Welt"))
	require.NotEmpty(t, p.Check("Ein Test Test."))
}
