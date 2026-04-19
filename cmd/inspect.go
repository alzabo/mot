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
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alzabo/mot/output"
	"github.com/alzabo/mot/torrent/parser"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [file...]",
	Short: "Display torrent file information",
	Long: `Display information about one or more .torrent files.

Examples:
  # Inspect a single torrent file
  mot inspect file.torrent

  # Inspect multiple files
  mot inspect file1.torrent file2.torrent

  # Pipe multiple torrents from stdin
  cat file1.torrent file2.torrent | mot inspect

  # Or with process substitution
  mot inspect <(cat file1.torrent file2.torrent)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		columns, err := cmd.Flags().GetStringSlice("columns")
		if err != nil {
			return err
		}

		if cmd.Flag("output").Value.String() == "hash" {
			columns = []string{"hash"}
			noHeaders = true
		}

		cmd.SilenceUsage = true

		isTTY := term.IsTerminal(int(os.Stdin.Fd()))
		useStdin := !isTTY

		var torrents []parser.Torrent
		if useStdin && len(args) == 0 {
			torrents, err = parseStdin(cmd.InOrStdin())
			if err != nil {
				return err
			}
		} else {
			args, err := parseArgs(cmd, args, true, false)
			if err != nil {
				return err
			}

			torrents, err = parseFiles(args)
			if err != nil {
				return err
			}
		}

		if len(torrents) == 0 {
			return nil
		}

		p := output.Table[parser.Torrent]{
			Writer:  output.NewTableWriter(os.Stdout),
			Headers: !noHeaders,
			Watch:   watch,
			Columns: columns,
		}

		return p.Print(func() ([]parser.Torrent, error) {
			return torrents, nil
		})
	},
}

func parseStdin(r io.Reader) ([]parser.Torrent, error) {
	var torrents []parser.Torrent
	err := parser.ParseStream(r, func(t parser.Torrent) {
		fmt.Printf("DEBUG: got torrent: Name=%q Size=%d Hash=%s\n", t.Name, t.Size, t.Hash)
		torrents = append(torrents, t)
	})
	if err != nil {
		return nil, fmt.Errorf("parsing stdin: %w", err)
	}
	return torrents, nil
}

func parseFiles(paths []string) ([]parser.Torrent, error) {
	torrents := make([]parser.Torrent, 0, len(paths))
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		torrent, err := parser.Parse(f)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		torrents = append(torrents, torrent)
	}
	return torrents, nil
}

func init() {
	rootCmd.AddCommand(inspectCmd)

	inspectCmd.Flags().StringSliceP("columns", "c", []string{"hash", "size", "name"}, "Columns to print in tabular output. Valid columns: "+strings.Join(parser.Torrent{}.Keys(), ", "))
	inspectCmd.Flags().StringP("output", "o", "table", "Output format.")
}
