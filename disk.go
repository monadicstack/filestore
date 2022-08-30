package filestore

import (
	"fmt"
	"os"
	"path"
)

// Disk creates a new file store that reads and writes files to/from
// the local file system. All operations will be rooted in the given directory.
//
// Example:
//
//	files := Disk("./data")
//
//	// Open ./data/input.txt for reading
//	input, err := files.Read("input.txt")
//	if err != nil {
//	    // handle your error nicely
//	}
//	defer input.Close()
//
//	inputBytes, err := io.ReadAll(input)
func Disk(basePath string) *DiskFS {
	return &DiskFS{basePath: basePath}
}

// DiskFS is a file store whose operations interact w/ the local file system.
type DiskFS struct {
	basePath string
}

// diskFile provides implementations for all reading, writing, and 'stat' information
// about a file read from a DiskFS.
type diskFile struct {
	file *os.File
}

// Seek moves to the given offset w/o reading/writing any data.
func (d diskFile) Seek(offset int64, whence int) (int64, error) {
	if d.file == nil {
		return 0, fmt.Errorf("disk fs: seek: file has not been opened")
	}
	return d.file.Seek(offset, whence)
}

// Write writes len(b) bytes from b to the File. It returns the number of bytes
// written and an error, if any. Write returns a non-nil error when n != len(b).
func (d diskFile) Write(p []byte) (n int, err error) {
	if d.file == nil {
		return 0, fmt.Errorf("disk fs: write: file has not been opened")
	}
	return d.file.Write(p)
}

// WriteAt writes len(b) bytes to the File starting at byte offset off. It returns the
// number of bytes written and an error, if any. WriteAt returns a non-nil error when n != len(b).
func (d diskFile) WriteAt(p []byte, off int64) (n int, err error) {
	if d.file == nil {
		return 0, fmt.Errorf("disk fs: write at: file has not been opened")
	}
	return d.file.WriteAt(p, off)
}

// Read reads up to len(b) bytes from the File and stores them in b. It returns the number of
// bytes read and any error encountered. At end of file, Read returns 0, io.EOF.
func (d diskFile) Read(p []byte) (n int, err error) {
	if d.file == nil {
		return 0, fmt.Errorf("disk fs: read: file has not been opened")
	}
	return d.file.Read(p)
}

// ReadAt reads len(b) bytes from the File starting at byte offset off. It returns the number
// of bytes read and the error, if any. ReadAt always returns a non-nil error when n < len(b).
// At end of file, that error is io.EOF.
func (d diskFile) ReadAt(p []byte, off int64) (n int, err error) {
	if d.file == nil {
		return 0, fmt.Errorf("disk fs: read at: file has not been opened")
	}
	return d.file.ReadAt(p, off)
}

// Close releases all file handle resources. You will not be able to read/write any more
// data once this has been performed.
func (d diskFile) Close() error {
	if d.file == nil {
		return nil
	}
	return d.file.Close()
}

// Stat fetches metadata about the file w/o actually opening it for reading/writing.
func (d DiskFS) Stat(filePath string) (FileInfo, error) {
	file, err := os.Stat(path.Join(d.basePath, filePath))
	if err != nil {
		return nil, fmt.Errorf("disk fs error: stat: %w", err)
	}
	return file, nil
}

// Exists returns true when the file/directory already exits in the file system.
func (d DiskFS) Exists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// Read opens the given file at the given path, providing you with an io.Reader that
// you can use to stream bytes from it.
func (d DiskFS) Read(filePath string) (ReaderFile, error) {
	file, err := os.Open(path.Join(d.basePath, filePath))
	if err != nil {
		return nil, fmt.Errorf("disk fs error: open: %w", err)
	}

	// Make sure it's not a directory.
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("disk fs error: read: %w", err)
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("disk fs error: trying to read directory like a file: %s", filePath)
	}
	return diskFile{file: file}, nil
}

// Write opens the given file at the given path for writing. The resulting file
// behaves like a standard io.Writer/At.
//
// This operation will attempt to lazy-create the parent directory(s) if it does
// not exist. Should the file already exist, this will overwrite its entire contents
// so that it only contains what you write this time.
func (d DiskFS) Write(filePath string) (WriterFile, error) {
	fullPath := path.Join(d.basePath, filePath)

	// Ensure that the target directory actually exists.
	err := os.MkdirAll(path.Dir(fullPath), os.FileMode(0755))
	if err != nil {
		return nil, fmt.Errorf("disk fs error: mkdir: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("disk fs error: %w", err)
	}
	return diskFile{file: file}, nil
}

// List performs the equivalent of the "ls" command. It returns a slice of
// all files and directories found in the target dirPath.
//
// You can optionally provide a set of filters to limit which files/directories
// are included in the final set.
func (d DiskFS) List(dirPath string, filters ...FileFilter) ([]FileInfo, error) {
	entries, err := os.ReadDir(path.Join(d.basePath, dirPath))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("disk fs error: list files: %s %w", dirPath, err)
	}

	var results []FileInfo
	for _, entry := range entries {
		file, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("disk fs error: list files: %s %w", dirPath, err)
		}
		if !fileMatchesFilters(file, filters) {
			continue
		}
		results = append(results, file)
	}
	return results, nil
}

// WorkingDirectory returns the current FS context's path/directory.
func (d DiskFS) WorkingDirectory() string {
	return path.Clean(d.basePath)
}

// ChangeDirectory returns a new FS that is rooted in the given subdirectory of this FS.
func (d DiskFS) ChangeDirectory(dir string) FS {
	return Disk(path.Join(d.basePath, dir))
}

// Remove deletes the given file/directory and any of its children.
func (d DiskFS) Remove(fileOrDirPath string) error {
	if err := os.RemoveAll(path.Join(d.basePath, fileOrDirPath)); err != nil {
		return fmt.Errorf("disk fs error: remove %s: %w", fileOrDirPath, err)
	}
	return nil
}

// Move takes an existing file at the fromPath location and moves it to another
// spot in this file system; the toPath location.
func (d DiskFS) Move(fromPath string, toPath string) error {
	fromPath = path.Join(d.basePath, fromPath)
	toPath = path.Join(d.basePath, toPath)

	// Ensure the original file exists in the first place.
	if _, err := os.Stat(fromPath); err != nil {
		return fmt.Errorf("disk fs error: move: %v", err)
	}
	// Lazily create the directory where we will move the file to.
	if err := os.MkdirAll(path.Dir(toPath), os.FileMode(0755)); err != nil {
		return fmt.Errorf("disk fs error: move: %v", err)
	}
	// Move (the file), bitch. Get out the way!
	if err := os.Rename(fromPath, toPath); err != nil {
		return fmt.Errorf("disk fs error: move: %v", err)
	}
	return nil
}

func fileMatchesFilters(file FileInfo, filters []FileFilter) bool {
	for _, filter := range filters {
		if !filter(file) {
			return false
		}
	}
	return true
}

var _ FS = DiskFS{}
