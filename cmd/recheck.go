/*
Copyright © 2025 Ryan White

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"errors"
	"fmt"

	mot "github.com/alzabo/mot/pkg"
	"github.com/spf13/cobra"
)

// recheckCmd represents the recheck command
var recheckCmd = &cobra.Command{
	Use:     "recheck hash...",
	Aliases: []string{"check"},
	GroupID: "torrent",
	Short:   "Recheck torrent contents",
	Long: `Recheck torrent files.

Examples:
  # Recheck a torrent
  mot recheck hash`,
	RunE: func(cmd *cobra.Command, args []string) error {
		args, err := parseArgs(cmd, args, true)
		if err != nil {
			return fmt.Errorf("failed to parse args: %s", err)
		}
		if len(args) == 0 {
			return errors.New("specify torrent hashes as args or pipe to stdin")
		}

		// Don't print usage for errors after flag validation
		cmd.SilenceUsage = true

		c := newClient()

		// TODO: watch option to exec get torrents command after starting recheck
		// should implement watch until... progress == 100% for hashes.
		return c.Recheck(mot.WithHashes(args))
	},
}

func init() {
	rootCmd.AddCommand(recheckCmd)
}
