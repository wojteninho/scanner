package scanner

import (
	"context"
	"io"
	"os"
)

type SimpleScannerOptionFn func(s *SimpleScanner) error

func WithDir(directory string) SimpleScannerOptionFn {
	return func(s *SimpleScanner) error {
		info, err := os.Stat(directory)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return ErrNotDirectory
		}

		s.directory = directory
		return nil
	}
}

func WithBulkSize(bulkSize int) SimpleScannerOptionFn {
	return func(s *SimpleScanner) error {
		s.bulkSize = bulkSize
		return nil
	}
}

// this guy simply scan single directory
type SimpleScanner struct {
	directory string
	bulkSize  int
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

	if s.directory == "" {
		return nil, ErrDirectoryNotSet
	}

	return &s, nil
}

func (s *SimpleScanner) Scan(ctx context.Context) (FileChan, error) {
	d, err := os.Open(s.directory)
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
