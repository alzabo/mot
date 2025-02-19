package torrent

import (
	"fmt"
	"strings"
)

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
	k, mod, _ := strings.Cut(key, "+")

	fn := fileKeyMapping[k]
	if fn == nil {
		return val{}, fmt.Errorf("key %q not found", key)
	}

	switch mod {
	case "raw":
		return val{fn(f), fmtRaw}, nil
	case "":
		return val{fn(f), fileFmtMapping[k]}, nil
	default:
		return val{}, fmt.Errorf("unknown format modifier %q specified in key %s", mod, key)
	}
}
