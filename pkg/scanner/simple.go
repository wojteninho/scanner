package scanner

import (
	"context"
	"io"
	"os"
)

type SimpleScannerOptionFn func(s *SimpleScanner) error

func WithBulkSize(bulkSize int) SimpleScannerOptionFn {
	return func(s *SimpleScanner) error {
		s.bulkSize = bulkSize
		return nil
	}
}

// this guy simply scan single directory
type SimpleScanner struct {
	bulkSize int
}

func NewSimpleScanner(options ...SimpleScannerOptionFn) (*SimpleScanner, error) {
	s := SimpleScanner{
		bulkSize: 20,
	}

	for _, option := range options {
		if err := option(&s); err != nil {
			return nil, err
		}
	}

	return &s, nil
}

func (s *SimpleScanner) Scan(ctx context.Context, directory string) (FileChan, error) {
	fi, err := os.Stat(directory)
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		return nil, ErrNotDirectory
	}

	d, err := os.Open(directory)
	if err != nil {
		return nil, err
	}

	fileChan := make(FileChan)

	go func() {
		defer d.Close()
		defer close(fileChan)

		for {
			bulk, err := d.Readdir(s.bulkSize)

			if err != nil {
				if err == io.EOF {
					break
				}

				fileChan <- File{nil, err}
			}

			for _, file := range bulk {
				fileChan <- File{file, nil}
			}
		}
	}()

	return fileChan, nil
}
