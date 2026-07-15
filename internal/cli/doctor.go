package cli

import (
	"fmt"

	"github.com/lucasew/lang/internal/engine"
	"github.com/lucasew/lang/internal/pipeline"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Show data path, languages, and pipeline stage status (dev helper)",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := engine.New(dataDirFlag(cmd))
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "data_root\t%s\n", c.DataRoot())
			fmt.Fprintf(out, "languages\t%d\n", len(c.Languages()))
			fmt.Fprintln(out, "pipeline_stages:")
			implemented := map[string]string{
				pipeline.StageSentenceSplit: "partial (single-span text for text-level rules)",
				pipeline.StageTokenize:      "partial (whitespace-aware tokenizer)",
				pipeline.StageTag:           "not yet",
				pipeline.StageDisambiguate:  "not yet",
				pipeline.StageRules:         "partial (WHITESPACE_RULE)",
				pipeline.StageFilters:       "not yet",
				pipeline.StageSuggestions:   "partial (via rules)",
			}
			for _, s := range pipeline.AllStages {
				fmt.Fprintf(out, "  %s\t%s\n", s, implemented[s])
			}
			return nil
		},
	}
}
