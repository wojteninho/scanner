package scanner_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/wojteninho/scanner/pkg/scanner"
)

func TestFilterRegularFilesFn(t *testing.T) {
	t.Run("When is FileItem.FileInfo nil", GomegaTest(func(t *testing.T) {
		Expect(FilterRegularFilesFn(FileItem{})).To(BeFalse())
	}))
}

func TestFilterDirectoriesFn(t *testing.T) {
	t.Run("When is FileItem.FileInfo nil", GomegaTest(func(t *testing.T) {
		Expect(FilterDirectoriesFn(FileItem{})).To(BeFalse())
	}))
}

func TestFilterRegularFilesScanner(t *testing.T) {
	t.Run("When inner scanner returns an error", GomegaTest(func(t *testing.T) {
		fileChan, err := NewFilterRegularFilesScanner(&FailingScanner{}).Scan(context.TODO())

		Expect(err).To(HaveOccurred())
		Expect(fileChan).To(BeNil())
	}))

	t.Run("When inner scanner return files only", GomegaTest(func(t *testing.T) {
		dir := NewDirectoryPath("directory-with-files-only")
		defer MustNewWorkspace(dir, WithItems(
			NewWorkspaceFile("level-0-file-1.jpg"),
			NewWorkspaceFile("level-0-file-2.jpg"),
			NewWorkspaceFile("level-0-file-3.jpg"),
		)).Purge()

		innerScanner := MustScanner(NewBasicScanner(WithDir(dir)))
		fileChan, err := NewFilterRegularFilesScanner(innerScanner).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(3),
			HaveRegularFiles(3),
		)))
	}))

	t.Run("When inner scanner return directories only", GomegaTest(func(t *testing.T) {
		dir := NewDirectoryPath("directory-with-directories-only")
		defer MustNewWorkspace(dir, WithItems(
			NewWorkspaceDir("level-0-directory-1"),
			NewWorkspaceDir("level-0-directory-2"),
			NewWorkspaceDir("level-0-directory-3"),
		)).Purge()

		innerScanner := MustScanner(NewBasicScanner(WithDir(dir)))
		fileChan, err := NewFilterRegularFilesScanner(innerScanner).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(0),
			HaveDirectories(0),
		)))
	}))

	t.Run("When inner scanner return both files and directories", GomegaTest(func(t *testing.T) {
		dir := NewDirectoryPath("directory-with-directories-only")
		defer MustNewWorkspace(dir, WithItems(
			NewWorkspaceFile("level-0-file-1.jpg"),
			NewWorkspaceFile("level-0-file-2.jpg"),
			NewWorkspaceFile("level-0-file-3.jpg"),
			NewWorkspaceDir("level-0-directory-1"),
			NewWorkspaceDir("level-0-directory-2"),
			NewWorkspaceDir("level-0-directory-3"),
		)).Purge()

		innerScanner := MustScanner(NewBasicScanner(WithDir(dir)))
		fileChan, err := NewFilterRegularFilesScanner(innerScanner).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(3),
			HaveRegularFiles(3),
		)))
	}))
}

func TestFilterDirectoriesScanner(t *testing.T) {
	t.Run("When inner scanner returns an error", GomegaTest(func(t *testing.T) {
		fileChan, err := NewFilterDirectoriesScanner(&FailingScanner{}).Scan(context.TODO())

		Expect(err).To(HaveOccurred())
		Expect(fileChan).To(BeNil())
	}))

	t.Run("When inner scanner return files only", GomegaTest(func(t *testing.T) {
		dir := NewDirectoryPath("directory-with-files-only")
		defer MustNewWorkspace(dir, WithItems(
			NewWorkspaceFile("level-0-file-1.jpg"),
			NewWorkspaceFile("level-0-file-2.jpg"),
			NewWorkspaceFile("level-0-file-3.jpg"),
		)).Purge()

		innerScanner := MustScanner(NewBasicScanner(WithDir(dir)))
		fileChan, err := NewFilterDirectoriesScanner(innerScanner).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(0),
			HaveDirectories(0),
		)))
	}))

	t.Run("When inner scanner return directories only", GomegaTest(func(t *testing.T) {
		dir := NewDirectoryPath("directory-with-directories-only")
		defer MustNewWorkspace(dir, WithItems(
			NewWorkspaceDir("level-0-directory-1"),
			NewWorkspaceDir("level-0-directory-2"),
			NewWorkspaceDir("level-0-directory-3"),
		)).Purge()

		innerScanner := MustScanner(NewBasicScanner(WithDir(dir)))
		fileChan, err := NewFilterDirectoriesScanner(innerScanner).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(3),
			HaveDirectories(3),
		)))
	}))

	t.Run("When inner scanner return both files and directories", GomegaTest(func(t *testing.T) {
		dir := NewDirectoryPath("directory-with-directories-only")
		defer MustNewWorkspace(dir, WithItems(
			NewWorkspaceFile("level-0-file-1.jpg"),
			NewWorkspaceFile("level-0-file-2.jpg"),
			NewWorkspaceFile("level-0-file-3.jpg"),
			NewWorkspaceDir("level-0-directory-1"),
			NewWorkspaceDir("level-0-directory-2"),
			NewWorkspaceDir("level-0-directory-3"),
		)).Purge()

		innerScanner := MustScanner(NewBasicScanner(WithDir(dir)))
		fileChan, err := NewFilterDirectoriesScanner(innerScanner).Scan(context.TODO())

		Expect(err).ToNot(HaveOccurred())
		Expect(fileChan).ToNot(BeNil())
		Expect(fileChan).To(WithTransform(FileChanToSlice, And(
			HaveLen(3),
			HaveDirectories(3),
		)))
	}))
}
