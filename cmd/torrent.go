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
	"log"
	"os"
	"text/tabwriter"

	mot "github.com/alzabo/mot/pkg"
	"github.com/alzabo/mot/torrents"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
		get(args)
	},
	// TODO: Validation for Args: for valid hash or none
}

func get(args []string) {
	// TODO: Move client init into root?
	c, err := mot.NewClient(viper.GetString("url"), viper.GetString("username"), viper.GetString("password"))
	if err != nil {
		log.Fatal(err)
	}
	w := new(tabwriter.Writer)
	var torrents []torrents.Info
	w.Init(os.Stdout, 1, 4, 3, ' ', 0)
	if len(args) == 0 {
		torrents = c.TorrentList()
	}
	if len(torrents) == 0 {
		fmt.Println("No torrents found.")
		return
	}
	fmt.Fprint(w, "HASH\tSTATE\tNAME")
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
	//torrentCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
