package scanner

import (
	"context"
	"os"
	"runtime"
)

type RecursiveScannerOptionFn func(s *RecursiveScanner) error

func WithDirectories(directories ...string) RecursiveScannerOptionFn {
	return func(s *RecursiveScanner) error {
		var (
			uniqueDirectories    []string
			uniqueDirectoriesMap = make(map[string]bool)
		)

		for _, d := range directories {
			info, err := os.Stat(d)
			if err != nil {
				return err
			}

			if !info.IsDir() {
				return ErrNotDirectory
			}

			if _, exists := uniqueDirectoriesMap[d]; exists {
				continue
			}

			uniqueDirectoriesMap[d] = true
			uniqueDirectories = append(uniqueDirectories, d)
		}

		s.directories = uniqueDirectories

		return nil
	}
}

func WithWorkers(workers uint) RecursiveScannerOptionFn {
	return func(s *RecursiveScanner) error {
		s.workers = workers
		return nil
	}
}

type RecursiveScanner struct {
	directories []string
	workers     uint
}

func NewRecursiveScanner(options ...RecursiveScannerOptionFn) (Scanner, error) {
	s := RecursiveScanner{
		workers: uint(runtime.NumCPU()),
	}

	for _, option := range options {
		if err := option(&s); err != nil {
			return nil, err
		}
	}

	return &s, nil
}

func (s *RecursiveScanner) Scan(ctx context.Context) (FileItemChan, error) {
	outFileItemChan := make(FileItemChan)
	if len(s.directories) == 0 {
		defer close(outFileItemChan)
		return outFileItemChan, nil
	}

	var (
		directoriesToScanQueue []string
		workers                = make(map[string]interface{})
		finishedScanningChan   = make(chan string)
		scheduleScanningChan   = make(chan string, s.workers)
		doneChan               = make(chan struct{})
	)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case dir := <-scheduleScanningChan:
				// spawn scanning immediately if possible
				if uint(len(workers)) < s.workers {
					workers[dir] = dir
					go doScan(ctx, dir, outFileItemChan, finishedScanningChan, scheduleScanningChan)
					continue
				}

				// append to the queue
				directoriesToScanQueue = append(directoriesToScanQueue, dir)
			case scannedDir := <-finishedScanningChan:
				delete(workers, scannedDir)

				// still something in the queue?
				if len(directoriesToScanQueue) > 0 {
					dir := directoriesToScanQueue[0]
					directoriesToScanQueue = directoriesToScanQueue[1:]
					workers[dir] = dir
					go doScan(ctx, dir, outFileItemChan, finishedScanningChan, scheduleScanningChan)
					continue
				}

				// no workers & nothing in the queue? done!
				if len(workers) == 0 {
					close(doneChan)
					return
				}
			}
		}
	}()

	// shutdown
	go func() {
		select {
		case <-ctx.Done():
		case <-doneChan:
		}

		close(finishedScanningChan)
		close(scheduleScanningChan)
		close(outFileItemChan)
	}()

	// schedule initial directories scanning
	go func() {
		for _, d := range s.directories {
			scheduleScanningChan <- d
		}
	}()

	return outFileItemChan, nil
}

func doScan(ctx context.Context, dir string, outFileItemChan FileItemChan, finishedScanningChan, scheduleScanningChan chan string) {
	defer func() { finishedScanningChan <- dir }()

	defer func() {
		if err := recover(); err != nil {
			outFileItemChan <- FileItem{nil, err.(error)}
		}
	}()

	for item := range MustScan(MustScanner(NewBasicScanner(WithDir(dir))).Scan(ctx)) {
		if item.FileInfo.IsDir() {
			scheduleScanningChan <- item.FileInfo.PathName()
		}

		outFileItemChan <- item
	}
}
