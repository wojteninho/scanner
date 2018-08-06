package scanner

import (
	"context"
	"sync"
)

type MultiScanner struct {
	scanners []Scanner
}

func NewMultiScanner(scanners ...Scanner) Scanner {
	return &MultiScanner{scanners: scanners}
}

func (ms *MultiScanner) Scan(ctx context.Context) (FileItemChan, error) {
	var (
		fileChan = make(FileItemChan)
		wg       sync.WaitGroup
	)

	for _, s := range ms.scanners {
		ch, err := s.Scan(ctx)
		if err != nil {
			return nil, err
		}

		wg.Add(1)
		go func(ch FileItemChan) {
			defer wg.Done()

			for item := range ch {
				fileChan <- item
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(fileChan)
	}()

	return fileChan, nil
}
