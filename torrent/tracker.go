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

import (
	"fmt"
	"strings"
)

type Trackers []Tracker

type Tracker struct {
	Hash          string `json:"hash,omitempty"`
	Msg           string `json:"msg"`
	NumDownloaded int64  `json:"num_downloaded"`
	NumLeeches    int64  `json:"num_leeches"`
	NumPeers      int64  `json:"num_peers"`
	NumSeeds      int64  `json:"num_seeds"`
	Status        int64  `json:"status"`
	Tier          int64  `json:"tier"`
	URL           string `json:"url"`
}

func (t Tracker) Get(key string) (Value, error) {
	k, mod, _ := strings.Cut(key, "+")

	f := trackerKeyMapping[k]
	if f == nil {
		return val{}, fmt.Errorf("key %q not found", key)
	}

	switch mod {
	case "raw":
		return val{f(t), fmtRaw}, nil
	case "":
		return val{f(t), trackerFmtMapping[k]}, nil
	default:
		return val{}, fmt.Errorf("unknown format modifier %q specified in key %s", mod, key)
	}
}

func (t Tracker) Keys() []string {
	keys := make([]string, 0, len(trackerKeyMapping))
	for k := range trackerKeyMapping {
		keys = append(keys, k)
	}
	return keys
}

var trackerKeyMapping = map[string]func(Tracker) any{
	"hash":       func(t Tracker) any { return t.Hash },
	"message":    func(t Tracker) any { return t.Msg },
	"msg":        func(t Tracker) any { return t.Msg },
	"downloaded": func(t Tracker) any { return t.NumDownloaded },
	"leeches":    func(t Tracker) any { return t.NumLeeches },
	"peers":      func(t Tracker) any { return t.NumPeers },
	"seeds":      func(t Tracker) any { return t.NumSeeds },
	"status":     func(t Tracker) any { return t.Status },
	"tier":       func(t Tracker) any { return t.Tier },
	"url":        func(t Tracker) any { return t.URL },
}

var trackerFmtMapping = map[string]func(any) string{
	"status": func(v any) string {
		m := map[int64]string{
			0: "DISABLED",
			1: "NOT_CONTACTED",
			2: "OK",
			3: "UPDATING",
			4: "NOT_WORKING",
		}
		switch s := v.(type) {
		case int64:
			status, ok := m[s]
			if ok {
				return status
			}
		}
		return "INVALID"
	},
}
