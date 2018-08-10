package scanner_test

import (
	"context"
	"errors"
	"runtime"
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/wojteninho/scanner/pkg/scanner"
)

func TestRecursiveScannerOptions(t *testing.T) {
	t.Run("When erroring option is passed", ScannerTest(func(t *testing.T) {
		s, err := NewRecursiveScanner(func(_ *RecursiveScanner) error {
			return errors.New("dummy error")
		})

		Expect(err).To(HaveOccurred())
		Expect(s).To(BeNil())
	}))

	t.Run("When no directories passed", ScannerTest(func(t *testing.T) {
		err := WithDirectories()(&RecursiveScanner{})
		Expect(err).ToNot(HaveOccurred())
	}))
}

func TestRecursiveScanner(t *testing.T) {
	t.Run("When directory does not exist", ScannerTest(func(t *testing.T) {
		scanner, err := NewRecursiveScanner(WithDirectories("this/directory/does/not/exist"))

		Expect(err).To(HaveOccurred())
		Expect(scanner).To(BeNil())
	}))

	t.Run("When directory is not a dir", ScannerTest(func(t *testing.T) {
		_, filename, _, _ := runtime.Caller(0)
		scanner, err := NewRecursiveScanner(WithDirectories(filename))

		Expect(err).To(HaveOccurred())
		Expect(scanner).To(BeNil())
	}))

	t.Run("When directory is not readable", ScannerTest(func(t *testing.T) {
		t.Run("Directory from options", ScannerTest(func(t *testing.T) {
			dir := NewDirectoryPath("not-readable-directory")
			defer MustNewWorkspace(dir, WithPermission(0000)).Purge()

			fileChan, err := MustScanner(NewRecursiveScanner(WithDirectories(dir))).Scan(context.TODO())

			Expect(err).ToNot(HaveOccurred())
			Expect(fileChan).ToNot(BeNil())
			Expect(fileChan).To(WithTransform(FileChanToSlice, And(
				HaveLen(1),
				HaveErrors(1),
			)))
		}))

		t.Run("Nested directory", ScannerTest(func(t *testing.T) {
			t.Skip("TODO")
		}))
	}))

	t.Run("When directory is not set", ScannerTest(func(t *testing.T) {
		fileChan, err := MustScanner(NewRecursiveScanner()).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, BeEmpty()))
	}))

	t.Run("When directory is empty", ScannerTest(func(t *testing.T) {
		dir := NewDirectoryPath("empty-directory")
		defer MustNewWorkspace(dir).Purge()

		fileChan, err := MustScanner(NewRecursiveScanner(WithDirectories(dir))).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, BeEmpty()))
	}))

	t.Run("When directory is duplicated, but empty", ScannerTest(func(t *testing.T) {
		dir := NewDirectoryPath("empty-directory")
		defer MustNewWorkspace(dir).Purge()

		fileChan, err := MustScanner(NewRecursiveScanner(WithDirectories(dir, dir, dir))).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, BeEmpty()))
	}))

	t.Run("When directory is not empty, but is flat and contains files only", ScannerTest(func(t *testing.T) {
		dir := NewDirectoryPath("flat-directory-with-files-only")
		defer MustNewWorkspace(dir, WithItems(
			NewWorkspaceFile("level-0-file-1.jpg"),
			NewWorkspaceFile("level-0-file-2.jpg"),
			NewWorkspaceFile("level-0-file-3.jpg"),
		)).Purge()

		fileChan, err := MustScanner(NewRecursiveScanner(WithDirectories(dir))).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(3),
			HaveRegularFiles(3),
		)))
	}))

	t.Run("When directory is not empty, but is flat and contains empty directories only", ScannerTest(func(t *testing.T) {
		dir := NewDirectoryPath("flat-directory-with-empty-directories-only")
		defer MustNewWorkspace(dir, WithItems(
			NewWorkspaceDir("level-0-directory-1"),
			NewWorkspaceDir("level-0-directory-2"),
			NewWorkspaceDir("level-0-directory-3"),
		)).Purge()

		fileChan, err := MustScanner(NewRecursiveScanner(WithDirectories(dir))).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(3),
			HaveDirectories(3),
		)))
	}))

	t.Run("When directory is not empty, but nested with 1 level depth and contains not empty directories", ScannerTest(func(t *testing.T) {
		dir := NewDirectoryPath("nested-directory-with-1-level-depth-contains-not-empty-directories")
		defer MustNewWorkspace(dir, WithItems(
			NewWorkspaceDir("level-0-directory-1",
				NewWorkspaceFile("level-1-file-1.1.jpg"),
				NewWorkspaceFile("level-1-file-1.2.jpg"),
				NewWorkspaceFile("level-1-file-1.3.jpg"),
			),
			NewWorkspaceDir("level-0-directory-2",
				NewWorkspaceFile("level-1-file-2.1.jpg"),
				NewWorkspaceFile("level-1-file-2.2.jpg"),
				NewWorkspaceFile("level-1-file-2.3.jpg"),
			),
			NewWorkspaceDir("level-0-directory-3",
				NewWorkspaceFile("level-1-file-3.1.jpg"),
				NewWorkspaceFile("level-1-file-3.2.jpg"),
				NewWorkspaceFile("level-1-file-3.3.jpg"),
			),
		)).Purge()

		fileChan, err := MustScanner(NewRecursiveScanner(WithDirectories(dir))).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(12),
			HaveDirectories(3),
			HaveRegularFiles(9),
		)))
	}))

	t.Run("When directory is not empty, but nested with 2 level depth and contains not empty directories", ScannerTest(func(t *testing.T) {
		dir := NewDirectoryPath("nested-directory-with-2-level-depth-contains-not-empty-directories")
		defer MustNewWorkspace(dir, WithItems(
			NewWorkspaceDir("level-0-directory-1",
				NewWorkspaceFile("level-1-file-1.1.jpg"),
				NewWorkspaceFile("level-1-file-1.2.jpg"),
				NewWorkspaceFile("level-1-file-1.3.jpg"),
				NewWorkspaceDir("level-1-directory-1.4",
					NewWorkspaceFile("level-2-file-1.4.1.jpg"),
					NewWorkspaceFile("level-2-file-1.4.2.jpg"),
					NewWorkspaceFile("level-2-file-1.4.3.jpg"),
				),
			),
			NewWorkspaceDir("level-0-directory-2",
				NewWorkspaceFile("level-1-file-2.1.jpg"),
				NewWorkspaceFile("level-1-file-2.2.jpg"),
				NewWorkspaceFile("level-1-file-2.3.jpg"),
				NewWorkspaceDir("level-1-directory-2.4",
					NewWorkspaceFile("level-2-file-2.4.1.jpg"),
					NewWorkspaceFile("level-2-file-2.4.2.jpg"),
					NewWorkspaceFile("level-2-file-2.4.3.jpg"),
				),
			),
			NewWorkspaceDir("level-0-directory-3",
				NewWorkspaceFile("level-1-file-3.1.jpg"),
				NewWorkspaceFile("level-1-file-3.2.jpg"),
				NewWorkspaceFile("level-1-file-3.3.jpg"),
				NewWorkspaceDir("level-1-directory-3.4",
					NewWorkspaceFile("level-2-file-3.4.1.jpg"),
					NewWorkspaceFile("level-2-file-3.4.2.jpg"),
					NewWorkspaceFile("level-2-file-3.4.3.jpg"),
				),
			),
		)).Purge()

		fileChan, err := MustScanner(NewRecursiveScanner(WithDirectories(dir))).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(24),
			HaveDirectories(6),
			HaveRegularFiles(18),
		)))
	}))

	t.Run("When directory is not empty, but nested with more levels than workers depth and contains not empty directories", ScannerTest(func(t *testing.T) {
		dir := NewDirectoryPath("nested-directory-with-2-level-depth-contains-not-empty-directories")
		defer MustNewWorkspace(dir, WithItems(
			NewWorkspaceDir("level-0-directory-1",
				NewWorkspaceFile("level-1-file-1.1.jpg"),
				NewWorkspaceFile("level-1-file-1.2.jpg"),
				NewWorkspaceFile("level-1-file-1.3.jpg"),
				NewWorkspaceDir("level-1-directory-1.4",
					NewWorkspaceFile("level-2-file-1.4.1.jpg"),
					NewWorkspaceFile("level-2-file-1.4.2.jpg"),
					NewWorkspaceFile("level-2-file-1.4.3.jpg"),
					NewWorkspaceDir("level-2-directory-1.4.4",
						NewWorkspaceFile("level-3-file-1.4.4.1.jpg"),
						NewWorkspaceFile("level-3-file-1.4.4.2.jpg"),
						NewWorkspaceFile("level-3-file-1.4.4.3.jpg"),
						NewWorkspaceDir("level-3-directory-1.4.4.4",
							NewWorkspaceFile("level-4-file-1.4.4.4.1.jpg"),
							NewWorkspaceFile("level-4-file-1.4.4.4.2.jpg"),
							NewWorkspaceFile("level-4-file-1.4.4.4.3.jpg"),
							NewWorkspaceDir("level-4-directory-1.4.4.4.4"),
						),
					),
				),
			),
			NewWorkspaceDir("level-0-directory-2",
				NewWorkspaceFile("level-1-file-2.1.jpg"),
				NewWorkspaceFile("level-1-file-2.2.jpg"),
				NewWorkspaceFile("level-1-file-2.3.jpg"),
				NewWorkspaceDir("level-1-directory-2.4",
					NewWorkspaceFile("level-2-file-2.4.1.jpg"),
					NewWorkspaceFile("level-2-file-2.4.2.jpg"),
					NewWorkspaceFile("level-2-file-2.4.3.jpg"),
					NewWorkspaceDir("level-2-directory-2.4.4",
						NewWorkspaceFile("level-3-file-2.4.4.4.1.jpg"),
						NewWorkspaceFile("level-3-file-2.4.4.4.2.jpg"),
						NewWorkspaceFile("level-3-file-2.4.4.4.3.jpg"),
						NewWorkspaceDir("level-3-directory-2.4.4.4",
							NewWorkspaceFile("level-4-file-2.4.4.4.1.jpg"),
							NewWorkspaceFile("level-4-file-2.4.4.4.2.jpg"),
							NewWorkspaceFile("level-4-file-2.4.4.4.3.jpg"),
							NewWorkspaceDir("level-4-directory-2.4.4.4.4"),
						),
					),
				),
			),
			NewWorkspaceDir("level-0-directory-3",
				NewWorkspaceFile("level-1-file-3.1.jpg"),
				NewWorkspaceFile("level-1-file-3.2.jpg"),
				NewWorkspaceFile("level-1-file-3.3.jpg"),
				NewWorkspaceDir("level-1-directory-3.4",
					NewWorkspaceFile("level-2-file-3.4.1.jpg"),
					NewWorkspaceFile("level-2-file-3.4.2.jpg"),
					NewWorkspaceFile("level-2-file-3.4.3.jpg"),
					NewWorkspaceDir("level-2-directory-3.4.4",
						NewWorkspaceFile("level-3-file-3.4.4.4.1.jpg"),
						NewWorkspaceFile("level-3-file-3.4.4.4.2.jpg"),
						NewWorkspaceFile("level-3-file-3.4.4.4.3.jpg"),
						NewWorkspaceDir("level-3-directory-3.4.4.4",
							NewWorkspaceFile("level-4-file-3.4.4.4.1.jpg"),
							NewWorkspaceFile("level-4-file-3.4.4.4.2.jpg"),
							NewWorkspaceFile("level-4-file-3.4.4.4.3.jpg"),
							NewWorkspaceDir("level-4-directory-3.4.4.4.4"),
						),
					),
				),
			),
		)).Purge()

		fileChan, err := MustScanner(NewRecursiveScanner(WithDirectories(dir), WithWorkers(2))).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(51),
			HaveDirectories(15),
			HaveRegularFiles(36),
		)))
	}))

	t.Run("When context is terminated", ScannerTest(func(t *testing.T) {
		t.Skip("TODO")
	}))
}
