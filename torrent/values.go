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
	"strconv"
	"strings"
)

const (
	KiB  int64 = 1024
	MiB  int64 = 2 << 19
	GiB  int64 = 2 << 29
	TiB  int64 = 2 << 39
	KiBF       = float64(KiB)
	MiBF       = float64(KiB)
	GiBF       = float64(GiB)
	TiBF       = float64(TiB)
)

const (
	modDefault = ""
	modRaw     = "raw"
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
	return toString(v.value)
}

func (v val) RawString() string {
	return toString(v.value)
}

func toString(value any) string {
	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', 2, 64)
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func fmtRaw(a any) string {
	return toString(a)
}

func fmtPercent(a any) string {
	switch v := a.(type) {
	case float64:
		return strconv.FormatFloat(v*100, 'f', 2, 64) + "%"
	default:
		panic("unreachable")
	}
}

func fmtIEC(v int64) string {
	var q float64
	var suffix string
	if v < KiB {
		return strconv.FormatInt(v, 10) + " B"
	} else if v < MiB {
		q = float64(v) / KiBF
		suffix = " KiB"
	} else if v < GiB {
		q = float64(v) / MiBF
		suffix = " GiB"
	} else if v < TiB {
		q = float64(v) / GiBF
		suffix = " GiB"
	} else {
		q = float64(v) / TiBF
		suffix = " TiB"
	}
	return strconv.FormatFloat(q, 'f', 2, 64) + suffix
}

func fmtRate(a any) string {
	switch v := a.(type) {
	case int64:
		return fmtIEC(v) + "/s"
	default:
		panic("unreachable")
	}
}

func fmtBytes(a any) string {
	switch v := a.(type) {
	case int64:
		return fmtIEC(v)
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

func parseKey(k string) (string, string) {
	if !strings.ContainsRune(k, '+') {
		return k, modDefault
	}
	key, mod, _ := strings.Cut(k, "+")
	return key, mod
}

func (i Info) Get(k string) (Value, error) {
	return get(i, k, infoKeyMapping, infoFmtMapping)
}

func get[T any](item T, k string, valueMap map[string]func(T) any, fmtMap map[string]func(any) string) (Value, error) {
	key, mod := parseKey(k)
	f := valueMap[key]
	if f == nil {
		return nil, fmt.Errorf("key %q not found", k)
	}
	switch mod {
	case modDefault:
		return val{f(item), fmtMap[key]}, nil
	case modRaw:
		return val{f(item), fmtRaw}, nil
	default:
		return nil, fmt.Errorf("unknown format modifier %q specified in key %s", mod, k)
	}
}

func (i Info) Keys() []string {
	keys := make([]string, 0, len(infoKeyMapping))
	for k := range infoKeyMapping {
		keys = append(keys, k)
	}
	return keys
}
