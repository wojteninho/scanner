package scanner

import (
	"context"
	"fmt"
)

type DebugFn func(item FileItem)

type DebugScanner struct {
	scanner Scanner
	debugFn DebugFn
}

func (s *DebugScanner) Scan(ctx context.Context) (FileItemChan, error) {
	innerFileChan, err := s.scanner.Scan(ctx)
	if err != nil {
		return nil, err
	}

	fileChan := make(FileItemChan)

	go func() {
		defer close(fileChan)

		for item := range innerFileChan {
			s.debugFn(item)
			fileChan <- item
		}
	}()

	return fileChan, nil
}

func NewPrintPathNameDebugScanner(scanner Scanner) Scanner {
	return NewDebugScanner(scanner, func(item FileItem) {
		fmt.Println(item.String())
	})
}

func NewDebugScanner(scanner Scanner, debugFn DebugFn) Scanner {
	return &DebugScanner{scanner, debugFn}
}
