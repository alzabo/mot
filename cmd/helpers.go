// Copyright 2025 Ryan White
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/alzabo/mot/filter"
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

var hashExpr *regexp.Regexp = regexp.MustCompile(`[0-9a-f]{40}`)

func parseMixedArgs(cmd *cobra.Command, args []string, readStdin bool) ([]string, filter.Filters, error) {
	args, err := parseArgs(cmd, args, readStdin)
	if err != nil {
		return nil, nil, err
	}
	hashes := []string{}
	filters := filter.Filters{}
	for _, arg := range args {
		if hashExpr.MatchString(arg) {
			hashes = append(hashes, arg)
			continue
		}

		filter, err := filter.Parse(arg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse arg as filter expression or hash: %s", arg)
		}
		filters = append(filters, filter)
	}
	return hashes, filters, nil
}
