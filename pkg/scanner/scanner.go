package scanner

import (
	"context"
	"errors"
	"os"
)

var (
	ErrNotDirectory    = errors.New("not directory")
	ErrDirectoryNotSet = errors.New("directory is not set, use WithDir(/path/to/directory) to set it up")
)

type File struct {
	FileInfo os.FileInfo
	Err      error
}

type FileChan chan File

type Scanner interface {
	Scan(ctx context.Context) (FileChan, error)
}

func MustScanner(s Scanner, err error) Scanner {
	if err != nil {
		panic(err)
	}

	return s
}

func MustScan(filesChan FileChan, err error) FileChan {
	if err != nil {
		panic(err)
	}

	return filesChan
}
