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

package parser

import (
	"fmt"
	"strconv"

	"github.com/alzabo/mot/torrent"
)

type val struct {
	value any
}

func (v val) String() string {
	switch value := v.value.(type) {
	case string:
		return value
	case int64:
		return fmt.Sprintf("%v", value)
	case bool:
		return strconv.FormatBool(value)
	default:
		panic("unreachable")
	}
}

func (v val) Raw() any {
	return v.value
}

func (v val) RawString() string {
	return v.String()
}

var keys = []string{
	"announce",
	"name",
	"comment",
	"hash",
	"size",
	"private",
}

var keyMap = map[string]func(Torrent) val{
	"announce": func(t Torrent) val { return val{t.Announce} },
	"name":     func(t Torrent) val { return val{t.Name} },
	"comment":  func(t Torrent) val { return val{t.Comment} },
	"hash":     func(t Torrent) val { return val{t.Hash} },
	"size":     func(t Torrent) val { return val{t.Size} },
	"private":  func(t Torrent) val { return val{t.Private} },
}

func (t Torrent) Get(k string) (torrent.Value, error) {
	f, ok := keyMap[k]
	if !ok {
		return nil, fmt.Errorf("key %s not found", k)
	}
	return f(t), nil
}

func (t Torrent) Keys() []string {
	return keys
}
