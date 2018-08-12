package testutils

import (
	"github.com/spf13/cobra"
	"testing"
)

func PrepareArgs(t *testing.T, cmd *cobra.Command, args []string) error {
	return cmd.ParseFlags(args)
}
