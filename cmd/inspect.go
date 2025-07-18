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
	"os"
	"strings"

	"github.com/alzabo/mot/output"
	"github.com/alzabo/mot/torrent"
	"github.com/alzabo/mot/torrent/parser"
	"github.com/spf13/cobra"
)

// inspectCmd represents the inspect command
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		args, err := parseArgs(cmd, args, true, false)
		if err != nil {
			return fmt.Errorf("failed to parse args: %s", err)
		}

		columns, err := cmd.Flags().GetStringSlice("columns")
		if err != nil {
			return err
		}

		if cmd.Flag("output").Value.String() == "hash" {
			columns = []string{"hash"}
			noHeaders = true
		}

		cmd.SilenceUsage = true

		p := output.Table[parser.Torrent]{
			Writer:  output.NewTableWriter(os.Stdout),
			Headers: !noHeaders,
			Watch:   watch,
			Columns: columns,
		}

		return p.Print(func() ([]parser.Torrent, error) {
			torrents := make([]parser.Torrent, len(args))
			for i, arg := range args {
				f, err := os.Open(arg)
				if err != nil {
					return nil, err
				}
				defer f.Close()

				torrents[i] = parser.Parse(f)
			}
			return torrents, nil
		})
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)

	inspectCmd.Flags().StringSliceP("columns", "c", []string{"hash", "size", "name"}, "Columns to print in tabular output. Valid columns: "+strings.Join(torrent.Info{}.Keys(), ", "))
	inspectCmd.Flags().StringP("output", "o", "table", "Output format.")
}
