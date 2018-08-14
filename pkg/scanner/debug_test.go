package scanner_test

import (
	"context"
	"testing"

	"errors"

	. "github.com/onsi/gomega"
	"github.com/wojteninho/scanner/pkg/scanner"
)

func TestDebugScanner_Scan(t *testing.T) {
	t.Run("Given failing internal scanner should fail fast", ScannerTest(func(t *testing.T) {
		failingScanner := &FailingScanner{}
		debugScanner := scanner.NewDebugScanner(failingScanner, func(scanner.FileItem) {})

		itemChan, err := debugScanner.Scan(context.TODO())
		Expect(itemChan).To(BeNil())
		Expect(err).To(HaveOccurred())
	}))

	t.Run("Given successful internal scanner should call debug function", ScannerTest(func(t *testing.T) {
		internalScanner := &SuccessfulScanner{
			items: []scanner.FileItem{{}},
		}
		debugCallHappened := false
		debugFunc := func(item scanner.FileItem) {
			debugCallHappened = true
		}
		debugScanner := scanner.NewDebugScanner(internalScanner, debugFunc)

		itemChan, err := debugScanner.Scan(context.TODO())
		Expect(err).NotTo(HaveOccurred())
		Expect(itemChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(1),
		)))
		Expect(debugCallHappened).To(BeTrue())
	}))
}

func TestNewPrintPathNameDebugScanner(t *testing.T) {
	t.Run("Scanner should not panic", ScannerTest(func(t *testing.T) {
		internalScanner := &SuccessfulScanner{
			items: []scanner.FileItem{
				{Err: errors.New("dummy-error")},
			},
		}
		debugScanner := scanner.NewPrintPathNameDebugScanner(internalScanner)
		Expect(func() {
			debugScanner.Scan(context.TODO())
		}).NotTo(Panic())
	}))
}
