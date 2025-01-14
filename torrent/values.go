package torrent

import (
	"fmt"
	"slices"
)

const (
	KiB int64 = 1024
	MiB int64 = 2 << 19
	GiB int64 = 2 << 29
	TiB int64 = 2 << 39
)

type Values interface {
	Get(string) Value
	Keys() []string
}

type Value interface {
	String() string
}

type InfoValues struct {
	info         Info
	stringValues map[string]Value
}

func (v InfoValues) Get(s string) Value {
	val, ok := v.stringValues[s]
	if !ok {
		return nil
	}
	return val
}

func (v InfoValues) Keys() []string {
	keys := []string{}
	for k := range v.stringValues {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

type stringValue struct {
	v string
}

func (v stringValue) String() string {
	return v.v
}

type percentValue struct {
	v float64
}

func (v percentValue) String() string {
	return fmt.Sprintf("%.2f%%", v.v*100)
}

type bytesValue struct {
	v int64
}

func (v bytesValue) String() string {
	if v.v < KiB {
		return fmt.Sprintf("%d B", v.v)
	} else if v.v < MiB {
		return fmt.Sprintf("%.2f KiB", float64(v.v)/float64(KiB))
	} else if v.v < GiB {
		return fmt.Sprintf("%.2f MiB", float64(v.v)/float64(MiB))
	} else if v.v < TiB {
		return fmt.Sprintf("%.2f GiB", float64(v.v)/float64(GiB))
	} else {
		return fmt.Sprintf("%.2f TiB", float64(v.v)/float64(TiB))
	}
}

type rateValue struct {
	v int64
}

func (v rateValue) String() string {
	if v.v < KiB {
		return fmt.Sprintf("%d B/s", v.v)
	} else if v.v < MiB {
		return fmt.Sprintf("%.2f KiB/s", float64(v.v)/float64(KiB))
	} else if v.v < GiB {
		return fmt.Sprintf("%.2f MiB/s", float64(v.v)/float64(MiB))
	} else if v.v < TiB {
		return fmt.Sprintf("%.2f GiB/s", float64(v.v)/float64(GiB))
	} else {
		return fmt.Sprintf("%.2f TiB/s", float64(v.v)/float64(TiB))
	}
}

// Explicitly maps keys to value representations
func (i Info) Values() Values {
	return InfoValues{
		info: i,
		stringValues: map[string]Value{
			"name":       stringValue{i.Name},
			"hash":       stringValue{i.Hash},
			"state":      stringValue{i.State},
			"progress":   percentValue{i.Progress},
			"size":       bytesValue{i.Size},
			"size_raw":   stringValue{fmt.Sprintf("%d", i.Size)},
			"tags":       stringValue{i.Tags},
			"category":   stringValue{i.Category},
			"speed_down": rateValue{i.Dlspeed},
			"speed_up":   rateValue{i.Upspeed},
			"tracker":    stringValue{i.Tracker},
			// dateadded
			// datelastactive
			// downloaded
			// ratio
			// uploaded
			// downloaded
		},
	}
}

type valueWrapper struct {
	item         any
	stringValues map[string]Value
}

func (v valueWrapper) Get(s string) Value {
	val, ok := v.stringValues[s]
	if !ok {
		return nil
	}
	return val
}

func (v valueWrapper) Keys() []string {
	keys := []string{}
	for k := range v.stringValues {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func (f File) Values(vals map[string]string) Values {
	w := valueWrapper{
		item: f,
		stringValues: map[string]Value{
			"name":     stringValue{f.Name},
			"size":     bytesValue{f.Size},
			"size_raw": stringValue{fmt.Sprintf("%d", f.Size)},
			"progress": percentValue{f.Progress},
			"index":    stringValue{fmt.Sprintf("%d", f.Index)},
		},
	}
	for k, v := range vals {
		w.stringValues[k] = stringValue{v}
	}
	return w
}
