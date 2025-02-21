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

	mot "github.com/alzabo/mot/pkg"
	"github.com/spf13/cobra"
)

// recheckCmd represents the recheck command
var recheckCmd = &cobra.Command{
	Use:     "recheck hash...",
	Aliases: []string{"check"},
	GroupID: "torrent",
	Short:   "Recheck torrent contents",
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

		// Don't print usage for errors after flag validation
		cmd.SilenceUsage = true

		c := newClient()

		// TODO: watch option to exec get torrents command after starting recheck
		// should implement watch until... progress == 100% for hashes.
		return c.Recheck(mot.WithHashes(args))
	},
}

func init() {
	rootCmd.AddCommand(recheckCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// recheckCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// recheckCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
