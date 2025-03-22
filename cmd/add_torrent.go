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
	"fmt"
	"reflect"

	mot "github.com/alzabo/mot/pkg"
	"github.com/spf13/cobra"
)

var flagToName = map[string]string{}

// torrentAddCmd represents the torrent command
var torrentAddCmd = &cobra.Command{
	Use:     "torrent",
	Aliases: []string{"torrents", "tor"},
	Short:   "Add torrent",
	Long:    `Add torrents to qBittorrent server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		req, err := newRequest(cmd)
		if err != nil {
			return err
		}
		return newClient().AddTorrent(&req)
	},
}

func newRequest(cmd *cobra.Command) (mot.AddTorrent, error) {
	req := mot.AddTorrent{}
	for f, field := range flagToName {
		flag := cmd.Flag(f)
		if !flag.Changed {
			continue
		}
		value := reflect.ValueOf(&req).Elem().FieldByName(field)
		switch flag.Value.Type() {
		case "bool":
			v, _ := cmd.Flags().GetBool(f)
			value.Set(reflect.ValueOf(&v))
		case "float64":
			v, _ := cmd.Flags().GetFloat64(f)
			value.Set(reflect.ValueOf(&v))
		case "int64":
			v, _ := cmd.Flags().GetInt64(f)
			value.Set(reflect.ValueOf(&v))
		case "string":
			v, _ := cmd.Flags().GetString(f)
			value.Set(reflect.ValueOf(&v))
		case "stringSlice":
			v, _ := cmd.Flags().GetStringSlice(f)
			value.Set(reflect.ValueOf(&v))
		default:
			return req, fmt.Errorf("failed to set value from flag %s", f)
		}
	}
	return req, nil
}

func init() {
	addCmd.AddCommand(torrentAddCmd)

	t := reflect.TypeOf(mot.AddTorrent{})
	for i := range t.NumField() {
		f := t.Field(i)
		flag := ToKebabCase(f.Name)
		flagToName[flag] = f.Name
		usage := f.Tag.Get("usage")
		short := f.Tag.Get("short")

		switch f.Type.Kind() {
		case reflect.Pointer:
			switch f.Type.Elem().Kind() {
			case reflect.String:
				torrentAddCmd.Flags().StringP(flag, short, "", usage)
			case reflect.Bool:
				torrentAddCmd.Flags().BoolP(flag, short, false, usage)
			case reflect.Int64:
				torrentAddCmd.Flags().Int64P(flag, short, 0, usage)
			case reflect.Float64:
				torrentAddCmd.Flags().Float64P(flag, short, 0, usage)
			case reflect.Slice:
				switch f.Type.Elem().Elem().Kind() {
				case reflect.String:
					torrentAddCmd.Flags().StringSliceP(flag, short, nil, usage)
				default:
					panic("unreachable")
				}
			default:
				panic("unreachable")
			}
		default:
			panic("unreachable")
		}
	}
}
