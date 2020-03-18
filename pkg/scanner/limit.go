package scanner

import (
	"context"
)

type LimitFn func(item FileItem)

type LimitScanner struct {
	scanner Scanner
	limit   Limit
}

type Limit int

func NewLimitScanner(scanner Scanner, limit Limit) *LimitScanner {
	return &LimitScanner{scanner: scanner, limit: limit}
}

func (ls *LimitScanner) Scan(ctx context.Context) (FileItemChan, error) {
	innerFileChan, err := ls.scanner.Scan(ctx)

	if err != nil {
		return nil, err
	}

	fileChan := make(FileItemChan, ls.limit)

	go func() {
		defer close(fileChan)

		var count Limit
		for item := range innerFileChan {
			count++
			if count > ls.limit {
				continue
			}
			fileChan <- item
		}
	}()

	return fileChan, nil
}
