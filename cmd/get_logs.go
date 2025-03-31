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

	"github.com/alzabo/mot/filter"
	"github.com/alzabo/mot/output"
	mot "github.com/alzabo/mot/pkg"
	"github.com/alzabo/mot/torrent"
	"github.com/spf13/cobra"
)

// getLogsCmd represents the logs command
var getLogsCmd = &cobra.Command{
	Use:     "logs [filter...]",
	Aliases: []string{"log"},
	Short:   "Get qBittorrent server logs",
	Long: `Retrieve logs from the server.
	
Examples:
  # Get all logs
  mot get logs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, filters, err := parseMixedArgs(cmd, args, true)
		if err != nil {
			return fmt.Errorf("failed to parse args: %s", err)
		}

		var req mot.MainLog
		if err := updateFromCmd(cmd, &req); err != nil {
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
		parsedFilters, err := filter.ParseAll(rawFilters)
		if err != nil {
			return fmt.Errorf("encountered error while parsing filters: %s", err)
		}
		filters = append(filters, parsedFilters...)

		// Don't print usage for errors after flag validation
		cmd.SilenceUsage = true

		p := output.Table[torrent.Log]{
			Writer:  output.NewTableWriter(os.Stdout),
			Headers: !noHeaders,
			Watch:   watch,
			Columns: columns,
			Filters: filters,
		}

		c := newClient()

		return p.Print(func() ([]torrent.Log, error) {
			logs, err := c.Logs(req)
			if err != nil {
				return nil, err
			}
			return logs, nil
		})
	},
}

func init() {
	getCmd.AddCommand(getLogsCmd)
	flagsFromStruct(getLogsCmd, mot.MainLog{})
	getLogsCmd.Flags().StringSliceP("columns", "c", []string{"id", "timestamp", "type", "message"}, "Columns to print in tabular output. Valid columns: "+strings.Join(torrent.Log{}.Keys(), ", "))
	getLogsCmd.Flags().StringArray("filter", nil, "Filters to apply to the torrent list")
}
