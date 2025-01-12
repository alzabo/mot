package cmd

import (
	"io"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

func parseArgs(cmd *cobra.Command, args []string) []string {
	if len(args) == 1 && args[0] == "-" {
		input := cmd.InOrStdin()
		b, err := io.ReadAll(input)
		if err != nil {
			log.Fatalf("failed to read from stdin: %s", err)
		}
		return strings.Fields(string(b))
	}
	return args
}
