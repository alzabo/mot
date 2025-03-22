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

// resumeCmd represents the resume command
var resumeCmd = &cobra.Command{
	Use:     "resume hash...",
	Aliases: []string{"unpause", "start"},
	GroupID: "torrent",
	Short:   "Resume torrents",
	Long: `Resume torrents.

Examples:
  # Resume a torrent
  mot resume hash`,
	RunE: func(cmd *cobra.Command, args []string) error {
		args, err := parseArgs(cmd, args, true)
		if err != nil {
			return fmt.Errorf("failed to parse args: %s", err)
		}

		// Don't print usage for errors after flag validation
		cmd.SilenceUsage = true

		c := newClient()
		return c.Resume(mot.WithHashes(args))
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}
