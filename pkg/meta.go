package mot

import (
	"fmt"
	"reflect"

	"github.com/spf13/cobra"
)

func lookupTag[T ~string](field reflect.StructField, tag T) (string, bool) {
	return field.Tag.Lookup(string(tag))
}

func getTag[T ~string](field reflect.StructField, tag T) string {
	v, _ := lookupTag(field, tag)
	return v
}

func kebabCase(s string) string {
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

// AddFlagsForPayload adds a CLI flag to the provided Command for each field
// in the request payload struct.
func AddFlagsForPayload(cmd *cobra.Command, payload any) {
	t := reflect.TypeOf(payload)
	for i := range t.NumField() {
		field := t.Field(i)
		flagName := kebabCase(field.Name)
		usage := getTag(field, tagUsage)
		short := getTag(field, tagShort)

		switch field.Type.Kind() {
		case reflect.Pointer:
			switch field.Type.Elem().Kind() {
			case reflect.String:
				cmd.Flags().StringP(flagName, short, "", usage)
			case reflect.Bool:
				cmd.Flags().BoolP(flagName, short, false, usage)
			case reflect.Int64:
				cmd.Flags().Int64P(flagName, short, 0, usage)
			case reflect.Float64:
				cmd.Flags().Float64P(flagName, short, 0, usage)
			case reflect.Slice:
				switch field.Type.Elem().Elem().Kind() {
				case reflect.String:
					cmd.Flags().StringSliceP(flagName, short, nil, usage)
				case reflect.Struct:
					switch field.Type.Elem().Elem().Name() {
					case "TorrentFile":
						cmd.Flags().VarP(&TorrentFiles{}, flagName, short, usage)
					default:
						panic("unreachable")
					}
				default:
					fmt.Println(field.Type.Elem().Elem().Kind())
					panic("unreachable")
				}
			default:
				panic("unreachable")
			}
		default:
			panic("unreachable")
		}

		if _, ok := lookupTag(field, tagOverrideChangeDetection); ok {
			cmd.Flags().SetAnnotation(flagName, string(tagOverrideChangeDetection), nil)
		}
	}
}

// UpdateFromCmd reads user-supplied flag data from cmd and stores the results
// in the struct pointed to by ptr. If ptr is nil or not a pointer,
// UpdateFromCmd returns an error.
func UpdateFromCmd(cmd *cobra.Command, ptr any) error {
	t := reflect.TypeOf(ptr)
	if t == nil || t.Kind() != reflect.Pointer {
		return fmt.Errorf("cannot update non-pointer value %v", ptr)
	}
	elem := t.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("cannot update non-struct value %v", ptr)
	}
	for i := range elem.NumField() {
		field := elem.Field(i)
		flagName := kebabCase(field.Name)
		flag := cmd.Flag(flagName)

		_, force := flag.Annotations[string(tagOverrideChangeDetection)]
		if !flag.Changed && !force {
			continue
		}

		value := reflect.ValueOf(ptr).Elem().FieldByName(field.Name)
		switch flag.Value.Type() {
		case "bool":
			v, _ := cmd.Flags().GetBool(flagName)
			value.Set(reflect.ValueOf(&v))
		case "float64":
			v, _ := cmd.Flags().GetFloat64(flagName)
			value.Set(reflect.ValueOf(&v))
		case "int64":
			v, _ := cmd.Flags().GetInt64(flagName)
			value.Set(reflect.ValueOf(&v))
		case "string":
			v, _ := cmd.Flags().GetString(flagName)
			value.Set(reflect.ValueOf(&v))
		case "stringSlice":
			v, _ := cmd.Flags().GetStringSlice(flagName)
			value.Set(reflect.ValueOf(&v))
		case "torrentFiles":
			v, _ := cmd.Flags().Lookup(flagName).Value.(*TorrentFiles)
			value.Set(reflect.ValueOf(v))
		default:
			return fmt.Errorf("failed to update struct field %s value from flag %q", field.Name, flagName)
		}
	}
	return nil
}
