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

func ToKebabCase(s string) string {
	c := make([]rune, 0, len(s)+4)
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 && s[i-1] >= 'a' && s[i-1] <= 'z' {
				c = append(c, '-')
			}
			r = r + 32 // convert to lower case
		}
		c = append(c, r)
	}
	return string(c)
}
