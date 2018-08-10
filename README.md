[![Build Status](https://travis-ci.org/wojteninho/scanner.svg?branch=master)](https://travis-ci.org/wojteninho/scanner)
[![codecov](https://codecov.io/gh/wojteninho/scanner/branch/master/graph/badge.svg)](https://codecov.io/gh/wojteninho/scanner)

# Scanner

The objective of the **Scanner** package is to provide a convenient and versatile tool to iterate through files in the filesystem. It is a set of composable iterators
to execute common tasks like scanning single directory, scanning multiple directories, recursive scanning, filtering etc.

Say you want to iterate through the directory. With scanner package it looks like:
```go
scanner := MustScanner(NewBasicScanner(WithDir("/your/directory/to/scan")))

for item := range MustScan(scanner.Scan(context.TODO())) {
    ...
}
```

# Design

Every scanner is built on top of the same interface:
```go
type Scanner interface {
	Scan(ctx context.Context) (FileItemChan, error)
}
```
It returns `FileItemChan` which is the channel of following structs:
```go
type FileItem struct {
	FileInfo FileInfo
	Err      error
}
```
It means on every iteration you have to check against the error:
```go
fileItemChan := MustScan(dummyScanner.Scan(ctx))
for item := range fileItemChan {
    if item.Err != nil {
        // handle
    } else {
        fmt.Println(item.FileInfo.Pathname())
    }
}
```
`FileInfo` interface in the scanner package is the extension of native `os.FileInfo`
```go
type FileInfo interface {
	os.FileInfo
	PathName() string
}
```
Purpose of the extension is additional method `PathName()` which returns the full path of the filename.
Native `os.FileInfo` doesn't hold information about the directory and in some cases (when you recursively iterate through the directories for instance)
it is a crucial information. Custom interface fills up the missing gap.

# Available scanners

Out of all the available scanners, we can distinguish between 2 concrete types of scanners
* [BasicScanner](https://github.com/wojteninho/scanner#recursivecanner)
* [RescursiveScanner](https://github.com/wojteninho/scanner#recursivecanner)

and set of "wrappers" that are not independent iterators itself, but provide additional functionalities and enhance the behaviour of the concrete scanner they wrap
* [MultiScanner](https://github.com/wojteninho/scanner#multiscanner)
* [FilterScanner](https://github.com/wojteninho/scanner#filterscanner)
* [DebugScanner](https://github.com/wojteninho/scanner#debugscanner)

## BasicScanner

BasicScanner is the simplest possible iterator. It is able to iterate through the single directory and scan the files from this particular directory only.
```go
scanner := MustScanner(NewBasicScanner(WithDir("/directory/you/want/to/scan")))

for item := range MustScan(scanner.Scan(context.TODO())) {
    ...
}
```

## RecursiveScanner

RecursiveScanner is the more versatile and robust scanner. As the name says, its main feature is an ability to scan a directory recursively.
It makes use of the concurrent nature of the Golang itself and spawns up to the certain and fixed limit of workers concurrently.
By default, it set `runtime.NumCPU()` as the limit, but you can modify it to your needs accordingly by passing additional option to the constructor function:
```go
NewRecursiveScanner(WithWorkers(2))
```
Worth to say, implementation behind the scenes is a dynamic worker pool. Say we have a default limit of `runtime.NumCPU()` workers which are for instance 4 on the target machine.
It means if we have to scan 2 directories we will spawn only 2 workers if we have to scan 4 directories we will spawn only 4 workers, but anything above the limit
will not spawn new workers, but append the directories to the internal queue and spawn new processing when, and only when, the pool of the workers allows to spawn new worker.

Another attribute of RecursiveScanner is an ability to scan multiple directories. You can set them by passing option to the constructor function:
```go
recursiveScanner := MustScanner(NewRecursiveScanner(WithDirectories(
    "/your/first/firectory/to/scan",
    "/your/second/directory/to/scan",
    "/one/more/directory/to/scan",
)))

for item := range MustScan(recursiveScanner.Scan(context.TODO())) {
    ...
}
```

## MultiScanner

MultiScanner is one amongst the "wrappers" family. It is not self-sufficient scanner itself, but needs to wrap concrete scanners. Objective of the MultiScanner is to merge
multiple scanners into one. More than enough is to see the example.
```go
firstScanner := MustScanner(NewBasicScanner(WithDirectory("/your/first/directory/to/scan")))
secondScanner := MustScanner(NewBasicScanner(WithDirectory("/your/second/directory/to/scan")))
thirdScanner := MustScanner(NewBasicScanner(WithDirectory("/your/third/directory/to/scan")))

multiScanner := NewMultiScanner(firstScanner, secondScanner, thirdScanner)
for item := range MustScan(multiScanner.Scan(context.TODO())) {
    ...
}
```

## FilterScanner

FilterScanner enhance the wrapped scanner with filtering feature. Constructor function is as follow:
```go
func NewFilterScanner(scanner Scanner, filterFn FilterFn) Scanner
```
while definition of `FilterFn` is
```go
type FilterFn func(file FileItem) bool
```
In the `Scanner` package already exist two implementations:
* FilterRegularFilesScanner
* FilterDirectoriesScanner

As always one example worth more than 1000 words:
```go
scanner := MustScanner(NewBasicScanner(WithDirectory("/first/directory")))
regularFilesScanner := NewFilterRegularFilesScanner(scanner)

for item := range MustScan(regularFilesScanner.Scan(context.TODO())) {
    ...
}
```
However, by providing a custom function that fulfills the `FilterFn` definition you can build your own Scanners. Say we want to filter out files with names longer than 7 characters
(I know it is a contrived use case, but let's take it into consideration only for the sake of example)
```go
scanner := MustScanner(NewBasicScanner(WithDir("/your/directory/to/scan")))
filterScanner := NewFilterScanner(scanner, func(file FileItem) bool {
    return len(file.FileInfo.Name()) > 7
})

for item := range MustScan(filterScanner.Scan(context.TODO())) {
    ...
}
```

## DebugScanner

Say you want to output to os.Stdout the full pathname of each file. This feature comes with custom implementation of `DebugScanner`:
```go
scanner := MustScanner(NewBasicScanner(WithDir("/your/directory/to/scan")))
debugScanner := NewPrintPathNameDebugScanner(scanner)

for item := range MustScan(debugScanner.Scan(context.TODO())) {
    ...
}
```

If you need something more sophisticated you can write you own variation of DebugScanner. Simply implement your own `DebugFn` with following definition:
```go
type DebugFn func(item FileItem)
```
and create your own `DebugScanner`:
```go
scanner := MustScanner(NewBasicScanner(WithDir("/your/directory/to/scan")))
debugScanner := NewDebugScanner(scanner, func (item FileItem) {
    ...
})
```
