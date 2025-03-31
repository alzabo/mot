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
	"os"
	"strings"

	"github.com/alzabo/mot/filter"
	"github.com/alzabo/mot/output"
	"github.com/alzabo/mot/torrent"
	"github.com/spf13/cobra"
)

// getTrackerCmd represents the tracker command
var getTrackerCmd = &cobra.Command{
	Use:     "tracker [hash... | filter...]",
	Aliases: []string{"trackers"},
	Short:   "Display torrent trackers",
	Long: `Display torrent tracker information.

Examples:
  # Display trackers for a torrent
  mot get trackers hash

  # Display trackers with NOT_WORKING status
  mot get trackers --filter=status=NOT_WORKING`,
	RunE: func(cmd *cobra.Command, args []string) error {
		hashes, filters, err := parseMixedArgs(cmd, args, true)
		if err != nil {
			return fmt.Errorf("failed to parse args: %s", err)
		}

		if all, _ := cmd.Flags().GetBool("all"); all && len(args) > 0 {
			return errors.New("option --all may not be combined with hashes specified as args or through stdin")
		}

		columns, err := cmd.Flags().GetStringSlice("columns")
		if err != nil {
			return err
		}

		if cmd.Flag("output").Value.String() == "hash" {
			columns = []string{"hash"}
			noHeaders = true
		}

		rawFilters, err := cmd.Flags().GetStringArray("filter")
		if err != nil {
			return fmt.Errorf("failed to parse command line filters: %s", err)
		}
		parsedFilters, err := filter.ParseAll(rawFilters)
		if err != nil {
			return fmt.Errorf("encountered error while parsing filters: %s", err)
		}
		filters = append(filters, parsedFilters...)

		// Don't print usage for errors after flag validation
		cmd.SilenceUsage = true

		p := output.Table[torrent.Tracker]{
			Columns: columns,
			Filters: filters,
			Headers: !noHeaders,
			Watch:   watch,
			Writer:  output.NewTableWriter(os.Stdout),
			Sleep:   watchSleep,
		}

		c := newClient()

		return p.Print(func() ([]torrent.Tracker, error) {
			if len(hashes) == 0 {
				for _, t := range c.Torrents() {
					hashes = append(hashes, t.Hash)
				}
			}
			return c.Trackers(hashes), nil
		})
	},
}

func init() {
	getCmd.AddCommand(getTrackerCmd)

	getTrackerCmd.Flags().StringSliceP("columns", "c", []string{"hash", "status", "url", "message"}, "Columns to print in tabular output. Valid columns: "+strings.Join(torrent.Tracker{}.Keys(), ", "))
	getTrackerCmd.Flags().StringArray("filter", []string{"status!=DISABLED"}, "Filters to apply to the torrent list")
	getTrackerCmd.Flags().BoolP("all", "a", false, "Get all trackers for all torrents")
	getTrackerCmd.Flags().StringP("output", "o", "table", "Output format.")
}
