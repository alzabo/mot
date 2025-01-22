package torrent

import (
	"fmt"
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

// Explicitly maps keys to value representations
func (i Info) Values() Values {
	return valueWrapper{
		item: i,
		mapping: map[string]Value{
			"name":       val{value: i.Name},
			"hash":       val{value: i.Hash},
			"state":      val{value: i.State},
			"size":       val{value: i.Size, strFunc: fmtBytes},
			"tags":       val{value: i.Tags},
			"category":   val{value: i.Category},
			"speed_down": val{value: i.Dlspeed, strFunc: fmtRate},
			"speed_up":   val{value: i.Upspeed, strFunc: fmtRate},
			"tracker":    val{value: i.Tracker},
			"progress": val{
				value:   i.Progress,
				strFunc: fmtPercent,
			},
			// date_added
			// date_lastactive
			// downloaded
			// ratio
			// uploaded
			// downloaded
		},
	}
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
