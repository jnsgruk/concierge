package cmd

import (
	"github.com/jnsgruk/concierge/internal/concierge"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(restoreCmd)
}

// restoreCmd represents the restore subcommand
var restoreCmd = &cobra.Command{
	Use:           "restore",
	Short:         "Restore the machine to it's pre-provisioned state.",
	Long:          "Restore the machine to it's pre-provisioned state.",
	SilenceErrors: true,
	SilenceUsage:  true,
	PreRun: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		verbose, _ := flags.GetBool("verbose")
		setupLogging(verbose)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := &concierge.Manager{}
		return mgr.Restore()
	},
}
