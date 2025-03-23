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
	"io"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func parseArgs(cmd *cobra.Command, args []string, readStdin bool) ([]string, error) {
	isTtty := term.IsTerminal(int(os.Stdin.Fd()))
	if readStdin && len(args) == 0 && !isTtty || (len(args) == 1 && args[0] == "-") {
		input := cmd.InOrStdin()
		b, err := io.ReadAll(input)
		if err != nil {
			log.Fatalf("failed to read from stdin: %s", err)
		}
		args = strings.Fields(string(b))
	}
	return args, nil
}

func ToKebabCase(s string) string {
	c := make([]rune, 0, len(s)+4)
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 && s[i-1] >= 'a' && s[i-1] <= 'z' {
				c = append(c, '-')
			}
			r = r + 32 // convert to lower case
		}
		c = append(c, r)
	}
	return string(c)
}

func flagsFromStruct(cmd *cobra.Command, s any) {
	t := reflect.TypeOf(s)
	for i := range t.NumField() {
		f := t.Field(i)
		flagName := ToKebabCase(f.Name)
		usage := f.Tag.Get("usage")
		short := f.Tag.Get("short")

		switch f.Type.Kind() {
		case reflect.Pointer:
			switch f.Type.Elem().Kind() {
			case reflect.String:
				cmd.Flags().StringP(flagName, short, "", usage)
			case reflect.Bool:
				cmd.Flags().BoolP(flagName, short, false, usage)
			case reflect.Int64:
				cmd.Flags().Int64P(flagName, short, 0, usage)
			case reflect.Float64:
				cmd.Flags().Float64P(flagName, short, 0, usage)
			case reflect.Slice:
				switch f.Type.Elem().Elem().Kind() {
				case reflect.String:
					cmd.Flags().StringSliceP(flagName, short, nil, usage)
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

// updateFromCmd reads user-supplied flag data from cmd and stores the results
// in the value pointed to by v. If v is nil or not a pointer, updateFromCmd
// returns an error.
func updateFromCmd(cmd *cobra.Command, v any) error {
	pt := reflect.TypeOf(v)
	if pt == nil || pt.Kind() != reflect.Pointer {
		return fmt.Errorf("cannot update non-pointer value %v", v)
	}
	t := pt.Elem()
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("cannot update non-struct value %v", v)
	}
	for i := range t.NumField() {
		f := t.Field(i)
		flagName := ToKebabCase(f.Name)
		flag := cmd.Flag(flagName)
		if !flag.Changed {
			continue
		}
		value := reflect.ValueOf(v).Elem().FieldByName(f.Name)
		switch flag.Value.Type() {
		case "bool":
			fv, _ := cmd.Flags().GetBool(flagName)
			value.Set(reflect.ValueOf(&fv))
		case "float64":
			fv, _ := cmd.Flags().GetFloat64(flagName)
			value.Set(reflect.ValueOf(&fv))
		case "int64":
			fv, _ := cmd.Flags().GetInt64(flagName)
			value.Set(reflect.ValueOf(&fv))
		case "string":
			fv, _ := cmd.Flags().GetString(flagName)
			value.Set(reflect.ValueOf(&fv))
		case "stringSlice":
			fv, _ := cmd.Flags().GetStringSlice(flagName)
			value.Set(reflect.ValueOf(&fv))
		default:
			return fmt.Errorf("failed to update struct field %s value from flag %q", f.Name, flagName)
		}
	}
	return nil
}
