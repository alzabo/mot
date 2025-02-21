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

// filesCmd represents the files command
var getFilesCmd = &cobra.Command{
	Use:     "files hash...",
	Aliases: []string{"file"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

		for {
			for _, h := range args {
			file:
				for _, f := range c.Files(h) {
					ok, err := filters.All(f)
					if err != nil {
						return err
					}
					if !ok {
						continue file
					}

					fields := make([]string, len(columns))
					for i, key := range columns {
						val, err := f.Get(key)
						if err != nil {
							return fmt.Errorf("key %v not found in object; available keys: [%s]", key, strings.Join(f.Keys(), ","))
						}
						fields[i] = val.String()
					}
					w.WriteOnce(fields)
				}
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
	getCmd.AddCommand(getFilesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// filesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// filesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	getFilesCmd.Flags().StringSliceP("columns", "c", []string{"hash", "size", "progress", "name"}, "Columns to print in tabular output. Valid columns: "+strings.Join(torrent.File{}.Keys(), ", "))
	getFilesCmd.Flags().StringArray("filter", nil, "Filters to apply to the torrent list")
}
