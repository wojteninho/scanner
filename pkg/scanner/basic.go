package scanner

import (
	"context"
	"io"
	"os"
)

type BasicScannerOptionFn func(s *BasicScanner) error

func WithDir(directory string) BasicScannerOptionFn {
	return func(s *BasicScanner) error {
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

func WithBulkSize(bulkSize int) BasicScannerOptionFn {
	return func(s *BasicScanner) error {
		s.bulkSize = bulkSize
		return nil
	}
}

type BasicScanner struct {
	directory string
	bulkSize  int
}

func NewBasicScanner(options ...BasicScannerOptionFn) (Scanner, error) {
	s := BasicScanner{
		bulkSize: 20,
	}

	for _, option := range options {
		if err := option(&s); err != nil {
			return nil, err
		}
	}

	return &s, nil
}

func (s *BasicScanner) Scan(ctx context.Context) (FileItemChan, error) {
	fileChan := make(FileItemChan)
	if s.directory == "" {
		defer close(fileChan)
		return fileChan, nil
	}

	d, err := os.Open(s.directory)
	if err != nil {
		return nil, err
	}

	go func() {
		defer d.Close()
		defer close(fileChan)

		for {
			bulk, err := d.Readdir(s.bulkSize)

			if err != nil {
				if err == io.EOF {
					break
				}

				fileChan <- FileItem{nil, err}
			}

			for _, info := range bulk {
				fileChan <- FileItem{NewFile(info, s.directory), nil}
			}
		}
	}()

	return fileChan, nil
}
