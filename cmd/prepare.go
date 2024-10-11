package cmd

import (
	"fmt"

	"github.com/jnsgruk/concierge/internal/concierge"
	"github.com/jnsgruk/concierge/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	flags := prepareCmd.Flags()
	flags.StringP("config", "c", "", "path to a specific config file to use")
	flags.StringP("preset", "p", "", "config preset to use (k8s | machine | dev)")
	flags.String("juju-channel", "", "override the snap channel for juju")
	flags.String("microk8s-channel", "", "override snap channel for microk8s")
	flags.String("lxd-channel", "", "override snap channel for lxd")
	flags.String("charmcraft-channel", "", "override snap channel for charmcraft")
	flags.String("snapcraft-channel", "", "override snap channel for snapcraft")
	flags.String("rockcraft-channel", "", "override snap channel for rockcraft")

	// Additional package specification
	flags.StringSlice(
		"extra-snaps",
		[]string{},
		"comma-separated list of extra snaps to install. E.g. 'astral-uv/latest/edge,jhack'",
	)

	flags.StringSlice(
		"extra-debs",
		[]string{},
		"comma-separated list of extra debs to install. E.g. 'make,python3-tox'",
	)

	rootCmd.AddCommand(prepareCmd)
}

// prepareCmd represents the restore subcommand
var prepareCmd = &cobra.Command{
	Use:           "prepare",
	Short:         "Provision the machine according to the configuration.",
	Long:          "Provision the machine according to the configuration.",
	SilenceErrors: true,
	SilenceUsage:  true,
	PreRun: func(cmd *cobra.Command, args []string) {
		parseLoggingFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()

		configFile, _ := flags.GetString("config")
		preset, _ := flags.GetString("preset")

		// Concierge cannot merge a preset & manual configuration
		if len(preset) > 0 && len(configFile) > 0 {
			return fmt.Errorf("cannot proceed with both preset and configuration file specified")
		}

		conf, err := config.NewConfig(cmd, flags)
		if err != nil {
			return fmt.Errorf("failed to configure concierge: %w", err)
		}

		mgr, err := concierge.NewManager(conf), nil

		return mgr.Prepare()
	},
}
