package scanner_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/wojteninho/scanner/pkg/scanner"
)

const scanLimitCount = scanner.Limit(5)

func TestLimitScanner_Scan(t *testing.T) {
	t.Run("Given failing internal scanner should fail fast", ScannerTest(func(t *testing.T) {
		failingScanner := &FailingScanner{}
		debugScanner := scanner.NewLimitScanner(failingScanner, scanLimitCount)

		itemChan, err := debugScanner.Scan(context.TODO())
		Expect(itemChan).To(BeNil())
		Expect(err).To(HaveOccurred())
	}))

	t.Run("Given successful scanner returns limited results", ScannerTest(func(t *testing.T) {
		internalScanner := &SuccessfulScanner{
			items: []scanner.FileItem{{}, {}, {}, {}, {}, {}, {}},
		}

		limitScanner := scanner.NewLimitScanner(internalScanner, scanLimitCount)

		itemChan, err := limitScanner.Scan(context.TODO())

		Expect(err).NotTo(HaveOccurred())
		Expect(itemChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(5),
		)))
	}))
}
