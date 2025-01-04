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
	"io"
	"log"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	mot "github.com/alzabo/mot/pkg"
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
	Use:     "torrent",
	Aliases: []string{"torrents"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 && args[0] == "-" {
			input := cmd.InOrStdin()
			b, err := io.ReadAll(input)
			if err != nil {
				log.Fatalf("failed to read from stdin: %s", err)
			}
			args = strings.Fields(string(b))
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
			// TODO: More uniform validation
			if !slices.Contains(torrentStateFilters, state) {
				log.Fatalf("invalid torrent state filter provided: %s", state)
			}
			opts = append(opts, mot.WithValue("filter", state))
		}

		get(opts)
	},
	// TODO: Validation for Args: for valid hash or none
}

func get(opts []mot.QueryOption) {
	c := newClient()
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 1, 4, 3, ' ', 0)

	torrents := c.TorrentList(opts...)

	if len(torrents) == 0 {
		fmt.Fprintln(w, "No torrents found.")
		return
	}

	if !noHeaders {
		fmt.Fprint(w, "HASH\tSTATE\tNAME\n")
	}
	for _, t := range torrents {
		fmt.Fprintf(w, "%s\t%s\t%s\n", t.Hash, t.State, t.Name)
	}
	w.Flush()
	_ = c
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
}
