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

type Values interface {
	Get(string) (Value, error)
	Keys() []string
}

type Value interface {
	Raw() any
	String() string
	RawString() string
}

type valueWrapper struct {
	item    any
	mapping map[string]Value
}

func (v valueWrapper) Get(s string) (Value, error) {
	val, ok := v.mapping[s]
	if !ok {
		return nil, fmt.Errorf("key %v not found in %v", s, v.mapping)
	}
	return val, nil
}

func (v valueWrapper) Keys() []string {
	keys := make([]string, 0, len(v.mapping))
	for k := range v.mapping {
		keys = append(keys, k)
	}
	return keys
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

var infoValuesMapping = map[string]func(Info) Value{
	"name":       func(i Info) Value { return val{value: i.Name} },
	"hash":       func(i Info) Value { return val{value: i.Hash} },
	"state":      func(i Info) Value { return val{value: i.State} },
	"size":       func(i Info) Value { return val{value: i.Size, strFunc: fmtBytes} },
	"tags":       func(i Info) Value { return val{value: i.Tags} },
	"category":   func(i Info) Value { return val{value: i.Category} },
	"speed_down": func(i Info) Value { return val{value: i.Dlspeed, strFunc: fmtRate} },
	"speed_up":   func(i Info) Value { return val{value: i.Upspeed, strFunc: fmtRate} },
	"tracker":    func(i Info) Value { return val{value: i.Tracker} },
	"progress": func(i Info) Value {
		return val{
			value:   i.Progress,
			strFunc: fmtPercent,
		}
	},
	"ratio":      func(i Info) Value { return val{value: i.Ratio} },
	"downloaded": func(i Info) Value { return val{value: i.Downloaded, strFunc: fmtBytes} },
	"uploaded":   func(i Info) Value { return val{value: i.Downloaded, strFunc: fmtBytes} },
	// date_added
	// date_lastactive

}

// Values returns Values object containing requested keys
func (i Info) Values(keys []string) Values {
	vals := valueWrapper{
		item:    i,
		mapping: map[string]Value{},
	}
	for _, key := range keys {
		k, mod, _ := strings.Cut(key, "+")
		f := infoValuesMapping[k]
		if f == nil {
			continue
		}
		vals.mapping[k] = f(i)

		switch mod {
		case "raw":
			vals.mapping[key] = val{
				value:   vals.mapping[k].Raw(),
				strFunc: fmtRaw,
			}
		}
	}
	return vals
}

func (f File) Values(vals map[string]string) Values {
	w := valueWrapper{
		item: f,
		mapping: map[string]Value{
			"name":     val{value: f.Name},
			"size":     val{value: f.Size, strFunc: fmtBytes},
			"progress": val{value: f.Progress, strFunc: fmtPercent},
			"index":    val{value: f.Index},
		},
	}
	for k, v := range vals {
		w.mapping[k] = val{value: v}
	}
	return w
}

func (t Tracker) Values(vals map[string]string) Values {
	w := valueWrapper{
		item: t,
		mapping: map[string]Value{
			"message": val{value: t.Msg},
			"url":     val{value: t.URL},
			"status": val{value: t.Status, strFunc: func(v any) string {
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
			}},
			"peers":     val{value: t.NumPeers},
			"seeds":     val{value: t.NumSeeds},
			"leeches":   val{value: t.NumLeeches},
			"downloads": val{value: t.NumDownloaded},
		},
	}
	for k, v := range vals {
		w.mapping[k] = val{value: v}
	}
	return w
}
