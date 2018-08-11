package scanner

import (
	"context"
	"regexp"
	"strings"
)

const (
	nameExtension          = "ExtensionFilter"
	nameRegExpFilter       = "RegExpFilter"
	nameAndFilter          = "AndFilter"
	nameOrFilter           = "OrFilter"
	nameRegularFilesFilter = "RegularFilesFilter"
	nameDirectoriesFilter  = "DirectoriesFilter"
	nameErrFilter          = "ErrFilter"
)

var (
	DirectoriesFilter  = MakeNamedFilter(FilterFn(filterDirectoriesFn), nameDirectoriesFilter)
	RegularFilesFilter = MakeNamedFilter(FilterFn(filterRegularFilesFn), nameRegularFilesFilter)
	ErrFilter          = MakeNamedFilter(FilterFn(filterErrorsFn), nameErrFilter)
)

type Filter interface {
	Match(file FileItem) bool
}

type NamedFilter struct {
	Filter
	name string
}

func (f *NamedFilter) Match(file FileItem) bool {
	return f.Filter.Match(file)
}

func (f *NamedFilter) Name() string {
	return f.name
}

func MakeNamedFilter(filter Filter, name string) *NamedFilter {
	return &NamedFilter{filter, name}
}

type FilterFn func(file FileItem) bool

func (f FilterFn) Match(file FileItem) bool {
	return f(file)
}

func ExtensionFilter(ext string) Filter {
	return MakeNamedFilter(FilterFn(func(file FileItem) bool {
		if file.FileInfo == nil {
			return false
		}

		return strings.HasSuffix(file.FileInfo.Name(), ext)
	}), nameExtension)
}

func RegExpFilter(r *regexp.Regexp) Filter {
	return MakeNamedFilter(FilterFn(func(file FileItem) bool {
		if file.FileInfo == nil {
			return false
		}

		return r.MatchString(file.FileInfo.Name())
	}), nameRegExpFilter)
}

func AndFilter(filters ...Filter) Filter {
	return MakeNamedFilter(FilterFn(func(file FileItem) bool {
		if len(filters) == 0 {
			return true
		}

		for _, f := range filters {
			if !f.Match(file) {
				return false
			}
		}

		return true
	}), nameAndFilter)
}

func OrFilter(filters ...Filter) Filter {
	return MakeNamedFilter(FilterFn(func(file FileItem) bool {
		if len(filters) == 0 {
			return true
		}

		for _, f := range filters {
			if f.Match(file) {
				return true
			}
		}

		return false
	}), nameOrFilter)
}

func filterRegularFilesFn(f FileItem) bool {
	if f.FileInfo == nil {
		return false
	}

	return f.FileInfo.Mode().IsRegular()
}

func filterDirectoriesFn(f FileItem) bool {
	if f.FileInfo == nil {
		return false
	}

	return f.FileInfo.Mode().IsDir()
}

func filterErrorsFn(f FileItem) bool {
	return f.Err != nil
}

type FilterScanner struct {
	scanner Scanner
	filter  Filter
}

func (s *FilterScanner) Scan(ctx context.Context) (FileItemChan, error) {
	innerFileChan, err := s.scanner.Scan(ctx)
	if err != nil {
		return nil, err
	}

	fileChan := make(FileItemChan)

	go func() {
		defer close(fileChan)

		for item := range innerFileChan {
			if !s.filter.Match(item) {
				continue
			}
			fileChan <- item
		}
	}()

	return fileChan, nil
}

func NewFilterRegularFilesScanner(scanner Scanner) Scanner {
	return NewFilterScanner(scanner, RegularFilesFilter)
}

func NewFilterDirectoriesScanner(scanner Scanner) Scanner {
	return NewFilterScanner(scanner, DirectoriesFilter)
}

func NewFilterScanner(scanner Scanner, filter Filter) Scanner {
	return &FilterScanner{scanner, filter}
}
