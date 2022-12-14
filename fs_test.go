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
		"foo.bar.jpÃ©g",
		"foo.bar.ðºðºðºðºðºðº")

	s.allowName(filestore.WithExt("."),
		"",
		"foo",
		"foo.",
		"foo.bar",
		"foo.bar.",
		"foo.bar.jpeg",
		"foo.bar.jpÃ©g",
		"foo.bar.ðºðºðºðºðºðº")
}

func (s *FSTestSuite) TestWithExt_specific() {
	s.allowName(filestore.WithExt("a"),
		".a",
		"foo.a",
		"foo.bar.a",
		"foo.bar.a.a.a",
		"foo.bar.ðºðºðºðºðºðº.a",
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
		"foo.bar.ðºðºðºðºðºðº.a.b",
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
		"foo.bar.a.b.ðº",
	)

	s.allowName(filestore.WithExt("png"),
		".png",
		"foo.bar.png",
		"foo.bar.png.png",
		"foo.bar.ðºðºðºðºðºðº.png",
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

	s.allowName(filestore.WithExt("ðºðºðº"),
		".ðºðºðº",
		"foo.bar.ðºðºðº",
		"foo.bar.ðºðºðºðºðºðº.ðºðºðº",
	)
	s.rejectName(filestore.WithExt("ðºðºðº"),
		"",
		"ðºðºðº",
		"foo",
		"foo.",
		"foo.bar",
		"foo.bar.",
		"foo.bar.ðºðº",
		"foo.bar.ðºðºðºðºðºðº",
		"foo.bar.ðºðºðº.jpg",
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
		"foo.bar.ðºðºðºðºðºðº.a",
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
		"foo.bar.jpÃ©g",
		"foo.bar.ðºðºðºðºðºðº")
}

func (s *FSTestSuite) TestWithExts_one() {
	s.allowName(filestore.WithExts("png"),
		".png",
		"foo.bar.png",
		"foo.bar.png.png",
		"foo.bar.ðºðºðºðºðºðº.png",
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
	s.allowName(filestore.WithExts("png", "jpg", "ðº"),
		".png",
		".jpg",
		"foo.bar.png",
		"foo.bar.jpg",
		"foo.bar.png.jpg",
		"foo.bar.png.png",
		"foo.bar.jpg.jpg",
		"foo.bar.ðºðºðºðºðºðº.png",
		"foo.bar.ðºðºðºðºðºðº.jpg",
		"time.for.ðº",
	)
	s.rejectName(filestore.WithExts("png", "jpg", "ðº"),
		"",
		"png",
		"jpg",
		"ðº",
		"foo",
		"foo.",
		"foo.bar",
		"foo.bar.",
		"foo.bar.ping",
		"foo.bar.pn",
		"foo.bar.pngpng",
		"foo.bar.png.jpg.tiff",
		"ðº.png.jpg.txt",
	)
}

func (s *FSTestSuite) TestWithPattern() {
	s.allowName(filestore.WithPattern(""),
		"",
		".",
		"foo",
		"foo.bar",
		"foo.bar.baz",
		"foo.ðº",
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
		"fooðºtxt",
	)

	s.allowName(filestore.WithPattern("foo?txt"),
		"foo.txt",
		"foo txt",
		"foo,txt",
		"fooðºtxt",
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
		"foo/ðºðºðºðº.txt",
		"foo/barðºðºðºðº.txt",
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
		"foo/ðºðºðºðº.txt",
		"foo/barðº.ðºðº/ðº.txt",
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
