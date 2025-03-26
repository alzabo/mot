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
	"errors"
	"fmt"

	mot "github.com/alzabo/mot/pkg"
	"github.com/spf13/cobra"
)

var deleteFiles bool

// deleteTorrentCmd represents the torrent command
var deleteTorrentCmd = &cobra.Command{
	Use:     "torrent",
	Aliases: []string{"torrents", "tor"},
	Short:   "Delete torrents",
	Long: `Delete torrents.

Examples:
  # Delete torrent, preserving related files
  mot delete torrent hash

  # Delete torrent and associated files
  mot delete torrent hash --delete-files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		args, err := parseArgs(cmd, args, true)
		if err != nil {
			return fmt.Errorf("failed to parse args: %s", err)
		}
		if len(args) == 0 {
			return errors.New("specify torrent hashes as args or pipe to stdin")
		}
		opts := make([]mot.QueryOption, 0, len(args)+1)
		opts = append(opts, mot.WithHashes(args))
		opts = append(opts, mot.WithValue("deleteFiles", fmt.Sprintf("%v", deleteFiles)))

		c := newClient()
		return c.DeleteTorrents(opts...)
	},
}

func init() {
	deleteCmd.AddCommand(deleteTorrentCmd)

	deleteTorrentCmd.Flags().BoolVarP(&deleteFiles, "delete-data", "f", false, "delete torrent data")

	// TODO: interactive mode
}
