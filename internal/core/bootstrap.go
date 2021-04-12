package core

import (
	"fmt"
	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/UpCloudLtd/cli/internal/commands"
	"github.com/UpCloudLtd/cli/internal/config"
	"github.com/UpCloudLtd/cli/internal/log"
	"github.com/UpCloudLtd/cli/internal/terminal"
	"github.com/UpCloudLtd/cli/internal/ui"
)

// BuildRootCmd builds the root command
func BuildRootCmd(_ []string, conf *config.Config) cobra.Command {
	rootCmd := cobra.Command{
		Use:   "upctl",
		Short: "UpCloud CLI",
		Long:  "upctl a CLI tool for managing your UpCloud services.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			_, err := valid.ValidateStruct(conf.GlobalFlags)
			if err != nil {
				return err
			}

			terminal.ForceColours(conf.GlobalFlags.Colors)
			if err := log.SetDebugMode(conf.GlobalFlags.Debug); err != nil {
				return fmt.Errorf("cannot set debug mode: %w", err)
			}

			if err := conf.Load(); err != nil {
				return fmt.Errorf("cannot load configuration: %w", err)
			}

			return nil
		},

		Run: func(cmd *cobra.Command, args []string) {},
	}

	rootCmd.BashCompletionFunction = commands.CustomBashCompletionFunc(rootCmd.Use)

	flags := &pflag.FlagSet{}
	flags.StringVarP(
		&conf.GlobalFlags.ConfigFile, "config", "", "", "Config file",
	)
	flags.StringVarP(
		&conf.GlobalFlags.OutputFormat, "output", "o", "human",
		"Output format (supported: json, yaml and human)",
	)
	flags.BoolVar(
		&conf.GlobalFlags.Colors, "colours", true,
		"Use terminal colours (supported: auto, true, false)",
	)
	flags.BoolVar(
		&conf.GlobalFlags.Debug, "debug", false,
		"Print out more verbose debug logs",
	)
	flags.DurationVarP(
		&conf.GlobalFlags.ClientTimeout, "client-timeout", "t",
		time.Duration(60*time.Second),
		"CLI timeout when using interactive mode on some commands",
	)
	flags.BoolVarP(
		&conf.GlobalFlags.Wait, "wait", "w", false,
		"Wait for the command to be completed",
	)

	// Add flags
	flags.VisitAll(func(flag *pflag.Flag) {
		rootCmd.PersistentFlags().AddFlag(flag)
	})
	conf.ConfigBindFlagSet(flags)

	rootCmd.SetUsageTemplate(ui.CommandUsageTemplate())
	rootCmd.SetUsageFunc(ui.UsageFunc)

	return rootCmd
}
