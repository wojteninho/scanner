package scanner_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/wojteninho/scanner/pkg/scanner"
)

func BenchmarkSimpleScannerBulkSize(b *testing.B) {
	for _, filesNumber := range []uint{10, 100, 1000, 10000} {
		for _, bulkSize := range []int{10, 100, 1000, 10000} {
			b.Run(fmt.Sprintf("filesNumber-%d/bulkSize-%d", filesNumber, bulkSize), func(b *testing.B) {
				b.StopTimer()
				dir := NewDirectoryPath(fmt.Sprintf("directory-with-%d-files", filesNumber))
				defer MustNewWorkspace(dir, WithItems(NewWorkspaceFiles("file", filesNumber)...)).Purge()
				scanner := MustScanner(NewBasicScanner(WithDir(dir), WithBulkSize(bulkSize)))
				b.StartTimer()

				for i := 0; i < b.N; i++ {
					for f := range MustScan(scanner.Scan(context.TODO())) {
						f.FileInfo.Name()
					}
				}
			})
		}
	}
}

type MakeScanFn func(directory string) ScanFn
type ScanFn func()

func makeSimpleScannerScanFn(directory string) ScanFn {
	return func() {
		var doneChan = make(chan struct{})

		go func() {
			for item := range MustScan(MustScanner(NewBasicScanner(WithDir(directory))).Scan(context.TODO())) {
				if item.Err != nil {
					panic(item.Err)
				}

				item.FileInfo.Name()
			}

			close(doneChan)
		}()

		<-doneChan
	}
}

func makeFilepathWalkFn(directory string) ScanFn {
	return func() {
		err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			info.Name()
			return nil
		})

		if err != nil {
			panic(err)
		}
	}
}

func makeIoutilReadDirFn(directory string) ScanFn {
	return func() {
		fileInfo, err := ioutil.ReadDir(directory)
		if err != nil {
			panic(err)
		}

		for _, info := range fileInfo {
			info.Name()
		}
	}
}

func makeOsFileReaddirFn(directory string) ScanFn {
	return func() {
		f, err := os.Open(directory)
		if err != nil {
			panic(err)
		}
		fileInfo, err := f.Readdir(-1)
		f.Close()
		if err != nil {
			panic(err)
		}

		for _, file := range fileInfo {
			file.Name()
		}
	}
}

func BenchmarkCompareWithOtherMethodsByScanningFlatDirectory(b *testing.B) {
	var scanStrategies = []struct {
		Name       string
		MakeScanFn MakeScanFn
	}{
		{Name: "BasicScanner.Scan", MakeScanFn: makeSimpleScannerScanFn},
		{Name: "filepath.Walk", MakeScanFn: makeFilepathWalkFn},
		{Name: "ioutil.ReadDir", MakeScanFn: makeIoutilReadDirFn},
		{Name: "os.File.Readdir", MakeScanFn: makeOsFileReaddirFn},
	}

	for _, strategy := range scanStrategies {
		for _, filesNumber := range []uint{10, 50, 100, 200, 500, 1000, 2000, 5000, 10000, 20000, 50000} {
			b.Run(fmt.Sprintf("%s/filesNumber-%d", strategy.Name, filesNumber), func(b *testing.B) {
				b.StopTimer()
				dir := NewDirectoryPath(fmt.Sprintf("directory-with-%d-files", filesNumber))
				defer MustNewWorkspace(dir, WithItems(NewWorkspaceFiles("file", filesNumber)...)).Purge()
				scanFn := strategy.MakeScanFn(dir)
				b.StartTimer()

				for i := 0; i < b.N; i++ {
					scanFn()
				}
			})
		}
	}
}
