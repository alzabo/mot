// Copyright 2025 Ryan White
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var noHeaders bool
var watch bool

const watchSleep time.Duration = 1000

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:     "get",
	GroupID: "basic",
	Short:   "Display information for objects",
	Long: `Prints a table of objects.

Examples:
  # List all torrents
  mot get torrents

  # List all trackers
  mot get trackers`,
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.PersistentFlags().BoolVarP(&noHeaders, "no-headers", "H", false, "When using default or columnar output, don't print headers")
	getCmd.PersistentFlags().BoolVarP(&watch, "watch", "w", false, "After getting the object, watch for changes")
}
