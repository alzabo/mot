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

// getTrackerCmd represents the tracker command
var getTrackerCmd = &cobra.Command{
	Use:     "tracker",
	Aliases: []string{"trackers"},
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

		out := map[string]struct{}{}

		c := newClient()

		for {
			for _, h := range args {
			tracker:
				for _, t := range c.Trackers(h) {
					for _, f := range filters {
						match, err := f(t)
						if err != nil {
							return err
						}
						if !match {
							continue tracker
						}

					}
					fields := make([]string, len(columns))
					for i, col := range columns {
						key, mod, _ := strings.Cut(col, "+")
						val, err := t.Get(key)
						if err != nil {
							return fmt.Errorf("key %v not found in object; available keys: [%s]", key, strings.Join(t.Keys(), ","))
						}
						if mod == "raw" {
							fields[i] = val.RawString()
						} else {
							fields[i] = val.String()
						}
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
	getCmd.AddCommand(getTrackerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getTrackerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getTrackerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	getTrackerCmd.Flags().StringSliceP("columns", "c", []string{"hash", "status", "url", "message"}, "Columns to print in tabular output. Valid columns: "+strings.Join(torrent.Tracker{}.Values(map[string]string{"hash": ""}).Keys(), ", "))
	getTrackerCmd.Flags().StringArray("filter", []string{"status!=DISABLED"}, "Filters to apply to the torrent list")
}
