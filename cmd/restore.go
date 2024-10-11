package cmd

import (
	"fmt"

	"github.com/jnsgruk/concierge/internal/concierge"
	"github.com/jnsgruk/concierge/internal/config"
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
		flags := cmd.Flags()

		conf, err := config.NewConfig(cmd, flags)
		if err != nil {
			return fmt.Errorf("failed to configure concierge: %w", err)
		}

		mgr := concierge.NewManager(conf)
		return mgr.Restore()
	},
}
