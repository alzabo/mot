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
	mot "github.com/alzabo/mot/pkg"
	"github.com/spf13/cobra"
)

// torrentAddCmd represents the torrent command
var torrentAddCmd = &cobra.Command{
	Use:     "torrent",
	Aliases: []string{"torrents", "tor"},
	Short:   "Add torrent",
	Long:    `Add torrents to qBittorrent server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var req mot.AddTorrent
		err := mot.UpdateFromCmd(cmd, &req)
		if err != nil {
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
