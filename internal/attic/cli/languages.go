package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/lucasew/lang/internal/attic/engine"
	"github.com/spf13/cobra"
)

func newLanguagesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "languages",
		Short: "List languages discovered in the LanguageTool data tree",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := engine.New(dataDirFlag(cmd))
			if err != nil {
				return err
			}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "code\tfamily\tname\tclass")
			for _, l := range c.Languages() {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", l.Code, l.Family, l.Name, l.JavaClass)
			}
			return tw.Flush()
		},
	}
}
