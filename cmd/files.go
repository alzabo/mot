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
	"text/tabwriter"
	"time"

	"github.com/alzabo/mot/torrent"
	"github.com/spf13/cobra"
)

// filesCmd represents the files command
var filesCmd = &cobra.Command{
	Use:     "files hash...",
	Aliases: []string{"file"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		args, err := parseArgs(cmd, args)
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

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 1, 4, 3, ' ', 0)
		if !noHeaders {
			headers := make([]string, len(columns))
			for i, key := range columns {
				headers[i] = strings.ToUpper(key)
			}
			fmt.Fprint(w, strings.Join(headers, "\t")+"\t\n")
		}

		out := map[string]struct{}{}

		c := newClient()

		for {
			for _, h := range args {
				for _, f := range c.Files(h) {
					fields := make([]string, len(columns))
					for i, key := range columns {
						fields[i] = f.Get(key).String()
					}

					ln := strings.Join(fields, "\t") + "\t\n"
					if _, ok := out[ln]; ok {
						continue
					}
					out[ln] = struct{}{}
					fmt.Fprint(w, ln)
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
	getCmd.AddCommand(filesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// filesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// filesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// TODO: validate existence of columns. Passing an invalid key results in a nil pointer panic
	filesCmd.Flags().StringSliceP("columns", "c", []string{"hash", "size", "progress", "name"}, "Columns to print in tabular output. Valid columns: "+strings.Join(torrent.File{}.Values(map[string]string{"hash": ""}).Keys(), ", "))
}
