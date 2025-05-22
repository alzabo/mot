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
	"io/fs"
	"os"
	"slices"
	"strings"

	mot "github.com/alzabo/mot/pkg"
	"github.com/spf13/cobra"
)

// torrentAddCmd represents the torrent command
var torrentAddCmd = &cobra.Command{
	Use:     "torrent",
	Aliases: []string{"torrents", "tor"},
	Short:   "Add torrent",
	Long:    `Add torrents to qBittorrent server.`,
	Example: `
# Add a torrent from a local file
mot add torrent -f <path>

# Add multiple local torrent files by piping a list of file names to stdin
ls *.torrent | mot add torrent`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urls, files, err := parseAddTorrentArgs(cmd, args, true)
		if err != nil {
			return err
		}

		for _, url := range urls {
			err := cmd.Flags().Set("url", url)
			if err != nil {
				return err
			}
		}
		for _, file := range files {
			err := cmd.Flags().Set("torrent", file)
			if err != nil {
				return err
			}
		}

		var req mot.AddTorrent
		if err := mot.UpdateFromCmd(cmd, &req); err != nil {
			return err
		}

		// Don't print usage for errors after flag validation
		cmd.SilenceUsage = true

		return newClient().AddTorrent(&req)
	},
}

func init() {
	addCmd.AddCommand(torrentAddCmd)
	mot.AddFlagsForPayload(torrentAddCmd, mot.AddTorrent{})
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

func parseAddTorrentArgs(cmd *cobra.Command, args []string, readStdin bool) ([]string, []string, error) {
	args, err := parseArgs(cmd, args, readStdin)
	if err != nil {
		return nil, nil, err
	}

	urls := []string{}
	files := []string{}
	for _, i := range args {
		if supportedURL(i) {
			urls = append(urls, i)
			continue
		}
		info, err := os.Stat(i)
		if err != nil {
			return nil, nil, err
		}
		if info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink == 0 {
			files = append(files, i)
		}
	}
	return urls, files, nil
}
