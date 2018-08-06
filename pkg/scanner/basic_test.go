package scanner_test

import (
	"context"
	"errors"
	"runtime"
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/wojteninho/scanner/pkg/scanner"
)

func TestNewBasicScannerOptions(t *testing.T) {
	t.Run("When erroring option is passed", GomegaTest(func(t *testing.T) {
		s, err := NewBasicScanner(func(_ *BasicScanner) error {
			return errors.New("dummy error")
		})

		Expect(err).To(HaveOccurred())
		Expect(s).To(BeNil())
	}))
}

func TestBasicScannerWhenCannotScan(t *testing.T) {
	t.Run("When directory does not exist", GomegaTest(func(t *testing.T) {
		scanner, err := NewBasicScanner(WithDir("this/directory/does/not/exist"))

		Expect(err).To(HaveOccurred())
		Expect(scanner).To(BeNil())
	}))

	t.Run("When directory is not a dir", GomegaTest(func(t *testing.T) {
		_, filename, _, _ := runtime.Caller(0)
		scanner, err := NewBasicScanner(WithDir(filename))

		Expect(err).To(HaveOccurred())
		Expect(scanner).To(BeNil())
	}))

	t.Run("When directory is not readable", GomegaTest(func(t *testing.T) {
		dir := NewDirectoryPath("not-readable-directory")
		defer MustNewWorkspace(dir, WithPermission(0000)).Purge()

		fileChan, err := MustScanner(NewBasicScanner(WithDir(dir))).Scan(context.TODO())

		Expect(err).To(HaveOccurred())
		Expect(fileChan).To(BeNil())
	}))
}

func TestBasicScanner(t *testing.T) {
	t.Run("When directory is not set", GomegaTest(func(t *testing.T) {
		fileChan, err := MustScanner(NewBasicScanner()).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, BeEmpty()))
	}))

	t.Run("When directory is empty", GomegaTest(func(t *testing.T) {
		dir := NewDirectoryPath("empty-directory")
		defer MustNewWorkspace(dir).Purge()

		fileChan, err := MustScanner(NewBasicScanner(WithDir(dir))).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, BeEmpty()))
	}))

	t.Run("When directory is not empty, but has flat structure and contains only empty directories", GomegaTest(func(t *testing.T) {
		dir := NewDirectoryPath("flat-directory-with-empty-sub-directories-only")
		defer MustNewWorkspace(dir, WithItems(
			NewWorkspaceDir("level-0-directory-1"),
			NewWorkspaceDir("level-0-directory-2"),
			NewWorkspaceDir("level-0-directory-3"),
		)).Purge()

		fileChan, err := MustScanner(NewBasicScanner(WithDir(dir))).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(3),
			HaveDirectories(3),
		)))
	}))

	t.Run("When directory is not empty, but has flat structure and contains only files", GomegaTest(func(t *testing.T) {
		dir := NewDirectoryPath("flat-directory-with-files-only")
		defer MustNewWorkspace(dir, WithItems(
			NewWorkspaceFile("level-0-file-1.jpg"),
			NewWorkspaceFile("level-0-file-2.jpg"),
			NewWorkspaceFile("level-0-file-3.jpg"),
		)).Purge()

		fileChan, err := MustScanner(NewBasicScanner(WithDir(dir))).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(3),
			HaveRegularFiles(3),
		)))
	}))

	t.Run("When directory is not empty, but has flat structure and contains both empty directories and files", GomegaTest(func(t *testing.T) {
		dir := NewDirectoryPath("flat-directory-with-files-only")
		defer MustNewWorkspace(dir, WithItems(
			NewWorkspaceDir("level-0-directory-1"),
			NewWorkspaceDir("level-0-directory-2"),
			NewWorkspaceDir("level-0-directory-3"),
			NewWorkspaceFile("level-0-file-1.jpg"),
			NewWorkspaceFile("level-0-file-2.jpg"),
			NewWorkspaceFile("level-0-file-3.jpg"),
		)).Purge()

		fileChan, err := MustScanner(NewBasicScanner(WithDir(dir))).Scan(context.TODO())
		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(6),
			HaveDirectories(3),
			HaveRegularFiles(3),
		)))
	}))

	t.Run("When directory is not empty, but has nested structure and contains but not empty directories and files", GomegaTest(func(t *testing.T) {
		dir := NewDirectoryPath("flat-directory-with-not-empty-sub-directories")
		defer MustNewWorkspace(dir, WithItems(
			// empty directory
			NewWorkspaceDir("level-0-directory-1"),

			// directory with file inside
			NewWorkspaceDir("level-0-directory-2",
				NewWorkspaceFile("level-1-file-2.1.jpg"),
			),

			// directory with directory with file inside
			NewWorkspaceDir("level-0-directory-3",
				NewWorkspaceDir("level-1-directory-3.1",
					NewWorkspaceFile("level-2-file-3.1.1.jpg"),
				),
			),

			// regular file
			NewWorkspaceFile("level-0-file-3.jpg"),
		)).Purge()

		fileChan, err := MustScanner(NewBasicScanner(WithDir(dir))).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(4),
			HaveDirectories(3),
			HaveRegularFiles(1),
		)))
	}))
}
