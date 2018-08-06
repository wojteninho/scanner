package scanner_test

import (
	"context"
	"testing"

	"errors"

	. "github.com/onsi/gomega"
	. "github.com/wojteninho/scanner/pkg/scanner"
)

type FailingScanner struct{}

func (s *FailingScanner) Scan(ctx context.Context) (FileItemChan, error) {
	return nil, errors.New("dummy error")
}

type SuccessfulScanner struct {
	items []FileItem
}

func (s *SuccessfulScanner) Scan(ctx context.Context) (FileItemChan, error) {
	fileChan := make(FileItemChan)

	go func() {
		defer close(fileChan)
		for _, item := range s.items {
			fileChan <- item
		}
	}()

	return fileChan, nil
}

func TestMultiScanner(t *testing.T) {
	t.Run("When no scanners are passed", GomegaTest(func(t *testing.T) {
		fileChan, err := NewMultiScanner().Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, BeEmpty()))
	}))

	t.Run("When single failing scanner is passed", GomegaTest(func(t *testing.T) {
		fileChan, err := NewMultiScanner(&FailingScanner{}).Scan(context.TODO())

		Expect(err).To(HaveOccurred())
		Expect(fileChan).To(BeNil())
	}))

	t.Run("When multiple failing scanners are passed", GomegaTest(func(t *testing.T) {
		fileChan, err := NewMultiScanner(&FailingScanner{}, &FailingScanner{}, &FailingScanner{}).Scan(context.TODO())

		Expect(err).To(HaveOccurred())
		Expect(fileChan).To(BeNil())
	}))

	t.Run("When both successful and failing scanners are passed", GomegaTest(func(t *testing.T) {
		fileChan, err := NewMultiScanner(
			&SuccessfulScanner{items: []FileItem{{}, {}, {}}},
			&FailingScanner{},
		).Scan(context.TODO())

		Expect(err).To(HaveOccurred())
		Expect(fileChan).To(BeNil())
	}))

	t.Run("When single successful scanner is passed", GomegaTest(func(t *testing.T) {
		fileChan, err := NewMultiScanner(&SuccessfulScanner{items: []FileItem{{}, {}, {}}}).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, HaveLen(3)))
	}))

	t.Run("When multiple successful scanners are passed", GomegaTest(func(t *testing.T) {
		fileChan, err := NewMultiScanner(
			&SuccessfulScanner{items: []FileItem{{}, {}, {}}},
			&SuccessfulScanner{items: []FileItem{{}, {}, {}}},
			&SuccessfulScanner{items: []FileItem{{}, {}, {}}},
		).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, HaveLen(9)))
	}))

	t.Run("When real Scanners are passed", GomegaTest(func(t *testing.T) {
		firstDir := NewDirectoryPath("first-directory")
		defer MustNewWorkspace(firstDir, WithItems(
			NewWorkspaceDir("level-0-directory-1"),
			NewWorkspaceDir("level-0-directory-2"),
			NewWorkspaceDir("level-0-directory-3"),
		)).Purge()

		secondDir := NewDirectoryPath("second-directory")
		defer MustNewWorkspace(secondDir, WithItems(
			NewWorkspaceFile("level-0-file-1.jpg"),
			NewWorkspaceFile("level-0-file-2.jpg"),
			NewWorkspaceFile("level-0-file-3.jpg"),
		)).Purge()

		fileChan, err := NewMultiScanner(
			MustScanner(NewBasicScanner(WithDir(firstDir))),
			MustScanner(NewBasicScanner(WithDir(secondDir))),
		).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, HaveLen(6)))
	}))
}
