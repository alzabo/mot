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

package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/alzabo/mot/filter"
	"github.com/alzabo/mot/torrent"
)

type Table[T torrent.Values] struct {
	Columns []string
	Filters filter.Filters
	Watch   bool
	Headers bool
	Sleep   time.Duration
	Writer  *TableWriter
}

func (t Table[T]) Print(getter func() ([]T, error)) error {

	fields := make([]string, len(t.Columns))

	if t.Headers {
		// TODO: Print headers again every N lines when watching?
		t.Writer.WriteFunc(t.Columns, strings.ToUpper)
	}

	for {
		items, err := getter()
		if err != nil {
			return err
		}
		for _, item := range items {
			ok, err := t.Filters.All(item)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			for i, key := range t.Columns {
				v, err := item.Get(key)
				if err != nil {
					return fmt.Errorf("key %v not found in object; available keys: [%s]", key, strings.Join(item.Keys(), ","))
				}
				fields[i] = v.String()
			}
			t.Writer.WriteOnce(fields)
		}

		t.Writer.Flush()

		if !t.Watch {
			if len(items) == 0 {
				t.Writer.Write([]string{"No items found."})
			}
			break
		}
		time.Sleep(t.Sleep)
	}
	return nil
}
