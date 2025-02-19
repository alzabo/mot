package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/alzabo/mot/torrent"
)

type Filter func(torrent.Values) (bool, error)

type Filters []Filter

func (fs Filters) All(item torrent.Values) (bool, error) {
	for _, f := range fs {
		ok, err := f(item)
		if !ok || err != nil {
			return ok, err
		}
	}
	return true, nil
}

func parseFilter(expr string) (Filter, error) {
	if strings.Contains(expr, "!=") {
		key, substr, _ := strings.Cut(expr, "!=")
		return func(v torrent.Values) (bool, error) {
			value, err := v.Get(key)
			if err != nil {
				return false, fmt.Errorf("key %v not found in %v", key, v)
			}
			return value.String() != substr, nil
		}, nil
	} else if strings.Contains(expr, "=") {
		key, substr, _ := strings.Cut(expr, "=")
		return func(v torrent.Values) (bool, error) {
			value, err := v.Get(key)
			if err != nil {
				return false, fmt.Errorf("key %v not found in %v", key, v)
			}
			return value.String() == substr, nil
		}, nil
	} else if strings.Contains(expr, "!~") {
		k, e, _ := strings.Cut(expr, "!~")
		r, err := regexp.Compile(e)
		if err != nil {
			return nil, err
		}
		return func(v torrent.Values) (bool, error) {
			value, err := v.Get(k)
			if err != nil {
				return false, fmt.Errorf("key %v not found in %v", k, v)
			}
			return !r.MatchString(value.String()), nil
		}, nil
	} else if strings.Contains(expr, "~") {
		k, e, _ := strings.Cut(expr, "~")
		r, err := regexp.Compile(e)
		if err != nil {
			return nil, err
		}
		return func(v torrent.Values) (bool, error) {
			value, err := v.Get(k)
			if err != nil {
				return false, fmt.Errorf("key %v not found in %v", k, v)
			}
			return r.MatchString(value.String()), nil
		}, nil
	}
	return nil, fmt.Errorf("failed to parse filter expression: %v", expr)

}

func parseFilters(exprs []string) (Filters, error) {
	filters := make(Filters, len(exprs))
	errs := make([]error, len(exprs))
	for i, e := range exprs {
		filters[i], errs[i] = parseFilter(e)
	}
	return filters, errors.Join(errs...)
}
