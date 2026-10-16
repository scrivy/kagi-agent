package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/scrivy/kagi-agent/internal/agent"
	"github.com/scrivy/kagi-agent/internal/config"
)

const (
	appName      = config.AppName
	shortAppDesc = "A tui for coding and interacting with the kagi assistant"
)

var (
	rootCmd = &cobra.Command{
		Use:   appName,
		Short: shortAppDesc,
		RunE:  run,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(*cobra.Command, []string) error {
	err := agent.New().Run()
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}
