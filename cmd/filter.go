package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/alzabo/mot/torrent"
)

type Filter func(torrent.Values) (bool, error)

// TODO: Filters panic when Get(key) returns nil
func parseFilter(expr string) (Filter, error) {
	if strings.Contains(expr, "!=") {
		key, substr, _ := strings.Cut(expr, "!=")
		return func(v torrent.Values) (bool, error) {
			value := v.Get(key)
			if value == nil {
				return false, fmt.Errorf("key %v not found in %v", key, v)
			}
			return value.String() != substr, nil
		}, nil
	} else if strings.Contains(expr, "=") {
		key, substr, _ := strings.Cut(expr, "=")
		return func(v torrent.Values) (bool, error) {
			value := v.Get(key)
			if value == nil {
				return false, fmt.Errorf("key %v not found in %v", key, v)
			}
			return value.String() == substr, nil
		}, nil
	} else if strings.Contains(expr, "!~") {
		key, substr, _ := strings.Cut(expr, "!~")
		return func(v torrent.Values) (bool, error) {
			value := v.Get(key)
			if value == nil {
				return false, fmt.Errorf("key %v not found in %v", key, v)
			}
			r, err := regexp.Compile(substr)
			if err != nil {
				return false, err
			}
			return !r.MatchString(value.String()), nil
		}, nil
	} else if strings.Contains(expr, "~") {
		key, substr, _ := strings.Cut(expr, "~")
		return func(v torrent.Values) (bool, error) {
			value := v.Get(key)
			if value == nil {
				return false, fmt.Errorf("key %v not found in %v", key, v)
			}
			r, err := regexp.Compile(substr)
			if err != nil {
				return false, err
			}
			return r.MatchString(value.String()), nil
		}, nil
	}
	return nil, fmt.Errorf("failed to parse filter expression: %v", expr)

}

func parseFilters(exprs []string) ([]Filter, error) {
	filters := make([]Filter, len(exprs))
	errs := make([]error, len(exprs))
	for i, e := range exprs {
		filters[i], errs[i] = parseFilter(e)
	}
	return filters, errors.Join(errs...)
}
