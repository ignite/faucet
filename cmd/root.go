package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	
	"github.com/ignite/faucet/version"
)

const checkVersionTimeout = time.Millisecond * 600

// NewRootCmd creates a new root command for `Faucet` with its sub commands.
func NewRootCmd() *cobra.Command {
	c := &cobra.Command{
		Use:           "faucet",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Check for new versions only when shell completion scripts are not being
			// generated to avoid invalid output to stdout when a new version is available
			if cmd.Use != "completion" || !strings.HasPrefix(cmd.Use, cobra.ShellCompRequestCmd) {
				checkNewVersion(cmd.Context())
			}

			return nil
		},
		RunE: runCmdHandler,
	}

	faucetFlags(c)
	c.AddCommand(NewVersionCmd())

	return c
}

// NewVersionCmd creates a new version command to show the faucet version.
func NewVersionCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "version",
		Short: "Print the current build information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			v, err := version.Long(cmd.Context())
			if err != nil {
				return err
			}
			cmd.Println(v)
			return nil
		},
	}
	return c
}

func checkNewVersion(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, checkVersionTimeout)
	defer cancel()

	isAvailable, next, err := version.CheckNext(ctx)
	if err != nil || !isAvailable {
		return
	}

	fmt.Printf("⬆️ Gex %[1]v is available! To upgrade: https://github.com/ignite/gex/releases/%[1]v", next)
}
