package cmd

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func parseArgs(cmd *cobra.Command, args []string, readStdin bool) ([]string, error) {
	isTtty := term.IsTerminal(int(os.Stdin.Fd()))
	if readStdin && len(args) == 0 && !isTtty || (len(args) == 1 && args[0] == "-") {
		input := cmd.InOrStdin()
		b, err := io.ReadAll(input)
		if err != nil {
			log.Fatalf("failed to read from stdin: %s", err)
		}
		args = strings.Fields(string(b))
	}
	return args, nil
}
