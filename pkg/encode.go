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

package mot

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

func encode(r any) (string, error) {
	values, err := makeValues(r)
	if err != nil {
		return "", err
	}
	return values.Encode(), nil
}

func makeValues(r any) (url.Values, error) {
	values := url.Values{}
	rv := reflect.ValueOf(r)
	t := reflect.TypeOf(r)
	for i := range t.NumField() {
		value := rv.Field(i)
		if value.IsNil() {
			continue
		}
		if value.Kind() != reflect.Pointer {
			return nil, fmt.Errorf("value %v must be a pointer. found %v", value, value.Kind())
		}

		field := t.Field(i)
		form := getTag(field, tagParam)

		elem := value.Elem()
		switch elem.Kind() {
		case reflect.Bool:
			values[form] = []string{strconv.FormatBool(elem.Bool())}
		case reflect.Int64:
			values[form] = []string{strconv.FormatInt(elem.Int(), 10)}
		case reflect.Float64:
			values[form] = []string{strconv.FormatFloat(elem.Float(), 'f', 2, 64)}
		case reflect.String:
			values[form] = []string{elem.String()}
		case reflect.Slice:
			val, ok := (elem.Interface()).([]string)
			if !ok {
				return nil, fmt.Errorf("failed to convert value from field %s", value)
			}
			join := getTag(field, tagJoin)
			values[form] = []string{strings.Join(val, join)}
		default:
			return nil, fmt.Errorf("failed to encode field of type %s", elem.Kind())
		}
	}
	return values, nil
}
