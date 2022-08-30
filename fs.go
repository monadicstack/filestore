package filestore

import (
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

// ReaderFile encapsulates a file within a file system that you can read from.
type ReaderFile interface {
	io.ReadCloser
	io.ReaderAt
	io.Seeker
}

// WriterFile encapsulates a file within a file system that you can write to.
type WriterFile interface {
	io.WriteCloser
	io.WriterAt
	io.Seeker
}

// FileInfo contains 'stat' info about a file or directory.
type FileInfo fs.FileInfo

// FS represents a file system that you can interact with its directories and files.
type FS interface {
	// WorkingDirectory returns the current FS context's path/directory.
	WorkingDirectory() string
	// Stat fetches metadata about the file w/o actually opening it for reading/writing.
	Stat(path string) (FileInfo, error)
	// Read opens the given file for reading.
	Read(path string) (ReaderFile, error)
	// Write opens the given file for writing
	Write(path string) (WriterFile, error)
	// Exists returns true when the file/directory already exits in the file system.
	Exists(path string) bool
	// List performs a UNIX style "ls" operation, giving you the names of each file
	// in the given directory. The filters offer a way to limit which files/dirs are included
	// in the final slice.
	//
	// Example:
	//
	//    filesAndDirs, err := myFS.List("./conf")
	//    jsonFiles, err := myFS.List("./conf", filestore.WithExt("json"))
	List(path string, filters ...FileFilter) ([]FileInfo, error)
	// ChangeDirectory creates a new FS in the given subdirectory. All operations on this new
	// instance will be rooted in the given directory.
	//
	// It should NOT matter if the directory exists or not. You still always get a valid FS
	// pointing to that location and only get an error when you attempt to perform some other operation.
	//
	// Example:
	//
	//    usrFS := Disk("/usr")
	//    usrLocalBinFS := usrFS.ChangeDirectory("local/bin")
	ChangeDirectory(path string) FS
	// Remove deletes the given file/directory within the file system. If the given path
	// is a directory, it should recursively delete it and its children. Additionally,
	// if you attempt to remove a file/directory that does not exist, this should behave
	// quietly as a nop, returning a nil error but not changing the store's state.
	//
	// Example:
	//
	//    myDocumentsFS := Disk("/Users/rob/Documents")
	//    err = myDocumentsFS.Remove("foo.txt")
	//    if err != nil {
	//        // could not delete file "foo.txt"
	//    }
	//    err = myDocumentsFS.Remove("Pictures")
	//    if err != nil {
	//        // could not delete directory "Pictures/"
	//    }
	Remove(path string) error
	// Move takes an existing file at the fromPath location and moves it to another
	// spot in this file system; the toPath location.
	Move(fromPath string, toPath string) error
}

// FileFilter provides a way to exclude files/directories from a list/search.
type FileFilter func(info FileInfo) bool

// WithExt creates a file filter that only accepts files that have a specific extension.
func WithExt(extension string) FileFilter {
	// Not specifying any particular extension means you want to allow everything.
	if extension == "" || extension == "." {
		return WithEverything()
	}

	// Make comparison case-insensitive and allow you to pass an extension with
	// or without the leading "."; basically we'll prepend the "." whether you
	// supplied it or not.
	extension = strings.ToLower(extension)
	extension = strings.TrimPrefix(extension, ".")
	extension = "." + extension

	return func(f FileInfo) bool {
		return strings.HasSuffix(strings.ToLower(f.Name()), extension)
	}
}

// WithExts creates a file filter that only accepts files that have one of the given extensions.
func WithExts(extensions ...string) FileFilter {
	var filters []FileFilter
	for _, extension := range extensions {
		filters = append(filters, WithExt(extension))
	}
	return func(f FileInfo) bool {
		for _, filter := range filters {
			if filter(f) {
				return true
			}
		}
		return false
	}
}

// WithPattern only allows files to pass through that match the given glob pattern.
func WithPattern(pattern string) FileFilter {
	if pattern == "" {
		return WithEverything()
	}

	return func(f FileInfo) bool {
		matched, err := filepath.Match(pattern, f.Name())
		return matched && err == nil
	}
}

// WithEverything is a dummy non-nil file filter you can use to act as though there are no filters.
// Basically it behaves such that all files match.
func WithEverything() FileFilter {
	return func(f FileInfo) bool {
		return true
	}
}
