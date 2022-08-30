package filestore_test

import (
	"io/fs"
	"testing"
	"time"

	"github.com/monadicstack/filestore"
	"github.com/stretchr/testify/suite"
)

type FSTestSuite struct {
	suite.Suite
}

func (s *FSTestSuite) allowName(filter filestore.FileFilter, names ...string) {
	for _, name := range names {
		s.Require().True(filter(fakeFileInfo{name: name}), "Filter should allow file named '%s'", name)
	}
}

func (s *FSTestSuite) rejectName(filter filestore.FileFilter, names ...string) {
	for _, name := range names {
		s.Require().False(filter(fakeFileInfo{name: name}), "Filter should NOT allow file named '%s'", name)
	}
}

func (s *FSTestSuite) TestWithExt_everything() {
	s.allowName(filestore.WithExt(""),
		"",
		"foo",
		"foo.",
		"foo.bar",
		"foo.bar.",
		"foo.bar.jpeg",
		"foo.bar.jpég",
		"foo.bar.🍺🍺🍺🍺🍺🍺")

	s.allowName(filestore.WithExt("."),
		"",
		"foo",
		"foo.",
		"foo.bar",
		"foo.bar.",
		"foo.bar.jpeg",
		"foo.bar.jpég",
		"foo.bar.🍺🍺🍺🍺🍺🍺")
}

func (s *FSTestSuite) TestWithExt_specific() {
	s.allowName(filestore.WithExt("a"),
		".a",
		"foo.a",
		"foo.bar.a",
		"foo.bar.a.a.a",
		"foo.bar.🍺🍺🍺🍺🍺🍺.a",
	)
	s.rejectName(filestore.WithExt("a"),
		"",
		"a",
		"foo",
		"foo.",
		"foo.bar",
		"foo.bar.",
		"foo.bar.aa",
		"foo.bar.a.b",
	)

	s.allowName(filestore.WithExt("a.b"),
		".a.b",
		"foo.bar.a.b",
		"foo.bar.a.b.a.b",
		"foo.bar.🍺🍺🍺🍺🍺🍺.a.b",
	)
	s.rejectName(filestore.WithExt("a.b"),
		"",
		"a.b", // extension is only .b in this case
		"foo",
		"foo.",
		"foo.bar",
		"foo.bar.",
		"foo.bar.aa.b",
		"foo.bar.a.b.a",
		"foo.bar.a.b.🍺",
	)

	s.allowName(filestore.WithExt("png"),
		".png",
		"foo.bar.png",
		"foo.bar.png.png",
		"foo.bar.🍺🍺🍺🍺🍺🍺.png",
	)
	s.rejectName(filestore.WithExt("png"),
		"",
		"png",
		"foo",
		"foo.",
		"foo.bar",
		"foo.bar.",
		"foo.bar.ping",
		"foo.bar.pn",
		"foo.bar.pngpng",
		"foo.bar.png.jpg",
	)

	s.allowName(filestore.WithExt("🍺🍺🍺"),
		".🍺🍺🍺",
		"foo.bar.🍺🍺🍺",
		"foo.bar.🍺🍺🍺🍺🍺🍺.🍺🍺🍺",
	)
	s.rejectName(filestore.WithExt("🍺🍺🍺"),
		"",
		"🍺🍺🍺",
		"foo",
		"foo.",
		"foo.bar",
		"foo.bar.",
		"foo.bar.🍺🍺",
		"foo.bar.🍺🍺🍺🍺🍺🍺",
		"foo.bar.🍺🍺🍺.jpg",
	)
}

// Ensure that the filters still behave properly when you put a "." in front
// of the extension when building a WithExt() filter.
func (s *FSTestSuite) TestWithExt_withDot() {
	s.allowName(filestore.WithExt(".a"),
		".a",
		"foo.a",
		"foo.bar.a",
		"foo.bar.a.a.a",
		"foo.bar.🍺🍺🍺🍺🍺🍺.a",
	)
	s.rejectName(filestore.WithExt(".a"),
		"",
		"foo",
		"foo.",
		"foo.bar",
		"foo.bar.",
		"foo.bar.aa",
		"foo.bar.a.b",
	)
}

func (s *FSTestSuite) TestWithExts_everything() {
	s.allowName(filestore.WithExts("", "."),
		"",
		"foo",
		"foo.",
		"foo.bar",
		"foo.bar.",
		"foo.bar.jpeg",
		"foo.bar.jpég",
		"foo.bar.🍺🍺🍺🍺🍺🍺")
}

func (s *FSTestSuite) TestWithExts_one() {
	s.allowName(filestore.WithExts("png"),
		".png",
		"foo.bar.png",
		"foo.bar.png.png",
		"foo.bar.🍺🍺🍺🍺🍺🍺.png",
	)
	s.rejectName(filestore.WithExts("png"),
		"",
		"png",
		"foo",
		"foo.",
		"foo.bar",
		"foo.bar.",
		"foo.bar.ping",
		"foo.bar.pn",
		"foo.bar.pngpng",
		"foo.bar.png.jpg",
	)
}

func (s *FSTestSuite) TestWithExts_multiple() {
	s.allowName(filestore.WithExts("png", "jpg", "🍺"),
		".png",
		".jpg",
		"foo.bar.png",
		"foo.bar.jpg",
		"foo.bar.png.jpg",
		"foo.bar.png.png",
		"foo.bar.jpg.jpg",
		"foo.bar.🍺🍺🍺🍺🍺🍺.png",
		"foo.bar.🍺🍺🍺🍺🍺🍺.jpg",
		"time.for.🍺",
	)
	s.rejectName(filestore.WithExts("png", "jpg", "🍺"),
		"",
		"png",
		"jpg",
		"🍺",
		"foo",
		"foo.",
		"foo.bar",
		"foo.bar.",
		"foo.bar.ping",
		"foo.bar.pn",
		"foo.bar.pngpng",
		"foo.bar.png.jpg.tiff",
		"🍺.png.jpg.txt",
	)
}

func (s *FSTestSuite) TestWithPattern() {
	s.allowName(filestore.WithPattern(""),
		"",
		".",
		"foo",
		"foo.bar",
		"foo.bar.baz",
		"foo.🍺",
	)

	s.allowName(filestore.WithPattern("foo.txt"),
		"foo.txt",
	)
	s.rejectName(filestore.WithPattern("foo.txt"),
		"",
		".",
		"foo.txt.more",
		"foo txt",
		"foo,txt",
		"foo🍺txt",
	)

	s.allowName(filestore.WithPattern("foo?txt"),
		"foo.txt",
		"foo txt",
		"foo,txt",
		"foo🍺txt",
	)
	s.rejectName(filestore.WithPattern("foo?txt"),
		"",
		".",
		"foo.txt.foo.txt",
		"foo.txt*",
		"footxt",
	)

	s.allowName(filestore.WithPattern("foo/*.txt"),
		"foo/a.txt",
		"foo/hello.txt",
		"foo/.txt",
		"foo/🍺🍺🍺🍺.txt",
		"foo/bar🍺🍺🍺🍺.txt",
	)
	s.rejectName(filestore.WithPattern("foo/*.txt"),
		"",
		".",
		"foo.txt",
		"foo.bar.txt",
		"foo/bar.text",
		"foo/bar/baz.txt", // Go's glob doesn't support "/" as a matching char in "*"
		"foo\\bar.txt",
	)
}

func (s *FSTestSuite) TestWithEverything() {
	s.allowName(filestore.WithEverything(),
		"",
		".",
		"......",
		"foo/bar.a.txt",
		"foo/.txt",
		"foo/🍺🍺🍺🍺.txt",
		"foo/bar🍺.🍺🍺/🍺.txt",
	)
}

func TestFSTestSuite(t *testing.T) {
	suite.Run(t, &FSTestSuite{})
}

type fakeFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	dir     bool
	sys     any
}

func (f fakeFileInfo) Name() string {
	return f.name
}

func (f fakeFileInfo) Size() int64 {
	return f.size
}

func (f fakeFileInfo) Mode() fs.FileMode {
	return f.mode
}

func (f fakeFileInfo) ModTime() time.Time {
	return f.modTime
}

func (f fakeFileInfo) IsDir() bool {
	return f.dir
}

func (f fakeFileInfo) Sys() any {
	return f.sys
}
