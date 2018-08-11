package scanner_test

import (
	"context"
	"testing"

	"regexp"

	"os"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/wojteninho/scanner/pkg/scanner"
)

const (
	nameNegativeFilter = "NegativeFilter"
	namePositiveFilter = "PositiveFilter"
)

var (
	NegativeFilter = MakeNamedFilter(FilterFn(func(_ FileItem) bool { return false }), nameNegativeFilter)
	PositiveFilter = MakeNamedFilter(FilterFn(func(_ FileItem) bool { return true }), namePositiveFilter)
)

type fakeFileInfo struct {
	name string
}

func (f *fakeFileInfo) Name() string {
	return f.name
}

func (f *fakeFileInfo) Size() int64 {
	return 0
}

func (f *fakeFileInfo) Mode() os.FileMode {
	return os.ModePerm
}

func (f *fakeFileInfo) ModTime() time.Time {
	return time.Now()
}

func (f *fakeFileInfo) IsDir() bool {
	return false
}

func (f *fakeFileInfo) Sys() interface{} {
	return nil
}

func (f *fakeFileInfo) PathName() string {
	return ""
}

func TestNamedFilter(t *testing.T) {
	t.Run("Test name", ScannerTest(func(t *testing.T) {
		f := MakeNamedFilter(NegativeFilter, "dummy-name")
		Expect(f).ToNot(BeNil())
		Expect(f).To(BeAssignableToTypeOf(&NamedFilter{}))
		Expect(f.Name()).To(Equal("dummy-name"))
	}))
}

func TestExtensionFilter(t *testing.T) {
	filter := ExtensionFilter(".jpg")

	t.Run("When FileItem.FileInfo is nil", ScannerTest(func(t *testing.T) {
		Expect(filter.Match(FileItem{})).To(BeFalse())
	}))

	t.Run("When match pattern", ScannerTest(func(t *testing.T) {
		Expect(filter.Match(FileItem{FileInfo: &fakeFileInfo{"lorem.jpg"}, Err: nil})).To(BeTrue())
	}))

	t.Run("When does not match pattern", ScannerTest(func(t *testing.T) {
		Expect(filter.Match(FileItem{FileInfo: &fakeFileInfo{"ipsum.png"}, Err: nil})).To(BeFalse())
	}))
}

func TestRegExpFilter(t *testing.T) {
	r := regexp.MustCompile("^lorem")
	filter := RegExpFilter(r)

	t.Run("When FileItem.FileInfo is nil", ScannerTest(func(t *testing.T) {
		Expect(filter.Match(FileItem{})).To(BeFalse())
	}))

	t.Run("When match pattern", ScannerTest(func(t *testing.T) {
		Expect(filter.Match(FileItem{FileInfo: &fakeFileInfo{"lorem.jpg"}, Err: nil})).To(BeTrue())
	}))

	t.Run("When does not match pattern", ScannerTest(func(t *testing.T) {
		Expect(filter.Match(FileItem{FileInfo: &fakeFileInfo{"ipsum.jpg"}, Err: nil})).To(BeFalse())
	}))
}

func TestAndFilter(t *testing.T) {
	t.Run("When no filter functions are passed", ScannerTest(func(t *testing.T) {
		Expect(AndFilter().Match(FileItem{})).To(BeTrue())
	}))

	t.Run("When all filters are negative", ScannerTest(func(t *testing.T) {
		Expect(AndFilter(NegativeFilter, NegativeFilter, NegativeFilter).Match(FileItem{})).To(BeFalse())
	}))

	t.Run("When at least one filter is negative", ScannerTest(func(t *testing.T) {
		Expect(AndFilter(PositiveFilter, PositiveFilter, NegativeFilter).Match(FileItem{})).To(BeFalse())
	}))

	t.Run("When all filters are positive", ScannerTest(func(t *testing.T) {
		Expect(AndFilter(PositiveFilter, PositiveFilter, PositiveFilter).Match(FileItem{})).To(BeTrue())
	}))
}

func TestOrFilter(t *testing.T) {
	t.Run("When no filter functions are passed", ScannerTest(func(t *testing.T) {
		Expect(OrFilter().Match(FileItem{})).To(BeTrue())
	}))

	t.Run("When all filters are negative", ScannerTest(func(t *testing.T) {
		Expect(OrFilter(NegativeFilter, NegativeFilter, NegativeFilter).Match(FileItem{})).To(BeFalse())
	}))

	t.Run("When at least one filter is positive", ScannerTest(func(t *testing.T) {
		Expect(OrFilter(PositiveFilter, NegativeFilter, NegativeFilter).Match(FileItem{})).To(BeTrue())
	}))

	t.Run("When all filters are positive", ScannerTest(func(t *testing.T) {
		Expect(OrFilter(PositiveFilter, PositiveFilter, PositiveFilter).Match(FileItem{})).To(BeTrue())
	}))
}

func TestRegularFilesFilter(t *testing.T) {
	t.Run("When FileItem.FileInfo is nil", ScannerTest(func(t *testing.T) {
		Expect(RegularFilesFilter.Match(FileItem{})).To(BeFalse())
	}))
}

func TestDirectoriesFilter(t *testing.T) {
	t.Run("When is FileItem.FileInfo nil", ScannerTest(func(t *testing.T) {
		Expect(DirectoriesFilter.Match(FileItem{})).To(BeFalse())
	}))
}

func TestFilterRegularFilesScanner(t *testing.T) {
	t.Run("When inner scanner returns an error", ScannerTest(func(t *testing.T) {
		fileChan, err := NewFilterRegularFilesScanner(&FailingScanner{}).Scan(context.TODO())

		Expect(err).To(HaveOccurred())
		Expect(fileChan).To(BeNil())
	}))

	t.Run("When inner scanner return files only", ScannerTest(func(t *testing.T) {
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

	t.Run("When inner scanner return directories only", ScannerTest(func(t *testing.T) {
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

	t.Run("When inner scanner return both files and directories", ScannerTest(func(t *testing.T) {
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
	t.Run("When inner scanner returns an error", ScannerTest(func(t *testing.T) {
		fileChan, err := NewFilterDirectoriesScanner(&FailingScanner{}).Scan(context.TODO())

		Expect(err).To(HaveOccurred())
		Expect(fileChan).To(BeNil())
	}))

	t.Run("When inner scanner return files only", ScannerTest(func(t *testing.T) {
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

	t.Run("When inner scanner return directories only", ScannerTest(func(t *testing.T) {
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

	t.Run("When inner scanner return both files and directories", ScannerTest(func(t *testing.T) {
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
