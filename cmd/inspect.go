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
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"os"

	"github.com/alzabo/mot/torrent/tokenizer"
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
		if len(args) == 0 {
			return nil
		}
		f, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer f.Close()

		st := []tokenizer.Type{}
		var n int
		infoSlice := [2]int{}
		tokens, _ := tokenizer.Tokenize(f)
		for i, t := range tokens {
			if t.Type == tokenizer.DictStart {
				st = append(st, t.Type)
				if i < 1 {
					continue
				}
				if bytes.Equal(tokens[i-1].Bytes, []byte{'i', 'n', 'f', 'o'}) {
					n = len(st)
					fmt.Println("info start", fmt.Sprintf("%v", t), t.Pos)
					infoSlice[0] = t.Pos
				}
			}
			if t.Type == tokenizer.DictEnd {
				if len(st) == n {
					fmt.Println("info end", t.Pos)
					infoSlice[1] = t.Pos
					break
				}
				st = st[:len(st)-1]
			}
		}
		f.Seek(0, 0)
		b, _ := io.ReadAll(f)
		fmt.Println("first byte:", b[0], ":", string(b[0]))
		fmt.Println("around info:", string(b[infoSlice[0]-5:infoSlice[0]+5]))
		fmt.Println("around info end:", string(b[infoSlice[1]-5:infoSlice[1]+5]))
		info := b[infoSlice[0] : infoSlice[1]+1]
		fmt.Println("first:", info[0], ":", string(info[0]), "last:", info[len(info)-1], ":", string(info[len(info)-1]))
		hash := sha1.Sum(info)
		fmt.Printf("%x\n", hash)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// inspectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// inspectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
