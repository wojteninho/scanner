package scanner

import (
	"context"
	"errors"
	"os"
)

var (
	ErrNotDirectory = errors.New("not directory")
)

type File struct {
	FileInfo os.FileInfo
	Err      error
}

type FileChan chan File

type Scanner interface {
	Scan(ctx context.Context, directory string) (FileChan, error)
}

func MustScanner(s Scanner, err error) Scanner {
	if err != nil {
		panic(err)
	}

	return s
}
