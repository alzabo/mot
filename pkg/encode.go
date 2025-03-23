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
	rt := reflect.TypeOf(r)
	for i := range rt.NumField() {
		v := rv.Field(i)
		if v.IsNil() {
			continue
		}

		form := rt.Field(i).Tag.Get("param")
		switch v.Elem().Kind() {
		case reflect.Bool:
			values[form] = []string{fmt.Sprintf("%v", v.Elem().Bool())}
		case reflect.Int64:
			values[form] = []string{fmt.Sprintf("%d", v.Elem().Int())}
		case reflect.Float64:
			values[form] = []string{fmt.Sprintf("%.2f", v.Elem().Float())}
		case reflect.String:
			values[form] = []string{v.Elem().String()}
		case reflect.Slice:
			val, ok := (v.Elem().Interface()).([]string)
			if !ok {
				return nil, fmt.Errorf("failed to convert value from field %s", v)
			}
			join := rt.Field(i).Tag.Get("join")
			values[form] = []string{strings.Join(val, join)}
		default:
			return nil, fmt.Errorf("failed to encode field of type %s", v.Elem().Kind())
		}
	}
	return values, nil
}
