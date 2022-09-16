package main

import (
	"context"
	"fmt"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"

	"github.com/giantswarm/security-pack-helper/cmd"
)

func main() {
	err := mainE(context.Background())
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(2)
	}
}

func mainE(ctx context.Context) error {
	var err error

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		return microerror.Mask(err)
	}

	var rootCommand *cobra.Command
	{
		c := cmd.Config{
			Logger: logger,
		}
		rootCommand, err = cmd.New(c)
		if err != nil {
			return err
		}
	}

	err = rootCommand.ExecuteContext(ctx)
	if err != nil {
		return err
	}

	return nil
}
