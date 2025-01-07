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
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

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
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 && args[0] == "-" {
			input := cmd.InOrStdin()
			b, err := io.ReadAll(input)
			if err != nil {
				log.Fatalf("failed to read from stdin: %s", err)
			}
			args = strings.Fields(string(b))
		}

		if len(args) == 0 {
			fmt.Println("Specify torrent hashes as args or pipe to stdin")
			return
		}

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 1, 4, 3, ' ', 0)
		if !noHeaders {
			fmt.Fprint(w, "HASH\tSIZE\tPROGRESS\tNAME\n")
		}

		out := map[string]struct{}{}

		c := newClient()

		for {
			for _, h := range args {
				for _, f := range c.Files(h) {
					ln := fmt.Sprintf("%s\t%d\t%6.02f%%\t%s\n", h, f.Size, f.Progress*100, f.Name)
					if _, ok := out[ln]; ok {
						continue
					}
					out[ln] = struct{}{}
					fmt.Fprint(w, ln)
				}
			}
			w.Flush()
			if !watch {
				return
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
}
