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

type Logs []Log

type Log struct {
	ID        int64  `json:"id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"` // TODO Note: switched from milliseconds to seconds in v4.5.0
	Type      int64  `json:"type"`
}

func (l Log) Get(key string) (Value, error) {
	k, mod, _ := strings.Cut(key, "+")

	f := logKeyMapping[k]
	if f == nil {
		return val{}, fmt.Errorf("key %q not found", key)
	}

	switch mod {
	case "raw":
		return val{f(l), fmtRaw}, nil
	case "":
		return val{f(l), logFmtMapping[k]}, nil
	default:
		return val{}, fmt.Errorf("unknown format modifier %q specified in key %s", mod, key)
	}
}

func (l Log) Keys() []string {
	keys := make([]string, 0, len(logKeyMapping))
	for k := range logKeyMapping {
		keys = append(keys, k)
	}
	return keys
}

var logKeyMapping = map[string]func(Log) any{
	"id":        func(l Log) any { return l.ID },
	"message":   func(l Log) any { return l.Message },
	"msg":       func(l Log) any { return l.Message },
	"timestamp": func(l Log) any { return l.Timestamp },
	"type":      func(l Log) any { return l.Type },
}

var logFmtMapping = map[string]func(any) string{
	"type": func(v any) string {
		// Type of the message: Log::NORMAL: 1, Log::INFO: 2, Log::WARNING: 4, Log::CRITICAL: 8
		m := map[int64]string{
			1: "NORMAL",
			2: "INFO",
			4: "WARNING",
			8: "CRITICAL",
		}
		value, ok := v.(int64)
		if !ok {
			return "INVALID"
		}
		status, ok := m[value]
		if ok {
			return status
		}
		return "UNKNOWN"
	},
}
