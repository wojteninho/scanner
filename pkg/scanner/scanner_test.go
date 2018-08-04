package scanner_test

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
	. "github.com/wojteninho/scanner/pkg/scanner"
)

func TestFile(t *testing.T) {
	t.Run("When calling PathName", GomegaTest(func(t *testing.T) {
		_, filename, _, _ := runtime.Caller(0)
		f, err := os.Stat(filename)
		Expect(err).ToNot(HaveOccurred())
		Expect(f).ToNot(BeNil())

		file := NewFile(f, filepath.Dir(filename))
		Expect(file.PathName()).To(Equal(path.Join(filepath.Dir(filename), f.Name())))
	}))
}

func TestMustScanner(t *testing.T) {
	t.Run("When error occurred", GomegaTest(func(t *testing.T) {
		Expect(func() { MustScanner(nil, errors.New("dummy error")) }).To(Panic())
	}))

	t.Run("When error not occurred", GomegaTest(func(t *testing.T) {
		Expect(func() { MustScanner(nil, nil) }).ToNot(Panic())
	}))
}

func TestMustScan(t *testing.T) {
	t.Run("When error occurred", GomegaTest(func(t *testing.T) {
		Expect(func() { MustScan(nil, errors.New("dummy error")) }).To(Panic())
	}))

	t.Run("When error not occurred", GomegaTest(func(t *testing.T) {
		Expect(func() { MustScan(nil, nil) }).ToNot(Panic())
	}))
}

// test helpers
func NewDirectoryPath(name string) string {
	return path.Join(os.TempDir(), fmt.Sprintf("%d-%s", time.Now().UnixNano(), name))
}

type WorkspaceOptionFn func(w *Workspace) error

func WithPermission(permission os.FileMode) WorkspaceOptionFn {
	return func(w *Workspace) error {
		w.permission = permission
		return nil
	}
}

type WorkspaceItem interface {
	Name() string
}

type WorkspaceFile struct {
	name string
}

func (fi WorkspaceFile) Name() string {
	return fi.name
}

func NewWorkspaceFiles(prefix string, n uint) []WorkspaceItem {
	var items []WorkspaceItem
	for i := uint(0); i < n; i++ {
		items = append(items, NewWorkspaceFile(fmt.Sprintf("%s-%d", prefix, i)))
	}
	return items
}

func NewWorkspaceFile(name string) WorkspaceFile {
	return WorkspaceFile{name}
}

type WorkspaceDir struct {
	WorkspaceFile
	files []WorkspaceItem
}

func NewWorkspaceDir(name string, items ...WorkspaceItem) WorkspaceDir {
	return WorkspaceDir{
		WorkspaceFile{name},
		items,
	}
}

func WithItems(items ...WorkspaceItem) WorkspaceOptionFn {
	return func(w *Workspace) error {
		w.items = items
		return nil
	}
}

func WithDebug() WorkspaceOptionFn {
	return func(w *Workspace) error {
		w.debug = true
		return nil
	}
}

type Workspace struct {
	directory  string
	permission os.FileMode
	items      []WorkspaceItem
	debug      bool
}

func (w *Workspace) Purge() {
	if err := os.RemoveAll(w.directory); err != nil {
		panic(err)
	}
}

func MustNewWorkspace(directory string, options ...WorkspaceOptionFn) *Workspace {
	w, err := NewWorkspace(directory, options...)
	if err != nil {
		panic(err)
	}

	return w
}

func NewWorkspace(directory string, options ...WorkspaceOptionFn) (*Workspace, error) {
	w := &Workspace{
		directory:  directory,
		permission: os.ModePerm,
		items:      nil,
		debug:      false,
	}

	for _, option := range options {
		if err := option(w); err != nil {
			return nil, err
		}
	}

	if err := createItems(w.directory, w.permission, w.items...); err != nil {
		return nil, err
	}

	if w.debug {
		debug(w.directory)
	}

	return w, nil
}

func createItems(directory string, permission os.FileMode, items ...WorkspaceItem) error {
	if err := createDir(directory, permission); err != nil {
		return err
	}

	for _, item := range items {
		if dirItem, ok := item.(WorkspaceDir); ok {
			if err := createItems(path.Join(directory, dirItem.Name()), permission, dirItem.files...); err != nil {
				return err
			}

			continue
		}

		if err := createFile(path.Join(directory, item.Name()), permission); err != nil {
			return err
		}
	}

	return nil
}

func createDir(directory string, permission os.FileMode) error {
	return os.MkdirAll(directory, permission)
}

func createFile(name string, permission os.FileMode) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}

	defer f.Close()

	return os.Chmod(name, permission)
}

func debug(directory string) {
	pr, pw := io.Pipe()
	defer pw.Close()

	cmd := exec.Command("tree", directory)
	cmd.Stdout = pw

	go func() {
		defer pr.Close()
		if _, err := io.Copy(os.Stdout, pr); err != nil {
			panic(err)
		}
	}()

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

type FileSlice []FileItem

type FilterFn func(file FileItem) bool

func FilterRegularFilesFn(f FileItem) bool {
	if f.FileInfo == nil {
		return false
	}

	return f.FileInfo.Mode().IsRegular()
}

func FilterDirectoriesFn(f FileItem) bool {
	if f.FileInfo == nil {
		return false
	}

	return f.FileInfo.Mode().IsDir()
}

func (fs FileSlice) Filter(filterFn FilterFn) FileSlice {
	var regularFiles FileSlice

	for _, f := range fs {
		if !filterFn(f) {
			continue
		}

		regularFiles = append(regularFiles, f)
	}

	return regularFiles
}

func (fs FileSlice) FilterRegularFiles() FileSlice {
	return fs.Filter(FilterRegularFilesFn)
}

func (fs FileSlice) FilterDirectories() FileSlice {
	return fs.Filter(FilterDirectoriesFn)
}

func FileChanToSlice(fileChan FileItemChan) FileSlice {
	var files FileSlice

	for file := range fileChan {
		files = append(files, file)
	}

	return files
}

// custom matchers
type HaveFilesMatcher struct {
	matchers.HaveLenMatcher
	FilterFn FilterFn
}

func (m *HaveFilesMatcher) Match(actual interface{}) (success bool, err error) {
	files, ok := actual.(FileSlice)
	if !ok {
		return false, fmt.Errorf("Expected\n%s\n to be a FileSlice, while %s given", format.Object(actual, 1), reflect.ValueOf(actual).Kind().String())
	}

	return HaveLen(m.Count).Match(files.Filter(m.FilterFn))
}

func HaveRegularFiles(count int) types.GomegaMatcher {
	return &HaveFilesMatcher{
		HaveLenMatcher: matchers.HaveLenMatcher{Count: count},
		FilterFn:       FilterRegularFilesFn}
}

func HaveDirectories(count int) types.GomegaMatcher {
	return &HaveFilesMatcher{
		HaveLenMatcher: matchers.HaveLenMatcher{Count: count},
		FilterFn:       FilterDirectoriesFn}
}

// gomega wrapper
func GomegaTest(testFn func(t *testing.T)) func(t *testing.T) {
	return func(t *testing.T) {
		RegisterTestingT(t)
		testFn(t)
	}
}
