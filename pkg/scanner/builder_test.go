package scanner_test

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/format"
	. "github.com/onsi/gomega/types"
	. "github.com/wojteninho/scanner/pkg/scanner"
)

func TestBuilder(t *testing.T) {
	t.Run("Test MustBuild", ScannerTest(func(t *testing.T) {
		t.Run("When unable to build", ScannerTest(func(t *testing.T) {
			Expect(func() {
				NewBuilder().In("/not/existing/directory").MustBuild()
			}).To(Panic())
		}))

		t.Run("When able to build", ScannerTest(func(t *testing.T) {
			Expect(func() {
				NewBuilder().In("/tmp").MustBuild()
			}).ToNot(Panic())
		}))
	}))

	t.Run("When no specification is set", ScannerTest(func(t *testing.T) {
		scanner, err := NewBuilder().Build()

		Expect(err).ToNot(HaveOccurred())
		Expect(scanner).ToNot(BeNil())
		Expect(scanner).To(BeScanner(
			MustScanner(NewBasicScanner()),
		))
	}))

	t.Run("Files & Directories", ScannerTest(func(t *testing.T) {
		t.Run("When files only mode is specified", ScannerTest(func(t *testing.T) {
			scanner, err := NewBuilder().Files().Build()

			Expect(err).ToNot(HaveOccurred())
			Expect(scanner).ToNot(BeNil())
			Expect(scanner).To(BeScanner(
				NewFilterRegularFilesScanner(MustScanner(NewBasicScanner())),
			))
		}))

		t.Run("When directories only mode is specified", ScannerTest(func(t *testing.T) {
			scanner, err := NewBuilder().Directories().Build()
			Expect(err).ToNot(HaveOccurred())
			Expect(scanner).ToNot(BeNil())
			Expect(scanner).To(BeScanner(
				NewFilterDirectoriesScanner(MustScanner(NewBasicScanner())),
			))
		}))
	}))

	t.Run("Flat & Recursive", ScannerTest(func(t *testing.T) {
		t.Run("When flat mode is specified and no directory", ScannerTest(func(t *testing.T) {
			scanner, err := NewBuilder().Flat().Build()

			Expect(err).ToNot(HaveOccurred())
			Expect(scanner).ToNot(BeNil())
			Expect(scanner).To(BeScanner(
				MustScanner(NewBasicScanner()),
			))
		}))

		t.Run("When flat mode is specified and single invalid directory", ScannerTest(func(t *testing.T) {
			scanner, err := NewBuilder().Flat().In("/this/directory/does/not/exist").Build()

			Expect(err).To(HaveOccurred())
			Expect(scanner).To(BeNil())
		}))

		t.Run("When flat mode is specified and many invalid directories", ScannerTest(func(t *testing.T) {
			scanner, err := NewBuilder().Flat().In("/this/directory/does/not/exist", "/this/directory/does/not/exist/either").Build()

			Expect(err).To(HaveOccurred())
			Expect(scanner).To(BeNil())
		}))

		t.Run("When flat mode is specified and many directories", ScannerTest(func(t *testing.T) {
			scanner, err := NewBuilder().Flat().In("/tmp", "/var").Build()

			Expect(err).ToNot(HaveOccurred())
			Expect(scanner).ToNot(BeNil())
			Expect(scanner).To(BeScanner(
				NewMultiScanner(
					MustScanner(NewBasicScanner(WithDir("/tmp"))),
					MustScanner(NewBasicScanner(WithDir("/var"))),
				),
			))
		}))

		t.Run("When recursive mode is specified and no directory", ScannerTest(func(t *testing.T) {
			scanner, err := NewBuilder().Recursive().Build()

			Expect(err).ToNot(HaveOccurred())
			Expect(scanner).ToNot(BeNil())
			Expect(scanner).To(BeScanner(
				MustScanner(NewRecursiveScanner()),
			))
		}))

		t.Run("When recursive mode is specified and single invalid directory", ScannerTest(func(t *testing.T) {
			scanner, err := NewBuilder().Recursive().In("/this/directory/does/not/exist").Build()

			Expect(err).To(HaveOccurred())
			Expect(scanner).To(BeNil())
		}))

		t.Run("When recursive mode is specified and single directory", ScannerTest(func(t *testing.T) {
			scanner, err := NewBuilder().Recursive().In("/tmp").Build()

			Expect(err).ToNot(HaveOccurred())
			Expect(scanner).ToNot(BeNil())
			Expect(scanner).To(BeScanner(
				MustScanner(NewRecursiveScanner(WithDirectories("/tmp"))),
			))
		}))

		t.Run("When recursive mode is specified and many invalid directories", ScannerTest(func(t *testing.T) {
			scanner, err := NewBuilder().Recursive().In("/this/directory/does/not/exist", "/this/directory/does/not/exist/either").Build()

			Expect(err).To(HaveOccurred())
			Expect(scanner).To(BeNil())
		}))

		t.Run("When recursive mode is specified and many directories", ScannerTest(func(t *testing.T) {
			scanner, err := NewBuilder().Recursive().In("/tmp", "/var").Build()

			Expect(err).ToNot(HaveOccurred())
			Expect(scanner).ToNot(BeNil())
			Expect(scanner).To(BeScanner(
				MustScanner(NewRecursiveScanner(WithDirectories("/tmp", "/var"))),
			))
		}))
	}))

	t.Run("Filter", ScannerTest(func(t *testing.T) {
		t.Run("Single Filter", ScannerTest(func(t *testing.T) {
			scanner, err := NewBuilder().Match(PositiveFilter).Build()

			Expect(err).ToNot(HaveOccurred())
			Expect(scanner).ToNot(BeNil())
			Expect(scanner).To(BeScanner(
				NewFilterScanner(MustScanner(NewBasicScanner()), PositiveFilter),
			))
		}))

		t.Run("Files mode & single Filter", ScannerTest(func(t *testing.T) {
			scanner, err := NewBuilder().Files().Match(PositiveFilter).Build()

			Expect(err).ToNot(HaveOccurred())
			Expect(scanner).ToNot(BeNil())
			Expect(scanner).To(BeScanner(
				NewFilterScanner(MustScanner(NewBasicScanner()), AndFilter(RegularFilesFilter, PositiveFilter)),
			))
		}))

		t.Run("Directories mode & single Filter", ScannerTest(func(t *testing.T) {
			scanner, err := NewBuilder().Directories().Match(PositiveFilter).Build()

			Expect(err).ToNot(HaveOccurred())
			Expect(scanner).ToNot(BeNil())
			Expect(scanner).To(BeScanner(
				NewFilterScanner(MustScanner(NewBasicScanner()), AndFilter(DirectoriesFilter, PositiveFilter)),
			))
		}))
	}))
}

type beScannerMatcher struct {
	Expected interface{}
	Diff     []string
}

func (matcher *beScannerMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil && matcher.Expected == nil {
		return false, fmt.Errorf("Both actual and expected must not be nil.")
	}

	deep.CompareUnexportedFields = true

	if diff := deep.Equal(actual, matcher.Expected); diff != nil {
		matcher.Diff = diff
		return false, nil
	}

	return true, nil
}

func (matcher *beScannerMatcher) FailureMessage(actual interface{}) (message string) {
	return matcher.message(actual, "to be same scanner as", matcher.Expected, matcher.Diff)
}

func (matcher *beScannerMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return matcher.message(actual, "not to be same scanner as", matcher.Expected, matcher.Diff)
}

func (matcher *beScannerMatcher) message(actual interface{}, message string, expected interface{}, diff []string) string {
	return fmt.Sprintf(
		"Expected\n%s\n%s\n%s\n%s\n%s",
		Object(actual, 1),
		message,
		Object(expected, 1),
		"but it is not due to the diff:", Object(diff, 1),
	)
}

func BeScanner(expected interface{}) GomegaMatcher {
	return &beScannerMatcher{
		Expected: expected,
	}
}
