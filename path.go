package filestore

import (
	"path"
	"strings"
)

// ChangeExtension helps datasets maintain the same file name stem while replacing
// the extension.
//
//    // Example
//    changeExtension("foo.jpg", "txt")  // "foo.txt"
//    changeExtension("foo.bar.png", "jpg")  // "foo.bar.jpg"
//    changeExtension("foo", "txt")  // "foo.txt"
func ChangeExtension(fileName string, ext string) string {
	// Go's path.Ext() returns extensions w/ the dot (e.g. ".jpg" or ".txt"), so
	// we'll add it to make the comparisons consistent. It's probably more natural
	// for the caller to just use the extension "jpg" or "txt", but this lets them
	// call this with either "jpg" or ".jpg" and it will work just fine.
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	currentExt := path.Ext(fileName)
	switch currentExt {
	case ext:
		return fileName
	default:
		return strings.TrimSuffix(fileName, currentExt) + ext
	}
}
