package scanner

import (
	"context"
	"errors"
	"os"
	"path"
)

var ErrNotDirectory = errors.New("not a directory")

type FileInfo interface {
	os.FileInfo
	PathName() string
}

type File struct {
	os.FileInfo
	pathName string
}

func (f File) PathName() string {
	return path.Join(f.pathName, f.Name())
}

func NewFile(info os.FileInfo, pathName string) File {
	return File{info, pathName}
}

type FileItem struct {
	FileInfo FileInfo
	Err      error
}

type FileItemChan chan FileItem

type Scanner interface {
	Scan(ctx context.Context) (FileItemChan, error)
}

func MustScanner(s Scanner, err error) Scanner {
	if err != nil {
		panic(err)
	}

	return s
}

func MustScan(filesChan FileItemChan, err error) FileItemChan {
	if err != nil {
		panic(err)
	}

	return filesChan
}
