package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	version string = "dev"
	commit  string = "dev"
)

var rootLongDesc string = `concierge is an opinionated utility for provisioning charm development and testing machines.

It's role is to ensure that a given machine has the relevant "craft" tools and providers installed,
then bootstrap a Juju controller onto each of the providers. Additionally, it can install selected
tools from the [snap store](https://snapcraft.io) or the Ubuntu archive.
`

func init() {
	flags := rootCmd.PersistentFlags()
	flags.BoolP("verbose", "v", false, "enable verbose logging")
	flags.Bool("trace", false, "enable trace logging")
}

var rootCmd = &cobra.Command{
	Use:           "concierge",
	Version:       fmt.Sprintf("%s (%s)", version, commit),
	Short:         "A utility for configuring dev/test machines for charm development.",
	Long:          rootLongDesc,
	SilenceErrors: true,
	SilenceUsage:  true,
	PreRun: func(cmd *cobra.Command, args []string) {
		parseLoggingFlags(cmd.Flags())
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute runs the root command and exits the program if it fails.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		slog.Error("Failed to configure machine", "error", err.Error())
		os.Exit(1)
	}
}

func parseLoggingFlags(flags *pflag.FlagSet) {
	verbose, _ := flags.GetBool("verbose")
	trace, _ := flags.GetBool("trace")

	logLevel := new(slog.LevelVar)

	// Set the default log level to "DEBUG" if verbose is specified.
	level := slog.LevelInfo
	if !verbose && trace {
		level = slog.LevelDebug
	}

	// Setup the TextHandler and ensure our configured logger is the default.
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	logger := slog.New(h)
	slog.SetDefault(logger)
	logLevel.Set(level)
}
