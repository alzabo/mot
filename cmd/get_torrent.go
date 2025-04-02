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
	"slices"
	"strings"

	"github.com/alzabo/mot/filter"
	"github.com/alzabo/mot/output"
	mot "github.com/alzabo/mot/pkg"
	"github.com/alzabo/mot/torrent"
	"github.com/spf13/cobra"
)

var torrentStateFilters = []string{
	"all",
	"downloading",
	"seeding",
	"completed",
	"paused",
	"active",
	"inactive",
	"resumed",
	"stalled",
	"stalled_uploading",
	"stalled_downloading",
	"errored",
}

// getTorrentCmd represents the torrent command
var getTorrentCmd = &cobra.Command{
	Use:     "torrent [hash... | filter...]",
	Aliases: []string{"torrents", "tor"},
	Short:   "Display torrent information",
	Long: `Display torrent information.

Example:
  # Get information for a torrent
  mot get torrent <torrent_hash>

  # Get information for all torrents
  mot get torrents -a`,
	RunE: func(cmd *cobra.Command, args []string) error {
		hashes, filters, err := parseMixedArgs(cmd, args, true)
		if err != nil {
			return fmt.Errorf("failed to parse args: %s", err)
		}

		if all, _ := cmd.Flags().GetBool("all"); all && len(args) > 0 {
			return errors.New("option --all may not be combined with hashes specified as args or through stdin")
		}

		opts := []mot.QueryOption{}
		if len(hashes) > 0 {
			opts = append(opts, mot.WithHashes(hashes))
		}

		if cmd.Flags().Changed("category") {
			opts = append(opts, mot.WithValue("category", cmd.Flag("category").Value.String()))
		}
		if cmd.Flags().Changed("tag") {
			opts = append(opts, mot.WithValue("tag", cmd.Flag("tag").Value.String()))
		}
		if cmd.Flags().Changed("state") {
			state := cmd.Flag("state").Value.String()
			if !slices.Contains(torrentStateFilters, state) {
				return fmt.Errorf("invalid torrent state filter provided: %s", state)
			}
			opts = append(opts, mot.WithValue("filter", state))
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

		p := output.Table[torrent.Info]{
			Writer:  output.NewTableWriter(os.Stdout),
			Headers: !noHeaders,
			Watch:   watch,
			Columns: columns,
			Filters: filters,
		}

		c := newClient()

		return p.Print(func() ([]torrent.Info, error) {
			return c.Torrents(opts...), nil
		})
	},
}

func init() {
	getCmd.AddCommand(getTorrentCmd)

	getTorrentCmd.Flags().String("category", "", "Select torrents with the given category")
	getTorrentCmd.Flags().String("tag", "", "Select torrents with the given tag")
	getTorrentCmd.Flags().String("state", "", "Select torrents in one of the states: "+strings.Join(torrentStateFilters, ", "))

	getTorrentCmd.Flags().StringSliceP("columns", "c", []string{"hash", "state", "progress", "name"}, "Columns to print in tabular output. Valid columns: "+strings.Join(torrent.Info{}.Keys(), ", "))
	getTorrentCmd.Flags().StringArray("filter", nil, "Filters to apply to the torrent list")

	getTorrentCmd.Flags().StringP("output", "o", "table", "Output format.")
	getTorrentCmd.Flags().BoolP("all", "a", false, "Get all torrents")

	// -o hash; print the hash only
	// -o wide
	// -o activity?
	// -o columns=
	// output field separator
}
