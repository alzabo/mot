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

// filesCmd represents the files command
var getFilesCmd = &cobra.Command{
	Use:     "files hash...",
	Aliases: []string{"file"},
	Short:   "Display torrent files",
	Long: `Prints a table of torrent files.

Examples:
  # Print files for a torrent
  mot get files hash`,
	RunE: func(cmd *cobra.Command, args []string) error {
		args, err := parseArgs(cmd, args, true)
		if err != nil {
			return fmt.Errorf("failed to parse args: %s", err)
		}
		if len(args) == 0 {
			return errors.New("specify torrent hashes as args or pipe to stdin")
		}

		columns, err := cmd.Flags().GetStringSlice("columns")
		if err != nil {
			return err
		}

		rawFilters, err := cmd.Flags().GetStringArray("filter")
		if err != nil {
			return fmt.Errorf("failed to parse command line filters: %s", err)
		}
		filters, err := filter.ParseAll(rawFilters)
		if err != nil {
			return fmt.Errorf("encountered error while parsing filters: %s", err)
		}

		// Don't print usage for errors after flag validation
		cmd.SilenceUsage = true

		p := output.Table[torrent.File]{
			Writer:  output.NewTableWriter(os.Stdout),
			Headers: !noHeaders,
			Watch:   watch,
			Columns: columns,
			Filters: filters,
		}

		c := newClient()
		return p.Print(func() ([]torrent.File, error) {
			files := torrent.Files{}
			// TODO: is it any faster to apply filters here?
			for _, hash := range args {
				files = append(files, c.Files(hash)...)
			}
			return files, nil

		})
	},
}

func init() {
	getCmd.AddCommand(getFilesCmd)

	getFilesCmd.Flags().StringSliceP("columns", "c", []string{"hash", "size", "progress", "name"}, "Columns to print in tabular output. Valid columns: "+strings.Join(torrent.File{}.Keys(), ", "))
	getFilesCmd.Flags().StringArray("filter", nil, "Filters to apply to the torrent list")
}
