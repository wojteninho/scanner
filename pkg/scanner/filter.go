package scanner

import "context"

type FilterFn func(file FileItem) bool

func FilterRegularFilesFn(f FileItem) bool {
	if f.FileInfo == nil {
		return false
	}

	return f.FileInfo.Mode().IsRegular()
}

func FilterDirectoriesFn(f FileItem) bool {
	if f.FileInfo == nil {
		return false
	}

	return f.FileInfo.Mode().IsDir()
}

type FilterScanner struct {
	scanner  Scanner
	filterFn FilterFn
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
			if !s.filterFn(item) {
				continue
			}
			fileChan <- item
		}
	}()

	return fileChan, nil
}

func NewFilterRegularFilesScanner(scanner Scanner) Scanner {
	return NewFilterScanner(scanner, FilterRegularFilesFn)
}

func NewFilterDirectoriesScanner(scanner Scanner) Scanner {
	return NewFilterScanner(scanner, FilterDirectoriesFn)
}

func NewFilterScanner(scanner Scanner, filterFn FilterFn) Scanner {
	return &FilterScanner{scanner, filterFn}
}
