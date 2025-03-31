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

const (
	KiB int64 = 1024
	MiB int64 = 2 << 19
	GiB int64 = 2 << 29
	TiB int64 = 2 << 39
)

type ValuesCollection interface {
	ValuesSlice()
}

type Values interface {
	Get(string) (Value, error)
	Keys() []string
}

type Value interface {
	Raw() any
	String() string
	RawString() string
}

type val struct {
	value   any
	strFunc func(any) string
}

func (v val) Raw() any {
	return v.value
}

func (v val) String() string {
	if v.strFunc != nil {
		return v.strFunc(v.value)
	}
	f := "%v"
	switch v.value.(type) {
	case int, int64:
		f = "%d"
	case float64:
		f = "%.2f"
	case string, []byte:
		f = "%s"
	}
	return fmt.Sprintf(f, v.value)
}

func (v val) RawString() string {
	return fmt.Sprintf("%v", v.value)
}

func fmtRaw(a any) string {
	return fmt.Sprintf("%v", a)
}

func fmtPercent(a any) string {
	switch v := a.(type) {
	case float64:
		return fmt.Sprintf("%.2f%%", v*100)
	default:
		panic("unreachable")
	}
}

func fmtRate(a any) string {
	switch v := a.(type) {
	case int64:
		if v < KiB {
			return fmt.Sprintf("%d B/s", v)
		} else if v < MiB {
			return fmt.Sprintf("%.2f KiB/s", float64(v)/float64(KiB))
		} else if v < GiB {
			return fmt.Sprintf("%.2f MiB/s", float64(v)/float64(MiB))
		} else if v < TiB {
			return fmt.Sprintf("%.2f GiB/s", float64(v)/float64(GiB))
		} else {
			return fmt.Sprintf("%.2f TiB/s", float64(v)/float64(TiB))
		}
	default:
		panic("unreachable")
	}
}

func fmtBytes(a any) string {
	switch v := a.(type) {
	case int64:
		if v < KiB {
			return fmt.Sprintf("%d B", v)
		} else if v < MiB {
			return fmt.Sprintf("%.2f KiB", float64(v)/float64(KiB))
		} else if v < GiB {
			return fmt.Sprintf("%.2f MiB", float64(v)/float64(MiB))
		} else if v < TiB {
			return fmt.Sprintf("%.2f GiB", float64(v)/float64(GiB))
		} else {
			return fmt.Sprintf("%.2f TiB", float64(v)/float64(TiB))
		}
	default:
		panic("unreachable")
	}
}

var infoKeyMapping = map[string]func(Info) any{
	"name":       func(i Info) any { return i.Name },
	"hash":       func(i Info) any { return i.Hash },
	"state":      func(i Info) any { return i.State },
	"size":       func(i Info) any { return i.Size },
	"tags":       func(i Info) any { return i.Tags },
	"category":   func(i Info) any { return i.Category },
	"speed_down": func(i Info) any { return i.Dlspeed },
	"speed_up":   func(i Info) any { return i.Upspeed },
	"tracker":    func(i Info) any { return i.Tracker },
	"progress":   func(i Info) any { return i.Progress },
	"ratio":      func(i Info) any { return i.Ratio },
	"downloaded": func(i Info) any { return i.Downloaded },
	"uploaded":   func(i Info) any { return i.Downloaded },
	"added":      func(i Info) any { return i.AddedOn },
	// date_added
	// date_lastactive
}

var infoFmtMapping = map[string]func(any) string{
	"size":       fmtBytes,
	"speed_down": fmtRate,
	"speed_up":   fmtRate,
	"progress":   fmtPercent,
	"downloaded": fmtBytes,
	"uploaded":   fmtBytes,
	// date_added
	// date_lastactive
}

func (i Info) Get(key string) (Value, error) {
	k, mod, _ := strings.Cut(key, "+")

	f := infoKeyMapping[k]
	if f == nil {
		return val{}, fmt.Errorf("key %q not found", key)
	}

	switch mod {
	case "raw":
		return val{f(i), fmtRaw}, nil
	case "":
		return val{f(i), infoFmtMapping[k]}, nil
	default:
		return val{}, fmt.Errorf("unknown format modifier %q specified in key %s", mod, key)
	}
}

func (i Info) Keys() []string {
	keys := make([]string, 0, len(infoKeyMapping))
	for k := range infoKeyMapping {
		keys = append(keys, k)
	}
	return keys
}
