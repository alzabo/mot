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
	"os"
	"strings"
	"time"

	"github.com/alzabo/mot/torrent"
	"github.com/spf13/cobra"
)

// getTrackerCmd represents the tracker command
var getTrackerCmd = &cobra.Command{
	Use:     "tracker hash...",
	Aliases: []string{"trackers"},
	Short:   "Display torrent trackers",
	Long: `Display torrent tracker information.

Examples:
  # Display trackers for a torrent
  mot get trackers hash

  # Display trackers with NOT_WORKING status
  mot get trackers -a --filter=status=NOT_WORKING`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		all, _ := cmd.Flags().GetBool("all")
		args, err = parseArgs(cmd, args, !all)
		if err != nil {
			return fmt.Errorf("failed to parse args: %s", err)
		}
		if !all && len(args) == 0 {
			return errors.New("specify torrent hashes as args or pipe to stdin")
		}
		if all && len(args) > 0 {
			return errors.New("option --all may not be combined with hashes specified as args or via stdin")
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
		filters, err := parseFilters(rawFilters)
		if err != nil {
			return fmt.Errorf("encountered error while parsing filters: %s", err)
		}

		// Don't print usage for errors after flag validation
		cmd.SilenceUsage = true

		w := writer(os.Stdout)
		if !noHeaders {
			w.WriteFunc(columns, strings.ToUpper)
		}

		c := newClient()

		hashes := []string{}
		if all && len(args) == 0 {
			for _, t := range c.Torrents() {
				hashes = append(hashes, t.Hash)
			}
		} else {
			hashes = args
		}

		for {
		tracker:
			for _, t := range c.Trackers(hashes) {
				ok, err := filters.All(t)
				if err != nil {
					return err
				}
				if !ok {
					continue tracker
				}

				fields := make([]string, len(columns))
				for i, key := range columns {
					val, err := t.Get(key)
					if err != nil {
						return fmt.Errorf("key %v not found in object; available keys: [%s]", key, strings.Join(t.Keys(), ","))
					}
					fields[i] = val.String()
				}
				w.WriteOnce(fields)
			}
			w.Flush()
			if !watch {
				return nil
			}
			time.Sleep(watchSleep)
		}
	},
}

func init() {
	getCmd.AddCommand(getTrackerCmd)

	getTrackerCmd.Flags().StringSliceP("columns", "c", []string{"hash", "status", "url", "message"}, "Columns to print in tabular output. Valid columns: "+strings.Join(torrent.Tracker{}.Keys(), ", "))
	getTrackerCmd.Flags().StringArray("filter", []string{"status!=DISABLED"}, "Filters to apply to the torrent list")
	getTrackerCmd.Flags().BoolP("all", "a", false, "Get all trackers for all torrents")
	getTrackerCmd.Flags().StringP("output", "o", "table", "Output format.")
}
