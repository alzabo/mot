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

package torrent

type Files []File

type File struct {
	Hash         string  `json:"hash,omitempty"`
	Availability float64 `json:"availability"`
	Index        int64   `json:"index"`
	IsSeed       bool    `json:"is_seed"`
	Name         string  `json:"name"`
	PieceRange   []int64 `json:"piece_range"`
	Priority     int64   `json:"priority"`
	Progress     float64 `json:"progress"`
	Size         int64   `json:"size"`
}

var fileKeyMapping = map[string]func(File) any{
	"name":     func(f File) any { return f.Name },
	"hash":     func(f File) any { return f.Hash },
	"size":     func(f File) any { return f.Size },
	"progress": func(f File) any { return f.Progress },
}

var fileFmtMapping = map[string]func(any) string{
	"size":     fmtBytes,
	"progress": fmtPercent,
}

func (f File) Keys() []string {
	keys := make([]string, 0, len(fileKeyMapping))
	for k := range fileKeyMapping {
		keys = append(keys, k)
	}
	return keys
}

func (f File) Get(key string) (Value, error) {
	return get(f, key, fileKeyMapping, fileFmtMapping)
}
