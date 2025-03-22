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

	mot "github.com/alzabo/mot/pkg"
	"github.com/spf13/cobra"
)

// pauseCmd represents the pause command
var pauseCmd = &cobra.Command{
	Use:     "pause hash...",
	Aliases: []string{"stop"},
	GroupID: "torrent",
	Short:   "Pause torrents",
	Long: `Pause torrents.

Examples:
  # Pause a torrent
  mot pause hash

  # Pause all torrents
  mot get torrents -a -o hash | mot pause`,
	RunE: func(cmd *cobra.Command, args []string) error {
		args, err := parseArgs(cmd, args, true)
		if err != nil {
			return fmt.Errorf("failed to parse args: %s", err)
		}

		// Don't print usage for errors after flag validation
		cmd.SilenceUsage = true

		c := newClient()
		return c.Pause(mot.WithHashes(args))
	},
}

func init() {
	rootCmd.AddCommand(pauseCmd)
}
