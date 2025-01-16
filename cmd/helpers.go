package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

func parseArgs(cmd *cobra.Command, args []string) ([]string, error) {
	if len(args) == 1 && args[0] == "-" {
		input := cmd.InOrStdin()
		b, err := io.ReadAll(input)
		if err != nil {
			return args, fmt.Errorf("failed to read from stdin: %s", err)
		}
		return strings.Fields(string(b)), nil
	}
	return args, nil
}
