package scanner

import (
	"context"
)

type MultiScanner struct {
	scanners []Scanner
}

func NewMultiScanner(scanners ...Scanner) *MultiScanner {
	return &MultiScanner{scanners: scanners}
}

func (ms *MultiScanner) Scan(ctx context.Context, directory string) (FileChan, error) {
	// TODO
	return nil, nil
}