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
	"os"
	"strings"
	"time"

	mot "github.com/alzabo/mot/pkg"
	"github.com/alzabo/mot/torrent"
	"github.com/spf13/cobra"
)

// getLogsCmd represents the logs command
var getLogsCmd = &cobra.Command{
	Use:     "logs",
	Aliases: []string{"log"},
	Short:   "Get qBittorrent server logs",
	Long: `Retrieve logs from the server.
	
Examples:
  # Get all logs
  mot get logs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var req mot.MainLog
		err := updateFromCmd(cmd, &req)
		if err != nil {
			return err
		}

		columns, err := cmd.Flags().GetStringSlice("columns")
		if err != nil {
			return err
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

		c := newClient()

		w := writer(os.Stdout)

		fields := make([]string, len(columns))

		if !noHeaders {
			// TODO: Print headers again every N lines when watching?
			w.WriteFunc(columns, strings.ToUpper)
		}

		for {
			logs, err := c.Logs(req)
			if err != nil {
				return err
			}
		log:
			for _, t := range logs {
				ok, err := filters.All(t)
				if err != nil {
					return err
				}
				if !ok {
					continue log
				}
				// TODO: When states change, the width of lines may also
				for i, key := range columns {
					v, err := t.Get(key)
					if err != nil {
						return fmt.Errorf("key %v not found in object; available keys: [%s]", key, strings.Join(t.Keys(), ","))
					}
					fields[i] = v.String()
				}

				w.WriteOnce(fields)
			}

			w.Flush()

			if !watch {
				if len(logs) == 0 {
					w.Write([]string{"No logs found."})
				}
				break
			}
			time.Sleep(watchSleep)
		}

		return nil
	},
}

func init() {
	getCmd.AddCommand(getLogsCmd)
	flagsFromStruct(getLogsCmd, mot.MainLog{})
	getLogsCmd.Flags().StringSliceP("columns", "c", []string{"id", "timestamp", "type", "message"}, "Columns to print in tabular output. Valid columns: "+strings.Join(torrent.Log{}.Keys(), ", "))
	getLogsCmd.Flags().StringArray("filter", nil, "Filters to apply to the torrent list")
}
