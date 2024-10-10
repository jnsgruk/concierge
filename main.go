package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/jnsgruk/concierge/internal/concierge"
	"github.com/jnsgruk/concierge/internal/config"
	"github.com/spf13/cobra"
)

var (
	version string = "dev"
	commit  string = "dev"
)

var shortDesc = "A utility for configuring dev/test machines for charm development."
var longDesc string = `concierge is an opinionated utility for provisioning charm development and testing machines.

It's role is to ensure that a given machine has the relevant "craft" tools and providers installed,
then bootstrap a Juju controller onto each of the providers. Additionally, it can install selected
tools from the [snap store](https://snapcraft.io) or the Ubuntu archive.

Configuration is by flags/environment variables, or by configuration file. The configuration file
must be in the current working directory and named 'concierge.yaml', or the path specified using
the '-c' flag.

There are 3 presets available by default: 'machine', 'k8s' and 'dev'.

Some aspects of presets and config files can be overridden using flags such as '--juju-channel'.
Each of the override flags has an environment variable equivalent, 
such as 'CONCIERGE_JUJU_CHANNEL'.

More information at https://github.com/jnsgruk/concierge.
`

var rootCmd = &cobra.Command{
	Use:           "concierge",
	Version:       fmt.Sprintf("%s (%s)", version, commit),
	Short:         shortDesc,
	Long:          longDesc,
	SilenceErrors: true,
	SilenceUsage:  true,

	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.PersistentFlags()
		verbose, _ := flags.GetBool("verbose")
		configFile, _ := flags.GetString("config")
		preset, _ := flags.GetString("preset")

		// Ensure the slog logger is set for the correct format/log level
		setupLogging(verbose)

		// Concierge cannot merge a preset & manual configuration
		if len(preset) > 0 && len(configFile) > 0 {
			return fmt.Errorf("cannot proceed with both preset and configuration file specified")
		}

		conf, err := config.NewConfig(cmd, flags)
		if err != nil {
			return fmt.Errorf("failed to configure concierge: %w", err)
		}

		mgr := concierge.NewManager(conf)

		return mgr.Execute()
	},
}

func setupLogging(verbose bool) {
	logLevel := new(slog.LevelVar)

	// Set the default log level to "INFO", and "DEBUG" if verbose is specified.
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	// Setup the TextHandler and ensure our configured logger is the default.
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	logger := slog.New(h)
	slog.SetDefault(logger)
	logLevel.Set(level)
}

func init() {
	flags := rootCmd.PersistentFlags()
	flags.BoolP("verbose", "v", false, "enable verbose logging")
	flags.StringP("config", "c", "", "path to a specific config file to use")
	flags.StringP("preset", "p", "", "config preset to use (k8s | machine | dev)")

	// Version overrides
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
		"comma-separated list of extra snaps to install. Each item can simply be the name of a snap, but also include the channel. E.g. 'astral-uv/latest/edge,jhack'",
	)

	flags.StringSlice(
		"extra-debs",
		[]string{},
		"comma-separated list of extra debs to install. E.g. 'make,python3-tox'",
	)
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		slog.Error("Failed to configure machine", "error", err.Error())
		os.Exit(1)
	}
}
