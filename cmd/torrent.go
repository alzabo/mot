/*
Copyright © 2024 Ryan White

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
	"slices"
	"strings"
	"text/tabwriter"
	"time"

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

// torrentCmd represents the torrent command
var torrentCmd = &cobra.Command{
	Use:     "torrent [hash...]",
	Aliases: []string{"torrents"},
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

		opts := []mot.QueryOption{}
		if len(args) > 0 {
			opts = append(opts, mot.WithHashes(args))
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
		filters, err := parseFilters(rawFilters)
		if err != nil {
			return fmt.Errorf("encountered error while parsing filters: %s", err)
		}

		c := newClient()

		// TODO: This map can grow unbounded. Chances are it doesn't matter much in
		// practice, but interning the strings would allow the garbage collector to
		// reclaim memory. https://go.dev/blog/unique
		out := map[string]struct{}{}

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 1, 4, 3, ' ', 0)

		if !noHeaders {
			// TODO: Print headers again every N lines when watching?
			headers := make([]string, len(columns))
			for i, key := range columns {
				headers[i] = strings.ToUpper(key)
			}
			fmt.Fprint(w, strings.Join(headers, "\t")+"\t\n")
		}

		for {
			torrents := c.Torrents(opts...)
		torrents:
			for _, t := range torrents {
				for _, f := range filters {
					match, err := f(t)
					if err != nil {
						return err
					}
					if !match {
						continue torrents
					}
				}
				// TODO: When states change, the width of lines may also
				fields := make([]string, len(columns))
				for i, key := range columns {
					fields[i] = t.Get(key).String()
				}

				ln := strings.Join(fields, "\t") + "\t\n"
				if _, ok := out[ln]; !ok {
					out[ln] = struct{}{}
					fmt.Fprint(w, ln)
				}
			}

			w.Flush()

			if !watch {
				if len(torrents) == 0 {
					fmt.Fprintln(w, "No torrents found.")
				}
				break
			}
			time.Sleep(watchSleep)
		}
		return nil
	},
	// TODO: Validation for Args: for valid hash or none
}

func init() {
	getCmd.AddCommand(torrentCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// torrentCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	torrentCmd.Flags().String("category", "", "Select torrents with the given category")
	torrentCmd.Flags().String("tag", "", "Select torrents with the given tag")
	torrentCmd.Flags().String("state", "", "Select torrents in one of the states: "+strings.Join(torrentStateFilters, ", "))

	// TODO: validate existence of columns. Passing an invalid key results in a nil pointer panic
	torrentCmd.Flags().StringSliceP("columns", "c", []string{"hash", "state", "progress", "name"}, "Columns to print in tabular output. Valid columns: "+strings.Join(torrent.Info{}.Values().Keys(), ", "))
	torrentCmd.Flags().StringArray("filter", nil, "Filters to apply to the torrent list")

	torrentCmd.Flags().StringP("output", "o", "table", "Output format.")
	// -o hash; print the hash only
	// -o wide
	// -o activity?
	// -o columns=
	// output field separator
}
