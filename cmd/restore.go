package cmd

import (
	"fmt"

	"github.com/jnsgruk/concierge/internal/concierge"
	"github.com/jnsgruk/concierge/internal/config"
	"github.com/spf13/cobra"
)

// restoreCmd constructs the `restore` subcommand
func restoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "restore",
		Short:         "Restore the machine to it's pre-provisioned state.",
		Long:          "Restore the machine to it's pre-provisioned state.",
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun: func(cmd *cobra.Command, args []string) {
			parseLoggingFlags(cmd.Flags())
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
}
