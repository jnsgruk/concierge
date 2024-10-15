package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/pflag"
)

var (
	version string = "dev"
	commit  string = "dev"
)

// Execute runs the root command and exits the program if it fails.
func Execute() {
	cmd := rootCmd()

	err := cmd.Execute()
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
	if verbose || trace {
		level = slog.LevelDebug
	}

	// Setup the TextHandler and ensure our configured logger is the default.
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	logger := slog.New(h)
	slog.SetDefault(logger)
	logLevel.Set(level)
}
