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
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/alzabo/mot/filter"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func parseArgs(cmd *cobra.Command, args []string, readStdin bool, split bool) ([]string, error) {
	var parsed []string
	if split {
		for _, i := range args {
			parsed = append(parsed, strings.Fields(i)...)
		}
	} else {
		parsed = args
	}

	isTTY := term.IsTerminal(int(os.Stdin.Fd()))
	if !readStdin || len(parsed) == 0 && isTTY || (len(parsed) == 1 && parsed[0] != "-") {
		return parsed, nil
	}

	s := bufio.NewScanner(cmd.InOrStdin())
	for s.Scan() {
		if s.Err() != nil {
			log.Fatalf("failed to read from stdin: %s", s.Err())
		}
		if split {
			parsed = append(parsed, strings.Fields(s.Text())...)
			continue
		}
		parsed = append(parsed, s.Text())
	}
	return parsed, nil
}

var hashExpr *regexp.Regexp = regexp.MustCompile(`[0-9a-f]{40}`)

func parseMixedArgs(cmd *cobra.Command, args []string, readStdin bool) ([]string, filter.Filters, error) {
	args, err := parseArgs(cmd, args, readStdin, true)
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

// supportedURL matches URL formats supported by the qBittorrent add torrent API.
// Supported: http://, https://, magnet: and bc://bt/
func supportedURL(url string) bool {
	proto, _, ok := strings.Cut(url, ":")
	if !ok {
		return false
	}
	validProtos := []string{"http", "https", "magnet", "bc"}
	return slices.Contains(validProtos, proto)
}

func parseAddTorrentArgs(cmd *cobra.Command, args []string, readStdin bool) error {
	parsed, err := parseArgs(cmd, args, readStdin, false)
	if err != nil {
		return err
	}

	for _, i := range parsed {
		if supportedURL(i) {
			err := cmd.Flags().Set("url", i)
			if err != nil {
				return err
			}
			continue
		}
		info, err := os.Stat(i)
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink == 0 {
			err := cmd.Flags().Set("torrent", i)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
