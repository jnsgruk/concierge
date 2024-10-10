package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	version string = "dev"
	commit  string = "dev"
)

var shortDesc string = "A utility for configuring dev/test machines for charm development."
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

func init() {
	flags := rootCmd.PersistentFlags()
	flags.BoolP("verbose", "v", false, "enable verbose logging")
}

var rootCmd = &cobra.Command{
	Use:           "concierge",
	Version:       fmt.Sprintf("%s (%s)", version, commit),
	Short:         shortDesc,
	Long:          longDesc,
	SilenceErrors: true,
	SilenceUsage:  true,
	PreRun: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		verbose, _ := flags.GetBool("verbose")
		setupLogging(verbose)
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		slog.Error("Failed to configure machine", "error", err.Error())
		os.Exit(1)
	}
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
