package scanner

import "context"

type RecursiveScanner struct {
}

func (cs *RecursiveScanner) Scan(ctx context.Context, directory string) (FileChan, error) {
	// TODO
	return nil, nil
}
